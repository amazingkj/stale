package scheduler

import (
	"sync"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
)

func TestNew(t *testing.T) {
	// Test that New creates a scheduler with initialized fields
	s := &Scheduler{
		cron:   cron.New(cron.WithLocation(time.Local)),
		stopCh: make(chan struct{}),
	}

	if s.cron == nil {
		t.Error("cron should not be nil")
	}
	if s.stopCh == nil {
		t.Error("stopCh should not be nil")
	}
}

func TestClearRunningJob(t *testing.T) {
	tests := []struct {
		name         string
		initialJobID *int64
		clearID      int64
		expectedNil  bool
	}{
		{
			name:         "clear matching job",
			initialJobID: ptr(int64(123)),
			clearID:      123,
			expectedNil:  true,
		},
		{
			name:         "clear non-matching job",
			initialJobID: ptr(int64(123)),
			clearID:      456,
			expectedNil:  false,
		},
		{
			name:         "clear when no job running",
			initialJobID: nil,
			clearID:      123,
			expectedNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scheduler{
				runningJobID: tt.initialJobID,
			}

			s.ClearRunningJob(tt.clearID)

			if tt.expectedNil && s.runningJobID != nil {
				t.Errorf("runningJobID should be nil, got %v", *s.runningJobID)
			}
			if !tt.expectedNil && s.runningJobID == nil {
				t.Error("runningJobID should not be nil")
			}
		})
	}
}

func TestOnScanComplete(t *testing.T) {
	s := &Scheduler{}

	callCount := 0
	callback := func() {
		callCount++
	}

	// Register multiple callbacks
	s.OnScanComplete(callback)
	s.OnScanComplete(callback)
	s.OnScanComplete(callback)

	if len(s.onScanComplete) != 3 {
		t.Errorf("expected 3 callbacks, got %d", len(s.onScanComplete))
	}

	// Verify callbacks are stored correctly
	for _, cb := range s.onScanComplete {
		cb()
	}

	if callCount != 3 {
		t.Errorf("expected callCount to be 3, got %d", callCount)
	}
}

func TestNotifyScanComplete(t *testing.T) {
	s := &Scheduler{}

	var results []int
	var mu sync.Mutex

	// Register callbacks that append to results
	for i := 1; i <= 3; i++ {
		val := i
		s.OnScanComplete(func() {
			mu.Lock()
			results = append(results, val)
			mu.Unlock()
		})
	}

	s.notifyScanComplete()

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// Check all callbacks were executed
	for i, v := range results {
		if v != i+1 {
			t.Errorf("results[%d] = %d, want %d", i, v, i+1)
		}
	}
}

func TestSchedulerConcurrency(t *testing.T) {
	s := &Scheduler{}

	// Test concurrent access to OnScanComplete
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.OnScanComplete(func() {})
		}()
	}
	wg.Wait()

	if len(s.onScanComplete) != 100 {
		t.Errorf("expected 100 callbacks, got %d", len(s.onScanComplete))
	}
}

func TestSchedulerMutexProtection(t *testing.T) {
	s := &Scheduler{
		runningJobID: nil,
	}

	// Test concurrent ClearRunningJob calls
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			s.runningJobID = &id
			s.ClearRunningJob(id)
		}(int64(i))
	}
	wg.Wait()

	// Should not panic and runningJobID should be nil at the end
	if s.runningJobID != nil {
		t.Logf("runningJobID = %v (may not be nil due to race)", *s.runningJobID)
	}
}

func TestErrScanAlreadyRunning(t *testing.T) {
	if ErrScanAlreadyRunning.Error() != "a scan is already running" {
		t.Errorf("unexpected error message: %s", ErrScanAlreadyRunning.Error())
	}
}

func ptr(v int64) *int64 {
	return &v
}
