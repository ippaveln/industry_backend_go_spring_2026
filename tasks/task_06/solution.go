package main

type Node[K comparable, V any] struct {
	key   K
	value V
	prev  *Node[K, V]
	next  *Node[K, V]
}

type LRUCache[K comparable, V any] struct {
	store    map[K]*Node[K, V]
	capacity int
	head     *Node[K, V]
	tail     *Node[K, V]
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		store:    make(map[K]*Node[K, V], capacity),
		capacity: capacity,
	}
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	var zero V

	if c.capacity == 0 {
		return zero, false
	}

	node, ok := c.store[key]
	if !ok {
		return zero, false
	}

	c.moveToHead(node)
	return node.value, true
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	if c.capacity == 0 {
		return
	}

	if node, ok := c.store[key]; ok {
		node.value = value
		c.moveToHead(node)
		return
	}

	node := &Node[K, V]{key: key, value: value}
	c.store[key] = node
	c.addToHead(node)

	if len(c.store) > c.capacity {
		c.removeTail()
	}
}

func (c *LRUCache[K, V]) addToHead(node *Node[K, V]) {
	node.prev = nil
	node.next = c.head

	if c.head != nil {
		c.head.prev = node
	}
	c.head = node

	if c.tail == nil {
		c.tail = node
	}
}

func (c *LRUCache[K, V]) moveToHead(node *Node[K, V]) {
	if node == c.head {
		return
	}

	if node.prev != nil {
		node.prev.next = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	}
	if node == c.tail {
		c.tail = node.prev
	}

	node.prev = nil
	node.next = c.head
	if c.head != nil {
		c.head.prev = node
	}
	c.head = node
}

func (c *LRUCache[K, V]) removeTail() {
	if c.tail == nil {
		return
	}

	delete(c.store, c.tail.key)

	if c.tail.prev != nil {
		c.tail.prev.next = nil
		c.tail = c.tail.prev
	} else {
		c.head = nil
		c.tail = nil
	}
}
