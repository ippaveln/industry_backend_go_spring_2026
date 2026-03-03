package main

import "sync"

type LRU[K comparable, V any] interface {
	Get(key K) (value V, ok bool)
	Set(key K, value V)
}

type Node[K comparable, V any] struct {
	key  K
	val  V
	next *Node[K, V]
	prev *Node[K, V]
}

type LRUCache[K comparable, V any] struct {
	capacity int
	data     map[K]*Node[K, V]
	head     *Node[K, V]
	tail     *Node[K, V]
	mu       sync.Mutex
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	data := make(map[K]*Node[K, V])
	head := &Node[K, V]{}
	tail := &Node[K, V]{}
	head.next = tail
	tail.prev = head

	return &LRUCache[K, V]{
		capacity: capacity,
		data:     data,
		head:     head,
		tail:     tail,
	}
}

func (LRU *LRUCache[K, V]) remove(node *Node[K, V]) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

func (LRU *LRUCache[K, V]) addToHead(node *Node[K, V]) {
	node.next = LRU.head.next
	node.prev = LRU.head

	LRU.head.next.prev = node
	LRU.head.next = node
}

func (LRU *LRUCache[K, V]) Set(key K, val V) {
	if LRU.capacity == 0 {
		return
	}

	LRU.mu.Lock()
	defer LRU.mu.Unlock()

	node, ok := LRU.data[key]
	if ok {
		LRU.remove(node)
		LRU.addToHead(node)
		node.val = val
		return
	}

	if len(LRU.data) >= LRU.capacity {
		node := LRU.tail.prev
		LRU.remove(node)
		delete(LRU.data, node.key)
	}

	node = &Node[K, V]{
		key: key,
		val: val,
	}

	LRU.data[key] = node
	LRU.addToHead(node)
}

func (LRU *LRUCache[K, V]) Get(key K) (V, bool) {
	LRU.mu.Lock()
	defer LRU.mu.Unlock()

	node, ok := LRU.data[key]
	if !ok {
		var z V
		return z, false
	}

	LRU.remove(node)
	LRU.addToHead(node)

	return node.val, ok
}
