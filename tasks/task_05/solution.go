package main

type Cache[K comparable, V any] struct {
	capacity int
	items map[K]V
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	return &Cache[K, V]{
		capacity:  capacity,
		items: make(map[K]V),
	}
}

func (c *Cache[K, V]) Set(key K, value V) {
	if c.capacity <= 0 {
		return
	}
	c.items[key] = value
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	if val, ok := c.items[key]; c.capacity > 0 && ok {
		return val, true
	}
	var zero V
	return zero, false
}
