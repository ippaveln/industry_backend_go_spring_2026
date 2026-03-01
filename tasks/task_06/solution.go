package main

type element[K comparable, V any] struct {
	key   K
	value V
	prev  *element[K, V]
	next  *element[K, V]
}

type LRUCache[K comparable, V any] struct {
	capacity int
	size     int
	head     *element[K, V]
	tail     *element[K, V]
	items    map[K]*element[K, V]
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		capacity: capacity,
		items:    make(map[K]*element[K, V]),
	}
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	if c.capacity == 0 {
		var zero V
		return zero, false
	}
	if el, ok := c.items[key]; ok {
		c.moveToFront(el)
		return el.value, true
	}
	var zero V
	return zero, false
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	if c.capacity == 0 {
		return
	}
	if el, ok := c.items[key]; ok {
		el.value = value
		c.moveToFront(el)
		return
	}
	if c.size == c.capacity {
		c.evictLRU()
	}
	el := &element[K, V]{key: key, value: value}
	c.items[key] = el
	c.addToFront(el)
	c.size++
}

func (c *LRUCache[K, V]) moveToFront(el *element[K, V]) {
	if c.head == el {
		return
	}
	if el.prev != nil {
		el.prev.next = el.next
	}
	if el.next != nil {
		el.next.prev = el.prev
	}
	if c.tail == el {
		c.tail = el.prev
	}
	el.prev = nil
	el.next = c.head
	if c.head != nil {
		c.head.prev = el
	}
	c.head = el
	if c.tail == nil {
		c.tail = el
	}
}

func (c *LRUCache[K, V]) addToFront(el *element[K, V]) {
	el.prev = nil
	el.next = c.head
	if c.head != nil {
		c.head.prev = el
	}
	c.head = el
	if c.tail == nil {
		c.tail = el
	}
}

func (c *LRUCache[K, V]) evictLRU() {
	if c.tail == nil {
		return
	}
	delete(c.items, c.tail.key)
	if c.tail.prev != nil {
		c.tail.prev.next = nil
	} else {
		c.head = nil
	}
	c.tail = c.tail.prev
	c.size--
}
