package main

import (
	"container/list"
	"sync"
)

type LRU[K comparable, V any] interface {
	Get(key K) (value V, ok bool)
	Set(key K, value V)
}

type element[K comparable, V any] struct {
	key   K
	value V
}

type LRUCache[K comparable, V any] struct {
	store    map[K]*list.Element
	list     *list.List
	capacity int
	mu       sync.Mutex
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{store: make(map[K]*list.Element, capacity), list: list.New(), capacity: capacity}
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.capacity == 0 {
		return
	}

	if e, ok := c.store[key]; ok {
		c.list.MoveToFront(e)
		e.Value.(*element[K, V]).value = value
		return
	}

	newElement := &element[K, V]{key: key, value: value}
	e := c.list.PushFront(newElement)
	c.store[key] = e

	if c.list.Len() > c.capacity {
		LRUElement := c.list.Back()
		if LRUElement != nil {
			c.list.Remove(LRUElement)
			LRUKey := LRUElement.Value.(*element[K, V]).key
			delete(c.store, LRUKey)
		}
	}
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.store[key]

	if !ok || c.capacity == 0 {
		var zeroValue V
		return zeroValue, false
	}

	c.list.MoveToFront(e)
	return e.Value.(*element[K, V]).value, true
}
