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
	mu       sync.Mutex
	capacity int
	items    map[K]*node[K, V]
	head     *node[K, V]
	tail     *node[K, V]
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var zeroValue V

	n, ok := c.items[key]
	if !ok {
		return zeroValue, false
	}

	c.moveToFront(n)

	return n.value, true
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.capacity == 0 {
		return
	}

	if n, ok := c.items[key]; ok {
		n.value = value
		c.moveToFront(n)
		return
	}

	n := &node[K, V]{key: key, value: value}
	c.items[key] = n
	c.pushFront(n)

	if len(c.items) > c.capacity {
		c.removeTail()
	}
}

func (c *LRUCache[K, V]) pushFront(n *node[K, V]) {
	n.prev = nil
	n.next = c.head
	if c.head != nil {
		c.head.prev = n
	}
	c.head = n
	if c.tail == nil {
		c.tail = n
	}
}

func (c *LRUCache[K, V]) remove(n *node[K, V]) {
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		c.head = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	} else {
		c.tail = n.prev
	}
	n.prev, n.next = nil, nil
}

func (c *LRUCache[K, V]) moveToFront(n *node[K, V]) {
	if n == c.head {
		return
	}
	c.remove(n)
	c.pushFront(n)
}

func (c *LRUCache[K, V]) removeTail() {
	if c.tail == nil {
		return
	}
	victim := c.tail
	c.remove(victim)
	delete(c.items, victim.key)
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		capacity: capacity,
		items:    make(map[K]*node[K, V]),
	}
}
