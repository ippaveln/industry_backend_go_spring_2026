package main

type Cache[K comparable, V any] struct {
	storage  map[K]V
	capacity int
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	return &Cache[K, V]{
		storage:  make(map[K]V, capacity),
		capacity: capacity,
	}
}

func (c *Cache[K, V]) Set(key K, value V) {
	if c.capacity == 0 {
		return
	}
	c.storage[key] = value
}

func (c *Cache[K, V]) Get(key K) (value V, ok bool) {
	if c.capacity == 0 {
		return value, false
	}
	value, ok = c.storage[key]
	return value, ok
}
