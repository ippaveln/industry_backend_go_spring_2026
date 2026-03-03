package main

type Cache[K comparable, V any] struct {
	memory map[K]V
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	if capacity <= 0 {
		return &Cache[K, V]{memory: nil}
	}
	return &Cache[K, V]{memory: make(map[K]V, capacity)}
}

func (c *Cache[K, V]) Set(key K, value V) {
	if c.memory != nil {
		c.memory[key] = value
	}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	// reading from nil-map safe, returns zero value
	val, ok := c.memory[key]
	return val, ok
}
