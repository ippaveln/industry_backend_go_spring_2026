package main

// LRU interface in main.go

type node[K comparable, V any] struct {
	key   K
	value V
	prev  *node[K, V]
	next  *node[K, V]
}

type LRUCache[K comparable, V any] struct {
	m        map[K]*node[K, V]
	head     *node[K, V]
	tail     *node[K, V]
	capacity int
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	if capacity <= 0 {
		return &LRUCache[K, V]{capacity: 0}
	}
	return &LRUCache[K, V]{
		m:        make(map[K]*node[K, V], capacity),
		capacity: capacity,
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

func (c *LRUCache[K, V]) moveToFront(n *node[K, V]) {
	if c.head == n {
		return
	}
	c.remove(n)
	c.pushFront(n)
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	if n, ok := c.m[key]; ok && c.capacity > 0 {
		c.moveToFront(n)
		return n.value, ok
	}
	var zero V
	return zero, false
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	if c.capacity == 0 {
		return
	}
	if n, ok := c.m[key]; ok {
		n.value = value
		c.moveToFront(n)
		return
	}

	c.m[key] = &node[K, V]{
		key:   key,
		value: value,
	}
	c.pushFront(c.m[key])

	if len(c.m) > c.capacity {
		delete(c.m, c.tail.key)
		c.remove(c.tail)
	}
}
