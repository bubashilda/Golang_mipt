//go:build !solution

package lrucache

import (
	"container/list"
	"fmt"
)

type LRUCache struct {
	Cache list.List
	Cap   int
}

type Pair struct {
	Key   int
	Value int
}

func (cont *LRUCache) Get(key int) (int, bool) {
	for elem := cont.Cache.Front(); elem != nil; elem = elem.Next() {
		if elem.Value.(*Pair).Key == key {
			cont.Cache.MoveToFront(elem)
			return elem.Value.(*Pair).Value, true
		}
	}
	return 0, false
}

func (cont *LRUCache) Set(key, value int) {
	for elem := cont.Cache.Front(); elem != nil; elem = elem.Next() {
		if elem.Value.(*Pair).Key == key {
			elem.Value.(*Pair).Value = value
			cont.Cache.MoveToFront(elem)
			return
		}
	}

	if cont.Cap == 0 {
		return
	}

	if cont.Cap == cont.Cache.Len() {
		cont.Cache.Back().Value.(*Pair).Key = key
		cont.Cache.Back().Value.(*Pair).Value = value
		cont.Cache.MoveToFront(cont.Cache.Back())
		return
	}

	cont.Cache.PushFront(&Pair{key, value})
}

func (cont *LRUCache) Range(f func(key int, value int) bool) {
	for elem := cont.Cache.Back(); elem != nil; elem = elem.Prev() {
		if res := f(elem.Value.(*Pair).Key, elem.Value.(*Pair).Value); !res {
			return
		}
	}
}

func (cont *LRUCache) Clear() {
	for front := cont.Cache.Front(); front != nil; {
		next := front.Next()
		cont.Cache.Remove(front)
		front = next
	}
}

func New(Cap int) Cache {
	fmt.Printf("%T\n", list.List{})
	return &LRUCache{Cache: list.List{}, Cap: Cap}
}
