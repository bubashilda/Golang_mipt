//go:build !solution

package batcher

import (
	"gitlab.com/slon/shad-go/batcher/slow"
	"sync"
)

type Status int

const (
	Ready Status = 0
	Load  Status = 1
)

type Batcher struct {
	v       *slow.Value
	mutex   sync.Mutex
	series  chan struct{}
	value   *interface{}
	state   Status
	waiters []chan struct{}
}

func NewBatcher(v *slow.Value) *Batcher {
	return &Batcher{
		v:       v,
		series:  make(chan struct{}),
		value:   new(interface{}),
		waiters: make([]chan struct{}, 0),
		state:   Ready,
	}
}

func (b *Batcher) Load() interface{} {
	b.mutex.Lock()

	state := b.state
	value := b.value

	b.v.Updated.Lock()
	updated := b.v.FlagUpdated
	if updated {
		b.v.FlagUpdated = false
	}
	b.v.Updated.Unlock()

	if !updated && state == Ready {
		defer b.mutex.Unlock()
		return *value
	}

	if !updated && state == Load {
		waitCh := b.series
		b.mutex.Unlock()
		<-waitCh
		return *value
	}

	var newValue interface{}
	defer func() {
		if len(b.waiters) == 0 {
			b.state = Ready
		} else {
			close(b.waiters[0])
			b.waiters = b.waiters[1:]
		}
		b.mutex.Unlock()
	}()

	switch state {
	case Ready:
		b.state = Load
		b.series = make(chan struct{})
		waitCh := b.series
		b.mutex.Unlock()

		// Выполняем загрузку
		newValue = b.v.Load()

		b.mutex.Lock()
		*value = newValue
		close(waitCh)
	case Load:
		b.series = make(chan struct{})
		waitCh := b.series
		self := make(chan struct{})
		b.waiters = append(b.waiters, self)
		b.mutex.Unlock()

		<-self
		newValue = b.v.Load()

		b.mutex.Lock()
		*value = newValue
		close(waitCh)
	}

	return newValue
}
