package main

type Cache[K comparable, V any] struct {
	enabled bool
	data    map[K]V
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	if capacity <= 0 {
		return &Cache[K, V]{enabled: false}
	}
	newHintMap := make(map[K]V, capacity)

	return &Cache[K, V]{
		enabled: true,
		data:    newHintMap,
	}
}

func (c *Cache[K, V]) Set(key K, value V) {
	if !c.enabled {
		return
	}
	c.data[key] = value
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	var zeroValue V
	if !c.enabled {
		return zeroValue, false
	}
	value, ok := c.data[key]
	if !ok {
		return zeroValue, false
	}

	return value, true
}
