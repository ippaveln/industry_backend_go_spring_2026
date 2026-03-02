package main

import (
	"container/list"
	"sync"
)

type LRU[K comparable, V any] interface {
	Get(key K) (value V, ok bool)
	Set(key K, value V)
}

type node[K comparable, V any] struct {
	key   K
	value V
}

type LRUCache[K comparable, V any] struct {
	mu       sync.Mutex
	capacity int
	list     *list.List
	items    map[K]*list.Element
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		capacity: capacity,
		list:     list.New(),
		items:    make(map[K]*list.Element),
	}
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.capacity == 0 {
		var zero V
		return zero, false
	}

	element, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}

	c.list.MoveToFront(element)

	return element.Value.(node[K, V]).value, true
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.capacity == 0 {
		return
	}

	if element, ok := c.items[key]; ok {
		element.Value = node[K, V]{key: key, value: value}
		c.list.MoveToFront(element)
		return
	}

	newNode := node[K, V]{key: key, value: value}
	element := c.list.PushFront(newNode)
	c.items[key] = element

	if len(c.items) > c.capacity {
		oldest := c.list.Back()
		c.list.Remove(oldest)
		delete(c.items, oldest.Value.(node[K, V]).key)
	}
}
