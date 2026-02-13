package main

import (
	"container/list"
	"sync"
)

type LRU[K comparable, V any] interface {
	Get(key K) (value V, ok bool)
	Set(key K, value V)
}

type entry[K comparable, V any] struct {
	key   K
	value V
}

type LRUCache[K comparable, V any] struct {
	mu       sync.Mutex
	data     map[K]*list.Element
	list     *list.List
	capacity int
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	var m map[K]*list.Element

	if capacity > 0 {
		m = make(map[K]*list.Element, capacity)
	} else {
		capacity = 0
	}

	c := &LRUCache[K, V]{
		data:     m,
		list:     list.New(),
		capacity: capacity,
	}
	return c
}

func (cache *LRUCache[K, V]) Set(key K, value V) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if cache.capacity == 0 {
		return
	}

	elem, ok := cache.data[key]

	if ok {
		elem.Value.(*entry[K, V]).value = value
		cache.list.MoveToFront(elem)
		return
	}

	ent := &entry[K, V]{key: key, value: value}
	elem = cache.list.PushFront(ent)
	cache.data[key] = elem

	if cache.list.Len() > cache.capacity {
		elem = cache.list.Back()
		if elem != nil {
			ent := elem.Value.(*entry[K, V])
			delete(cache.data, ent.key)
			cache.list.Remove(elem)
		}
	}
}

func (cache *LRUCache[K, V]) Get(key K) (V, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if cache.capacity == 0 {
		var zero V
		return zero, false
	}

	elem, ok := cache.data[key]

	if !ok {
		var zero V
		return zero, false
	}

	cache.list.MoveToFront(elem)

	ent := elem.Value.(*entry[K, V])
	return ent.value, true
}
