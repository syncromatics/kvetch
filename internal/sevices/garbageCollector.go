package services

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// GarbageCollecter runs garbage collection on the datastore log files
type GarbageCollecter interface {
	GarbageCollect() error
}

// GarbageCollectorService runs garbage collection on an interval
type GarbageCollectorService struct {
	collector GarbageCollecter
	interval  time.Duration
}

// NewGarbageCollectorService creates a new garbage collector service
func NewGarbageCollectorService(collector GarbageCollecter, interval time.Duration) *GarbageCollectorService {
	return &GarbageCollectorService{
		collector: collector,
		interval:  interval,
	}
}

// Run runs the garbage collector service
func (s GarbageCollectorService) Run(ctx context.Context) func() error {
	return func() error {
		timer := time.NewTimer(0 * time.Second)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil

			case <-timer.C:
				err := s.collector.GarbageCollect()
				if err != nil {
					return errors.Wrap(err, "failed GarbageCollect")
				}
				timer = time.NewTimer(s.interval)
			}
		}
	}
}
