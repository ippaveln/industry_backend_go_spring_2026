package main

import (
    "sync"
)

type LRU[K comparable, V any] interface {
    Get(key K) (V, bool)
    Set(key K, value V)
}
type node[K comparable, V any] struct {
    key   K
    value V
    prev  *node[K, V]
    next  *node[K, V]
}
type LRUCache[K comparable, V any] struct {
    mu       sync.Mutex
    capacity int
    data     map[K]*node[K, V]
    head     *node[K, V]
    tail     *node[K, V]
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
    return &LRUCache[K, V]{
        capacity: capacity,
        data:     make(map[K]*node[K, V], capacity),
    }
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
    c.mu.Lock()
    defer c.mu.Unlock()

    var zero V
    if c.capacity == 0 {
        return zero, false
    }

    if node, ok := c.data[key]; ok {
        c.moveToFront(node)
        return node.value, true
    }
    return zero, false
}

func (c *LRUCache[K, V]) Set(key K, value V) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if c.capacity == 0 {
        return
    }
    if node, ok := c.data[key]; ok {
        node.value = value
        c.moveToFront(node)
        return
    }

    if len(c.data) >= c.capacity {
        c.removeLast()
    }

    newNode := &node[K, V]{
        key:   key,
        value: value,
    }
    c.data[key] = newNode
    c.addToFront(newNode)
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

func (c *LRUCache[K, V]) removeLast() {
    if c.tail == nil {
        return
    }

    delete(c.data, c.tail.key)

    if c.tail.prev != nil {
        c.tail.prev.next = nil
        c.tail = c.tail.prev
    } else {
        c.head = nil
        c.tail = nil
    }
}