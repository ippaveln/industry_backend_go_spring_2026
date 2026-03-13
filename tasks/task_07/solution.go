package main

import "sync"

type LRU[K comparable, V any] interface {
	Get(key K) (value V, ok bool)
	Set(key K, value V)
}

type Node[K comparable, V any] struct {
	Prev  *Node[K, V]
	Key   K
	Value V
	Next  *Node[K, V]
}

type DoublyLinkedList[K comparable, V any] struct {
	Head   *Node[K, V]
	Tail   *Node[K, V]
	Length int
}

func NewDoublyLinkedList[K comparable, V any]() *DoublyLinkedList[K, V] {
	return &DoublyLinkedList[K, V]{}
}

func (l *DoublyLinkedList[K, V]) InsertAtHead(node *Node[K, V]) {
	if l.Head == nil {
		l.Head = node
		l.Tail = node
	} else {
		node.Next = l.Head
		node.Prev = nil
		l.Head.Prev = node
		l.Head = node
	}
	l.Length++
}

func (l *DoublyLinkedList[K, V]) Delete(node *Node[K, V]) *Node[K, V] {
	if node == nil {
		return nil
	}

	if node.Next != nil {
		node.Next.Prev = node.Prev
	} else {
		l.Tail = node.Prev
	}

	if node.Prev != nil {
		node.Prev.Next = node.Next
	} else {
		l.Head = node.Next
	}

	node.Prev = nil
	node.Next = nil

	l.Length--
	return node
}

func (l *DoublyLinkedList[K, V]) Refresh(node *Node[K, V]) /* *Node[K, V] */ {
	l.Delete(node)
	l.InsertAtHead(node)
	// return tmp
}

// For debug only
func (l *DoublyLinkedList[K, V]) ForEach(f func(node *Node[K, V])) {
	if l.Head == nil {
		return
	}

	current := l.Head
	for current != nil {
		f(current)
		current = current.Next
	}
}

// ---

type LRUCache[K comparable, V any] struct {
	Cache    map[K]*Node[K, V]
	list     DoublyLinkedList[K, V]
	Capacity int
	mtx      sync.Mutex
}

// Get implements [LRU].
func (l *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	if v, ok := l.Cache[key]; ok {
		l.list.Refresh(v)
		value = v.Value
		return value, ok
	}
	var nothing V
	return nothing, ok
}

// Set implements [LRU].
func (l *LRUCache[K, V]) Set(key K, value V) {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	
	if val, ok := l.Cache[key]; ok {
		val.Value = value
		l.list.Refresh(val)
		return
	}

	node := &Node[K, V]{Prev: nil, Key: key, Value: value, Next: nil}
	l.list.InsertAtHead(node)
	l.Cache[key] = node

	if l.list.Length > l.Capacity {
		delete(l.Cache, l.list.Tail.Key)
		l.list.Delete(l.list.Tail)
	}
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		Cache:    make(map[K]*Node[K, V], capacity),
		Capacity: capacity,
		list:     DoublyLinkedList[K, V]{},
	}
}
