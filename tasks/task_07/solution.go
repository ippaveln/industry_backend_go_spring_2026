package main

import "sync"

type LRU[K comparable, V any] interface {
	Get(key K) (value V, ok bool)
	Set(key K, value V)
}

type node[K comparable, V any] struct {
	key   K
	value V
	prev  *node[K, V]
	next  *node[K, V]
}

type LRUCache[K comparable, V any] struct {
	mutex    sync.Mutex
	items    map[K]*node[K, V]
	capacity int
	MRU      *node[K, V]
	LRU      *node[K, V]
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		capacity: capacity,
		items:    make(map[K]*node[K, V], capacity),
		MRU:      nil,
		LRU:      nil}
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.capacity == 0 {
		return
	}

	if node, ok := c.items[key]; ok {
		node.value = value
		c.moveToMRU(node)
		return
	}

	node := &node[K, V]{key: key, value: value}
	c.items[key] = node
	if len(c.items) == 1 {
		c.MRU = node
		c.LRU = node
		return
	}
	c.MRU.prev = node
	node.next = c.MRU
	c.MRU = node
	if len(c.items) > c.capacity {
		old := c.LRU
		c.LRU = c.LRU.prev
		c.LRU.next = nil
		delete(c.items, old.key)
	}
	return
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.capacity == 0 {
		var val V
		return val, false
	}

	node, ok := c.items[key]
	if !ok {
		var val V
		return val, false
	}

	c.moveToMRU(node)
	return node.value, true
}

func (c *LRUCache[K, V]) moveToMRU(node *node[K, V]) {
	if !(len(c.items) == 1 || node == c.MRU) {
		if node == c.LRU {
			c.LRU = node.prev
		}
		if node.next != nil {
			node.next.prev = node.prev
		}
		if node.prev != nil {
			node.prev.next = node.next
		}
		c.MRU.prev = node
		node.next = c.MRU
		c.MRU = node
	}
}
