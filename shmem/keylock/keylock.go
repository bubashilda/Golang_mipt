//go:build !solution

package keylock

import (
	"sort"
	"sync"
)

type entry struct {
	sync.Once
	mutex chan struct{}
}

type KeyLock struct {
	mp sync.Map
}

func New() *KeyLock {
	return &KeyLock{}
}

func (kl *KeyLock) get(key string) *chan struct{} {
	newEntry := &entry{}
	old, loaded := kl.mp.LoadOrStore(key, newEntry)
	if loaded {
		newEntry = old.(*entry)
	}

	newEntry.Do(func() {
		newEntry.mutex = make(chan struct{}, 1)
		newEntry.mutex <- struct{}{}
	})

	return &newEntry.mutex
}

func (kl *KeyLock) LockKeys(keys []string, cancel <-chan struct{}) (canceled bool, unlock func()) {
	copiedKeys := make([]string, len(keys))
	copy(copiedKeys, keys)
	sort.Strings(copiedKeys)
	for idx, key := range copiedKeys {
		mutex := kl.get(key)
		select {
		case <-cancel:
			for i := idx - 1; i >= 0; i-- {
				retract := kl.get(copiedKeys[i])
				*retract <- struct{}{}
			}
			return true, func() {}
		case <-*mutex:

		}
	}
	return false, func() {
		for idx := range copiedKeys {
			mutex := kl.get(copiedKeys[len(copiedKeys)-idx-1])
			*mutex <- struct{}{}
		}
	}
}
