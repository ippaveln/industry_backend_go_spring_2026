package main

type node[K comparable, V any] struct {
	key   K
	value V
	prev  *node[K, V]
	next  *node[K, V]
}

type LRUCache[K comparable, V any] struct {
	capacity int
	cache    map[K]*node[K, V]
	head     *node[K, V]
	tail     *node[K, V]
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	head := &node[K, V]{}
	tail := &node[K, V]{}
	head.next = tail
	tail.prev = head

	return &LRUCache[K, V]{
		capacity: capacity,
		cache:    make(map[K]*node[K, V]),
		head:     head,
		tail:     tail,
	}
}

func (c *LRUCache[K, V]) removeNode(n *node[K, V]) {
	n.prev.next = n.next
	n.next.prev = n.prev
}

func (c *LRUCache[K, V]) addToHead(n *node[K, V]) {
	n.prev = c.head
	n.next = c.head.next
	c.head.next.prev = n
	c.head.next = n
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	if n, ok := c.cache[key]; ok && c.capacity > 0 {
		c.removeNode(n)
		c.addToHead(n)
		return n.value, true
	}
	var zero V
	return zero, false
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	if c.capacity <= 0 {
		return
	}
	if n, ok := c.cache[key]; ok {
		n.value = value
		c.removeNode(n)
		c.addToHead(n)
	} else {
		n := &node[K, V]{key: key, value: value}
		c.cache[key] = n
		c.addToHead(n)
		if len(c.cache) > c.capacity {
			lru := c.tail.prev
			c.removeNode(lru)
			delete(c.cache, lru.key)
		}
	}
}
