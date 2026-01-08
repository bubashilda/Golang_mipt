//go:build !solution

package pubsub

import (
	"container/list"
	"context"
	"errors"
	"sync"
)

var _ Subscription = (*MySubscription)(nil)

type MySubscription struct {
	*list.Element

	subName  string
	callback MsgHandler

	channel chan interface{}
	pubsub  *MyPubSub
}

func (s *MySubscription) Unsubscribe() {
	s.pubsub.Lock()
	defer s.pubsub.Unlock()
	s.pubsub.channels[s.subName].Remove(s.Element)
	close(s.channel)
}

var _ PubSub = (*MyPubSub)(nil)

type MyPubSub struct {
	channels map[string]*list.List
	closed   bool

	sync.WaitGroup
	sync.RWMutex
}

func NewPubSub() PubSub {
	return &MyPubSub{channels: make(map[string]*list.List), closed: false}
}

func (p *MyPubSub) Subscribe(subj string, cb MsgHandler) (Subscription, error) {
	p.Lock()
	defer p.Unlock()

	if p.closed {
		return nil, errors.New("pubsub closed")
	}

	old := p.channels[subj]
	if old == nil {
		old = list.New()
		p.channels[subj] = old
	}

	node := &MySubscription{
		subName:  subj,
		callback: cb,
		channel:  make(chan interface{}, 128),
		pubsub:   p,
	}
	node.Element = old.PushBack(node)

	go func() {
		p.Add(1)
		defer p.Done()
		for {
			msg := <-node.channel
			if msg == nil {
				return
			}
			cb(msg)
		}
	}()
	return node, nil
}

func (p *MyPubSub) Publish(subj string, msg interface{}) error {
	p.RLock()
	defer p.RUnlock()

	if p.closed {
		return errors.New("pubsub closed")
	}

	lst := p.channels[subj]
	for first := lst.Front(); first != nil; first = first.Next() {
		first.Value.(*MySubscription).channel <- msg
	}
	return nil
}

func (p *MyPubSub) Close(ctx context.Context) error {
	p.Lock()
	p.closed = true
	p.Unlock()

	for _, lst := range p.channels {
		for e := lst.Front(); e != nil; e = e.Next() {
			close(e.Value.(*MySubscription).channel)
		}
	}

	done := make(chan struct{})
	go func() {
		p.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
