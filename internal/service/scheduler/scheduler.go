package scheduler

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jiin/stale/internal/domain"
	"github.com/jiin/stale/internal/repository"
	"github.com/jiin/stale/internal/service/email"
	"github.com/jiin/stale/internal/service/scanner"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

var ErrScanAlreadyRunning = errors.New("a scan is already running")

type Scheduler struct {
	scanner          *scanner.Scanner
	scanRepo         *repository.ScanRepository
	depRepo          *repository.DependencyRepository
	settingsRepo     *repository.SettingsRepository
	emailService     *email.Service
	cron             *cron.Cron
	cronEntryID      cron.EntryID
	stopCh           chan struct{}
	mu               sync.Mutex
	runningJobID     *int64
	onScanComplete   []func() // Callbacks to run after scan completes
}

func New(
	scanner *scanner.Scanner,
	scanRepo *repository.ScanRepository,
	depRepo *repository.DependencyRepository,
	settingsRepo *repository.SettingsRepository,
	emailService *email.Service,
) *Scheduler {
	return &Scheduler{
		scanner:      scanner,
		scanRepo:     scanRepo,
		depRepo:      depRepo,
		settingsRepo: settingsRepo,
		emailService: emailService,
		cron:         cron.New(cron.WithLocation(time.Local)),
		stopCh:       make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	// Cleanup any stale scans from previous runs
	ctx := context.Background()
	if affected, err := s.scanRepo.CleanupStaleScans(ctx); err != nil {
		log.Warn().Err(err).Msg("failed to cleanup stale scans on startup")
	} else if affected > 0 {
		log.Info().Int64("cleaned_up", affected).Msg("cleaned up stale scans from previous runs")
	}

	// Load settings and configure cron
	s.ReloadSchedule()

	// Start cron scheduler
	s.cron.Start()
	log.Info().Str("timezone", time.Local.String()).Msg("cron scheduler started")

	<-s.stopCh
	log.Info().Msg("scheduler stopped")
}

func (s *Scheduler) ReloadSchedule() {
	ctx := context.Background()
	settings, err := s.settingsRepo.Get(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to load settings for scheduler")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove existing cron job if any
	if s.cronEntryID != 0 {
		s.cron.Remove(s.cronEntryID)
		s.cronEntryID = 0
	}

	if !settings.ScheduleEnabled {
		log.Info().Msg("scheduled scans disabled")
		return
	}

	// Add new cron job
	entryID, err := s.cron.AddFunc(settings.ScheduleCron, s.runScheduledScan)
	if err != nil {
		log.Error().Err(err).Str("cron", settings.ScheduleCron).Msg("invalid cron expression")
		return
	}

	s.cronEntryID = entryID
	log.Info().Str("cron", settings.ScheduleCron).Msg("scheduled scan configured")
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
	close(s.stopCh)
}

func (s *Scheduler) ClearRunningJob(scanID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.runningJobID != nil && *s.runningJobID == scanID {
		s.runningJobID = nil
	}
}

// OnScanComplete registers a callback to run after scan completes
func (s *Scheduler) OnScanComplete(callback func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onScanComplete = append(s.onScanComplete, callback)
}

func (s *Scheduler) notifyScanComplete() {
	s.mu.Lock()
	callbacks := s.onScanComplete
	s.mu.Unlock()
	for _, cb := range callbacks {
		cb()
	}
}

func (s *Scheduler) runScheduledScan() {
	s.mu.Lock()
	if s.runningJobID != nil {
		s.mu.Unlock()
		log.Info().Msg("skipping scheduled scan - another scan is already running")
		return
	}

	ctx := context.Background()

	scan, err := s.scanRepo.Create(ctx, nil)
	if err != nil {
		s.mu.Unlock()
		log.Error().Err(err).Msg("failed to create scheduled scan job")
		return
	}

	s.runningJobID = &scan.ID
	s.mu.Unlock()

	// Clear running job ID when done
	defer func() {
		s.mu.Lock()
		s.runningJobID = nil
		s.mu.Unlock()
	}()

	log.Info().Int64("scan_id", scan.ID).Msg("starting scheduled scan")

	if err := s.scanRepo.UpdateStatus(ctx, scan.ID, domain.ScanStatusRunning, nil); err != nil {
		log.Error().Err(err).Msg("failed to update scan status to running")
		return
	}

	// Mark current outdated status before scan
	if err := s.depRepo.MarkPreviouslyOutdated(ctx); err != nil {
		log.Warn().Err(err).Msg("failed to mark previously outdated dependencies")
	}

	scanErr := s.scanner.ScanAll(ctx, scan.ID)

	status := domain.ScanStatusCompleted
	if scanErr != nil {
		status = domain.ScanStatusFailed
		log.Error().Err(scanErr).Int64("scan_id", scan.ID).Msg("scheduled scan failed")
	} else {
		log.Info().Int64("scan_id", scan.ID).Msg("scheduled scan completed")
		// Send email notification for new outdated dependencies
		s.sendNewOutdatedNotification(ctx, scan.ID)
	}

	if err := s.scanRepo.UpdateStatus(ctx, scan.ID, status, scanErr); err != nil {
		log.Error().Err(err).Msg("failed to update scan status")
	}

	// Notify scan complete callbacks (cache invalidation, etc.)
	s.notifyScanComplete()
}

func (s *Scheduler) sendNewOutdatedNotification(ctx context.Context, scanID int64) {
	settings, err := s.settingsRepo.Get(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to load settings for email notification")
		return
	}

	if !settings.EmailEnabled || !settings.EmailNotifyNewOutdated {
		return
	}

	// Get newly outdated dependencies
	newOutdated, err := s.depRepo.GetNewlyOutdated(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get newly outdated dependencies")
		return
	}

	if len(newOutdated) == 0 {
		log.Debug().Msg("no new outdated dependencies to report")
		return
	}

	report := &domain.NewOutdatedReport{
		ScanID:      scanID,
		NewOutdated: newOutdated,
	}

	if err := s.emailService.SendNewOutdatedReport(settings, report); err != nil {
		log.Error().Err(err).Msg("failed to send email notification")
	}
}

func (s *Scheduler) TriggerScan(ctx context.Context, sourceID *int64) (*domain.ScanJob, error) {
	s.mu.Lock()
	if s.runningJobID != nil {
		s.mu.Unlock()
		return nil, ErrScanAlreadyRunning
	}

	scan, err := s.scanRepo.Create(ctx, sourceID)
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}

	s.runningJobID = &scan.ID
	s.mu.Unlock()

	go s.runScan(scan.ID, sourceID)

	return scan, nil
}

func (s *Scheduler) runScan(scanID int64, sourceID *int64) {
	ctx := context.Background()

	// Clear running job ID when done
	defer func() {
		s.mu.Lock()
		s.runningJobID = nil
		s.mu.Unlock()
	}()

	if err := s.scanRepo.UpdateStatus(ctx, scanID, domain.ScanStatusRunning, nil); err != nil {
		log.Error().Err(err).Msg("failed to update scan status to running")
		return
	}

	// Mark current outdated status before scan
	if err := s.depRepo.MarkPreviouslyOutdated(ctx); err != nil {
		log.Warn().Err(err).Msg("failed to mark previously outdated dependencies")
	}

	var scanErr error
	if sourceID != nil {
		scanErr = s.scanner.ScanSource(ctx, *sourceID, scanID)
	} else {
		scanErr = s.scanner.ScanAll(ctx, scanID)
	}

	status := domain.ScanStatusCompleted
	if scanErr != nil {
		status = domain.ScanStatusFailed
		log.Error().Err(scanErr).Int64("scan_id", scanID).Msg("scan failed")
	} else {
		log.Info().Int64("scan_id", scanID).Msg("scan completed")
		// Send email notification for new outdated dependencies
		s.sendNewOutdatedNotification(ctx, scanID)
	}

	if err := s.scanRepo.UpdateStatus(ctx, scanID, status, scanErr); err != nil {
		log.Error().Err(err).Msg("failed to update scan status")
	}

	// Notify scan complete callbacks (cache invalidation, etc.)
	s.notifyScanComplete()
}
