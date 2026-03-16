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
	return &LRUCache[K, V]{
		capacity: capacity,
		cache:    make(map[K]*node[K, V]),
	}
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	if c.capacity <= 0 {
		var zero V
		return zero, false
	}
	if node, exists := c.cache[key]; exists {
		c.moveToFront(node)
		return node.value, true
	}
	var zero V
	return zero, false
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	if c.capacity <= 0 {
		return
	}
	if node, exists := c.cache[key]; exists {
		node.value = value
		c.moveToFront(node)
		return
	}
	if len(c.cache) >= c.capacity {
		c.removeLRU()
	}
	newNode := &node[K, V]{
		key:   key,
		value: value,
	}
	c.addToFront(newNode)
	c.cache[key] = newNode
}

func (c *LRUCache[K, V]) moveToFront(n *node[K, V]) {
	if n == c.head {
		return
	}
	c.removeNode(n)
	c.addToFront(n)
}

func (c *LRUCache[K, V]) addToFront(n *node[K, V]) {
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

func (c *LRUCache[K, V]) removeNode(n *node[K, V]) {
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

func (c *LRUCache[K, V]) removeLRU() {
	if c.tail == nil {
		return
	}
	delete(c.cache, c.tail.key)
	c.removeNode(c.tail)
}
