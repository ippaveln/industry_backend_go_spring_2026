package main

type Cache[K comparable, V any] struct {
    Storage map[K]V
	Capacity int
}


func NewCache [K comparable, V any] (capacity int) *Cache[K, V] {
	return &Cache[K, V]{
		Storage: make(map[K]V, capacity),
		Capacity: capacity,
	}
}

func (c *Cache[K, V]) Set(key K, value V) {
	if c.Capacity > 0 {
		c.Storage[key] = value
	}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	v, ok := c.Storage[key]
	return v, ok
}