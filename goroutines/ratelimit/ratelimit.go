//go:build !solution

package ratelimit

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Limiter is precise rate limiter with context support.
type Limiter struct {
	interval    time.Duration
	sema        chan struct{}
	stopped     chan struct{}
	flagStopped bool
	mu          sync.Mutex
}

var ErrStopped = errors.New("limiter stopped")

// NewLimiter returns limiter that throttles rate of successful Acquire() calls
// to maxSize events at any given interval.
func NewLimiter(maxCount int, interval time.Duration) *Limiter {
	limiter := Limiter{
		interval: interval,
		sema:     make(chan struct{}, maxCount),
		stopped:  make(chan struct{}),
	}
	for _ = range maxCount {
		limiter.sema <- struct{}{}
	}
	return &limiter
}

func (l *Limiter) Acquire(ctx context.Context) error {
	select {
	case <-l.stopped:
		return ErrStopped
	default:
	}

	select {
	case _, ok := <-l.sema:
		if !ok {
			return ErrStopped
		}

		go func() {
			writeBack := func() {
				l.mu.Lock()
				defer l.mu.Unlock()
				if l.flagStopped {
					return
				}
				l.sema <- struct{}{}
			}

			if l.interval == 0 {
				select {
				case <-l.stopped:
				default:
					writeBack()
				}
				return
			}

			ticker := time.NewTicker(l.interval)
			defer ticker.Stop()

			select {
			case <-l.stopped:
			case <-ticker.C:
				writeBack()
			}
		}()

		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (l *Limiter) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flagStopped = true
	close(l.stopped)
	close(l.sema)
}
