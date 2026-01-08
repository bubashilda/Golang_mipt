//go:build !solution

package dupcall

import (
	"context"
	"sync"
)

type CallControlBlock struct {
	ctx     context.Context
	cancel  context.CancelFunc
	waiters int

	completedMu sync.Mutex
	val         interface{}
	err         error
}

type Call struct {
	mu  sync.Mutex
	ccb *CallControlBlock
}

func (o *Call) Do(
	ctx context.Context,
	cb func(context.Context) (interface{}, error),
) (result interface{}, err error) {
	o.mu.Lock()

	acquire := func(ccb *CallControlBlock) (interface{}, error) {
		select {
		case <-ccb.ctx.Done():
			ccb.completedMu.Lock()
			defer ccb.completedMu.Unlock()
			return ccb.val, ccb.err

		case <-ctx.Done():
			ccb.completedMu.Lock()
			defer ccb.completedMu.Unlock()
			ccb.waiters--
			if ccb.waiters == 0 {
				ccb.cancel()
			}
			return nil, ctx.Err()
		}
	}

	if o.ccb != nil {
		ccb := o.ccb
		o.mu.Unlock()

		ccb.completedMu.Lock()
		ccb.waiters++
		ccb.completedMu.Unlock()

		return acquire(ccb)
	}

	ccb := &CallControlBlock{
		waiters: 1,
	}
	ccb.ctx, ccb.cancel = context.WithCancel(context.Background())
	o.ccb = ccb
	o.mu.Unlock()

	go func() {
		defer ccb.cancel()
		v, e := cb(ccb.ctx)

		ccb.completedMu.Lock()
		ccb.val, ccb.err = v, e
		ccb.completedMu.Unlock()

		o.mu.Lock()
		o.ccb = nil
		o.mu.Unlock()
	}()

	return acquire(ccb)
}
