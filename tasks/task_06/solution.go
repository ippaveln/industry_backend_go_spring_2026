package main

type node[K comparable, V any] struct {
	key   K
	value V
	prev  *node[K, V]
	next  *node[K, V]
}

type linkedList[K comparable, V any] struct {
	head *node[K, V]
	tail *node[K, V]
}

func (l *linkedList[K, V]) moveToFront(n *node[K, V]) {
	if l.head == n {
		return
	}
	l.remove(n)
	l.pushFront(n)
}

func newLinkedList[K comparable, V any]() *linkedList[K, V] {
	return &linkedList[K, V]{}
}

func (l *linkedList[K, V]) pushFront(n *node[K, V]) {
	n.prev = nil
	n.next = l.head

	if l.head != nil {
		l.head.prev = n
	} else {
		l.tail = n
	}

	l.head = n
}

func (l *linkedList[K, V]) remove(n *node[K, V]) {
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		l.head = n.next
	}

	if n.next != nil {
		n.next.prev = n.prev
	} else {
		l.tail = n.prev
	}

	n.prev = nil
	n.next = nil
}

func (l *linkedList[K, V]) removeBack() *node[K, V] {
	if l.tail == nil {
		return nil
	}
	n := l.tail
	l.remove(n)
	return n
}

type LRUCache[K comparable, V any] struct {
	items          map[K]*node[K, V]
	expirationList *linkedList[K, V]
	capacity       int
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		items:          make(map[K]*node[K, V], capacity),
		expirationList: newLinkedList[K, V](),
		capacity:       capacity}
}

func (lru *LRUCache[K, V]) Get(key K) (V, bool) {
	if node, ok := lru.items[key]; ok {
		lru.expirationList.moveToFront(node)
		return node.value, true
	}
	var zeroValue V
	return zeroValue, false
}

func (lru *LRUCache[K, V]) Set(key K, value V) {
	if lru.capacity == 0 {
		return
	}
	if node, ok := lru.items[key]; ok {
		node.value = value
		lru.expirationList.moveToFront(node)
		return
	}
	newNode := &node[K, V]{key: key, value: value}
	lru.items[key] = newNode
	lru.expirationList.pushFront(newNode)
	if len(lru.items) > lru.capacity {
		oldNode := lru.expirationList.removeBack()
		delete(lru.items, oldNode.key)
	}
}
