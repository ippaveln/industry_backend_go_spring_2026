package main

import (
	"container/list"
	"sync"
)

type LRU[K comparable, V any] interface {
	Get(key K) (value V, ok bool)
	Set(key K, value V)
}

type LRUCache[K comparable, V any] struct {
	capacity int
	ll       *list.List                    
	cache    map[K]*list.Element         
	mu       sync.Mutex
}

type entry[K comparable, V any] struct {
	key   K
	value V
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	if capacity < 0 {
		capacity = 0
	}
	return &LRUCache[K, V]{
		capacity: capacity,
		ll:       list.New(),
		cache:    make(map[K]*list.Element, capacity),
	}
}

func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.capacity == 0 {
		var zero V
		return zero, false
	}

	if elem, exists := c.cache[key]; exists {
		c.ll.MoveToFront(elem)
		return elem.Value.(*entry[K, V]).value, true
	}
	var zero V
	return zero, false
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.capacity == 0 {
		return
	}

	if elem, exists := c.cache[key]; exists {
		c.ll.MoveToFront(elem)
		elem.Value.(*entry[K, V]).value = value
		return
	}

	e := &entry[K, V]{key: key, value: value}
	elem := c.ll.PushFront(e)
	c.cache[key] = elem

	if c.ll.Len() > c.capacity {
		backElem := c.ll.Back()
		if backElem != nil {
			c.ll.Remove(backElem)
			delete(c.cache, backElem.Value.(*entry[K, V]).key)
		}
	}
}
