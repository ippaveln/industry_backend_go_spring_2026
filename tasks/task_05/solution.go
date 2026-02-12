package main

type Cache[K comparable, V any] struct {
	data     map[K]V
	capacity int
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	var m map[K]V
	if capacity > 0 {
		m = make(map[K]V, capacity)
	} else {
		capacity = 0
	}
	return &Cache[K, V]{
		data:     m,
		capacity: capacity,
	}
}

func (cache *Cache[K, V]) Set(key K, value V) {
	if cache.capacity == 0 {
		return
	}

	cache.data[key] = value
}

func (cache *Cache[K, V]) Get(key K) (V, bool) {
	val, ok := cache.data[key]
	return val, ok
}
