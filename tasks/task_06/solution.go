package main

type node[K comparable, V any] struct {
	key   K
	value V
	prev  *node[K, V]
	next  *node[K, V]
}

type doubleLinkedList[K comparable, V any] struct {
	head *node[K, V]
	tail *node[K, V]
}

type LRUCache[K comparable, V any] struct {
	capacity int
	cache    map[K]*node[K, V]
	list     doubleLinkedList[K, V]
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		capacity: capacity,
		cache:    make(map[K]*node[K, V]),
	}
}

func (l *LRUCache[K, V]) moveToTail(n *node[K, V]) {
	if n == l.list.tail {
		return
	}

	if n.prev != nil {
		n.prev.next = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	}
	if n == l.list.head {
		l.list.head = n.next
	}
	n.prev = l.list.tail
	n.next = nil
	if l.list.tail != nil {
		l.list.tail.next = n
	}
	l.list.tail = n

	if l.list.head == nil {
		l.list.head = n
	}
}

func (l *LRUCache[K, V]) removeLRU() {
	if l.list.head == nil {
		return
	}
	delete(l.cache, l.list.head.key)
	l.list.head = l.list.head.next
	if l.list.head != nil {
		l.list.head.prev = nil
	} else {
		l.list.tail = nil
	}
}

func (l *LRUCache[K, V]) Get(key K) (V, bool) {
	n, ok := l.cache[key]
	if !ok {
		var zeroValue V
		return zeroValue, false
	}

	l.moveToTail(n)
	return n.value, true
}

func (l *LRUCache[K, V]) Set(key K, value V) {
	if l.capacity <= 0 {
		return
	}
	if n, ok := l.cache[key]; ok {
		n.value = value
		l.moveToTail(n)
		return
	}
	n := &node[K, V]{
		key:   key,
		value: value,
	}
	if len(l.cache) >= l.capacity {
		l.removeLRU()
	}

	l.cache[key] = n
	l.moveToTail(n)
}
