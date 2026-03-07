package main

type Cache[K comparable, V any] struct {
	cap  int
	data map[K]V
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	return &Cache[K, V]{
		cap:  capacity,
		data: make(map[K]V, capacity),
	}
}

func (c *Cache[K, V]) Set(key K, value V) {
	if c.cap == 0 {
		return
	}
	c.data[key] = value
}

func (c *Cache[K, V]) Get(key K) (value V, ok bool) {
	value, ok = c.data[key]
	return value, ok
}
