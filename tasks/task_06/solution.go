package main

import "container/list"

type LRUCache[K comparable, V any] struct {
	capacity int
	items    map[K]*list.Element
	list     *list.List
}

type entry[K comparable, V any] struct {
	key   K
	value V
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		capacity: capacity,
		items:    make(map[K]*list.Element),
		list:     list.New(),
	}
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	if c.capacity <= 0 {
		var zero V
		return zero, false
	}

	elem, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}

	c.list.MoveToFront(elem)
	return elem.Value.(entry[K, V]).value, true
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	if c.capacity <= 0 {
		return
	}

	if elem, ok := c.items[key]; ok {
		ent := elem.Value.(entry[K, V])
		ent.value = value
		elem.Value = ent
		c.list.MoveToFront(elem)
		return
	}

	ent := entry[K, V]{key: key, value: value}
	elem := c.list.PushFront(ent)
	c.items[key] = elem

	if c.list.Len() > c.capacity {
		back := c.list.Back()
		if back != nil {
			kv := back.Value.(entry[K, V])
			delete(c.items, kv.key)
			c.list.Remove(back)
		}
	}
}
