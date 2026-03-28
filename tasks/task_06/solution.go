package main

type node[K comparable, V any] struct {
	key        K
	val        V
	prev, next *node[K, V]
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
	n.prev = nil
	n.next = nil
}

func (c *LRUCache[K, V]) pushFront(n *node[K, V]) {
	n.next = c.head
	n.prev = nil
	if c.head != nil {
		c.head.prev = n
	}
	c.head = n
	if c.tail == nil {
		c.tail = n
	}
}

type LRUCache[K comparable, V any] struct {
	capacity int
	items    map[K]*node[K, V]
	head     *node[K, V]
	tail     *node[K, V]
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{capacity: capacity, items: make(map[K]*node[K, V], capacity)}
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	n, ok := c.items[key]
	if !ok {
		var zeroValue V
		return zeroValue, false
	}
	c.remove(n)
	c.pushFront(n)
	return n.val, ok
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	if c.capacity == 0 {
		return
	}
	if n, ok := c.items[key]; ok {
		n.val = value
		c.remove(n)
		c.pushFront(n)
		return
	}
	n := &node[K, V]{key: key, val: value}
	c.items[key] = n
	c.pushFront(n)
	if len(c.items) > c.capacity {
		tail := c.tail
		c.remove(tail)
		delete(c.items, tail.key)
	}
}
