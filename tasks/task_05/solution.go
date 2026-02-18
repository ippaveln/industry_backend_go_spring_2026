package main

type Cache[K comparable, V any] struct {
	cache    map[K]V
	capacity int
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	res := Cache[K, V]{}
	res.cache = make(map[K]V, capacity)
	res.capacity = capacity
	return &res
}

func (c *Cache[K, V]) Set(key K, value V) {
	if c.capacity == 0 {
		return
	}
	c.cache[key] = value
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	v, ok := c.cache[key]
	return v, ok
}
