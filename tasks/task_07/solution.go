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
	cache    map[K]*list.Element
	capacity int
	list     *list.List
	mu       sync.Mutex
}

type Entry[K comparable, V any] struct {
	key   K
	value V
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		cache:    make(map[K]*list.Element),
		capacity: capacity,
		list:     list.New(),
	}
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var zero V

	v, ok := c.cache[key]

	if !ok {
		return zero, false
	}

	entry := v.Value.(Entry[K, V])
	c.list.MoveToFront(v)

	return entry.value, true
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.capacity == 0 {
		return
	}

	v, ok := c.cache[key]

	if ok {
		entry := v.Value.(Entry[K, V])
		entry.value = value
		v.Value = entry
		c.list.MoveToFront(v)
		return
	} else {
		entry := Entry[K, V]{
			key:   key,
			value: value,
		}
		elem := c.list.PushFront(entry)
		c.cache[key] = elem
	}

	if len(c.cache) > c.capacity {
		tail := c.list.Back()
		oldKey := tail.Value.(Entry[K, V]).key
		c.list.Remove(tail)
		delete(c.cache, oldKey)
	}
}
