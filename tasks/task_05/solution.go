package main

type Cache[K comparable, V any] struct {
	capacity  int
	cachedMap map[K]V
}

func NewCache[K comparable, V any](capacity int) Cache[K, V] {
	return Cache[K, V]{
		capacity:  capacity,
		cachedMap: map[K]V{},
	}
}

func (c *Cache[K, V]) Set(key K, value V) {
	if c.capacity == 0 {
		return
	}

	c.cachedMap[key] = value
}

func (c *Cache[K, V]) Get(key K) (V, bool) {

	if c.capacity == 0 {
		var zeroValue V
		return zeroValue, false
	}

	value, isKeyExists := c.cachedMap[key]
	return value, isKeyExists
}
