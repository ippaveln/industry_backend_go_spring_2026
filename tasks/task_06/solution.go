package main

type node[K comparable, V any] struct {
	key   K
	value V
	prev  *node[K, V]
	next  *node[K, V]
}

type LRUCache[K comparable, V any] struct {
	cap   int
	items map[K]*node[K, V]
	head  *node[K, V]
	tail  *node[K, V]
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		cap:   capacity,
		items: make(map[K]*node[K, V]),
	}
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	var zero V
	n, ok := c.items[key]
	if !ok || c.cap == 0 {
		return zero, false
	}
	c.front(n)
	return n.value, true
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	if c.cap == 0 {
		return
	}

	if n, ok := c.items[key]; ok {
		n.value = value
		c.front(n)
		return
	}

	n := &node[K, V]{key: key, value: value}
	c.items[key] = n
	c.pushFront(n)

	if len(c.items) > c.cap {
		delete(c.items, c.tail.key)
		c.remove(c.tail)
	}
}

func (c *LRUCache[K, V]) front(n *node[K, V]) {
	if c.head == n {
		return
	}
	c.remove(n)
	c.pushFront(n)
}

func (c *LRUCache[K, V]) pushFront(n *node[K, V]) {
	n.prev = nil
	n.next = c.head

	if c.head != nil {
		c.head.prev = n
	} else {
		c.tail = n
	}

	c.head = n
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
}