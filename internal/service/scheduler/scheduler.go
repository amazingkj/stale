package scheduler

import (
	"context"
	"time"

	"github.com/jiin/stale/internal/domain"
	"github.com/jiin/stale/internal/repository"
	"github.com/jiin/stale/internal/service/scanner"
	"github.com/rs/zerolog/log"
)

type Scheduler struct {
	scanner      *scanner.Scanner
	scanRepo     *repository.ScanRepository
	intervalHrs  int
	stopCh       chan struct{}
	runningJobID *int64
}

func New(
	scanner *scanner.Scanner,
	scanRepo *repository.ScanRepository,
	intervalHrs int,
) *Scheduler {
	return &Scheduler{
		scanner:     scanner,
		scanRepo:    scanRepo,
		intervalHrs: intervalHrs,
		stopCh:      make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	if s.intervalHrs <= 0 {
		log.Info().Msg("scheduler disabled (interval <= 0)")
		return
	}

	ticker := time.NewTicker(time.Duration(s.intervalHrs) * time.Hour)
	defer ticker.Stop()

	log.Info().Int("interval_hours", s.intervalHrs).Msg("scheduler started")

	for {
		select {
		case <-s.stopCh:
			log.Info().Msg("scheduler stopped")
			return
		case <-ticker.C:
			s.runScheduledScan()
		}
	}
}

func (s *Scheduler) Stop() {
	close(s.stopCh)
}

func (s *Scheduler) runScheduledScan() {
	ctx := context.Background()

	scan, err := s.scanRepo.Create(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to create scheduled scan job")
		return
	}

	log.Info().Int64("scan_id", scan.ID).Msg("starting scheduled scan")

	if err := s.scanRepo.UpdateStatus(ctx, scan.ID, domain.ScanStatusRunning, nil); err != nil {
		log.Error().Err(err).Msg("failed to update scan status to running")
		return
	}

	scanErr := s.scanner.ScanAll(ctx, scan.ID)

	status := domain.ScanStatusCompleted
	if scanErr != nil {
		status = domain.ScanStatusFailed
		log.Error().Err(scanErr).Int64("scan_id", scan.ID).Msg("scheduled scan failed")
	} else {
		log.Info().Int64("scan_id", scan.ID).Msg("scheduled scan completed")
	}

	if err := s.scanRepo.UpdateStatus(ctx, scan.ID, status, scanErr); err != nil {
		log.Error().Err(err).Msg("failed to update scan status")
	}
}

func (s *Scheduler) TriggerScan(ctx context.Context, sourceID *int64) (*domain.ScanJob, error) {
	scan, err := s.scanRepo.Create(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	go s.runScan(scan.ID, sourceID)

	return scan, nil
}

func (s *Scheduler) runScan(scanID int64, sourceID *int64) {
	ctx := context.Background()

	if err := s.scanRepo.UpdateStatus(ctx, scanID, domain.ScanStatusRunning, nil); err != nil {
		log.Error().Err(err).Msg("failed to update scan status to running")
		return
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
	}

	if err := s.scanRepo.UpdateStatus(ctx, scanID, status, scanErr); err != nil {
		log.Error().Err(err).Msg("failed to update scan status")
	}
}
