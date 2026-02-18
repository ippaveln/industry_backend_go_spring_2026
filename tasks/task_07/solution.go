package main

import "sync"

type LRU[K comparable, V any] interface {
	Get(key K) (value V, ok bool)
	Set(key K, value V)
}

type Node[K comparable, V any] struct {
	key   K
	value V
	next  *Node[K, V]
	prev  *Node[K, V]
}

type DoubleLinkedList[K comparable, V any] struct {
	head  *Node[K, V]
	tail  *Node[K, V]
	count int
}

func (d *DoubleLinkedList[K, V]) Append(key K, value V) *Node[K, V] {
	node := &Node[K, V]{key: key, value: value}

	if d.tail == nil {
		d.head = node
		d.tail = node
		d.count++
		return node
	}

	d.tail.next = node
	node.prev = d.tail
	d.tail = node
	d.count++
	return node
}

func (d *DoubleLinkedList[K, V]) Remove(node *Node[K, V]) *Node[K, V] {
	if node == nil {
		return nil
	}

	if node.prev != nil {
		node.prev.next = node.next
	} else {
		d.head = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		d.tail = node.prev
	}

	node.next = nil
	node.prev = nil

	d.count--
	return node
}

func (d *DoubleLinkedList[K, V]) Pop() *Node[K, V] {
	if d.head == nil {
		return nil
	}
	return d.Remove(d.head)
}

type LRUCache[K comparable, V any] struct {
	mu         sync.Mutex
	values     map[K]*Node[K, V]
	linkedList DoubleLinkedList[K, V]
	cap        int
}

func NewLRUCache[K comparable, V any](cap int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		values: make(map[K]*Node[K, V], cap),
		cap:    cap,
	}
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cap == 0 {
		return
	}

	if n, ok := c.values[key]; ok {
		delete(c.values, n.key)
		c.linkedList.Remove(n)
	}

	if c.linkedList.count == c.cap {
		lru := c.linkedList.Pop()
		delete(c.values, lru.key)
	}
	node := c.linkedList.Append(key, value)
	c.values[key] = node
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var zero V
	if c.cap == 0 {
		return zero, false
	}
	node, ok := c.values[key]
	if ok {
		delete(c.values, node.key)
		c.linkedList.Remove(node)
		node = c.linkedList.Append(node.key, node.value)
		c.values[key] = node
		return node.value, true
	}

	return zero, false
}
