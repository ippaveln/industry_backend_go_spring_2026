package main

import (
	"container/list"
	"sync"
)

type LRU[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V)
}

type entry[K comparable, V any] struct {
	key   K
	value V
}

type LRUCache[K comparable, V any] struct {
	mu       sync.Mutex
	capacity int
	items    map[K]*list.Element
	list     *list.List
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	if capacity < 0 {
		capacity = 0
	}
	return &LRUCache[K, V]{
		capacity: capacity,
		items:    make(map[K]*list.Element),
		list:     list.New(),
	}
}

func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.capacity <= 0 {
		return
	}

	elem, exists := c.items[key]
	if !exists {
		return
	}

	c.list.MoveToFront(elem)
	ent := elem.Value.(entry[K, V])
	return ent.value, true
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.capacity <= 0 {
		return
	}

	if elem, exists := c.items[key]; exists {
		ent := elem.Value.(entry[K, V])
		ent.value = value
		elem.Value = ent
		c.list.MoveToFront(elem)
		return
	}

	newEnt := entry[K, V]{key: key, value: value}
	newElem := c.list.PushFront(newEnt)
	c.items[key] = newElem

	if c.list.Len() > c.capacity {
		back := c.list.Back()
		if back != nil {
			oldEnt := back.Value.(entry[K, V])
			delete(c.items, oldEnt.key)
			c.list.Remove(back)
		}
	}
}
