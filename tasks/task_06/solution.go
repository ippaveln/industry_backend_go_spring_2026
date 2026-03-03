package main

type LRUCache[K comparable, V any] struct {
    data     map[K]*node[K, V]
    capacity int
    head     *node[K, V]
    tail     *node[K, V]
}

type node[K comparable, V any] struct {
    key   K
    value V
    prev  *node[K, V]
    next  *node[K, V]
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
    return &LRUCache[K, V]{
        data:     make(map[K]*node[K, V], capacity),
        capacity: capacity,
    }
}

func (c *LRUCache[K, V]) moveToFront(n *node[K, V]) {
    if c.head == n {
        return
    }

    if n.prev != nil {
        n.prev.next = n.next
    }
    if n.next != nil {
        n.next.prev = n.prev
    }
    if c.tail == n {
        c.tail = n.prev
    }

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

func (c *LRUCache[K, V]) addToFront(n *node[K, V]) {
    if c.head == nil {
        c.head = n
        c.tail = n
        n.prev = nil
        n.next = nil
        return
    }

    n.next = c.head
    n.prev = nil
    c.head.prev = n
    c.head = n
}

func (c *LRUCache[K, V]) removeOldest() {
    if c.tail == nil {
        return
    }

    oldestKey := c.tail.key

    c.tail = c.tail.prev
    if c.tail != nil {
        c.tail.next = nil
    } else {
        c.head = nil
    }

    delete(c.data, oldestKey)
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
    if c.capacity == 0 {
        var zero V
        return zero, false
    }
    node, ok := c.data[key]
    if !ok {
        var zero V
        return zero, false
    }
    c.moveToFront(node)
    return node.value, true
}

func (c *LRUCache[K, V]) Set(key K, value V) {
    if c.capacity == 0 {
        return
    }

    if node, ok := c.data[key]; ok {
        node.value = value
        c.moveToFront(node)
        return
    }

    if len(c.data) >= c.capacity {
        c.removeOldest()
    }

    newNode := &node[K, V]{
        key:   key,
        value: value,
    }

    c.data[key] = newNode
    c.addToFront(newNode)
}