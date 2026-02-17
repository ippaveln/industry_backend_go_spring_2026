package main
type Cache[K comparable, V any] struct {
	data map[K]V
	cap  int
}

func NewCache[K comparable, V any](cap int) *Cache[K, V] {
	return &Cache[K, V]{
		data: make(map[K]V, cap),
		cap:  cap,
	}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	if c.cap == 0 {
		var zero V
		return zero, false
	}
	value, ok := c.data[key]
	if !ok {
		var zero V
		return zero, false
	}
	return value, true
}

func (c *Cache[K, V]) Set(key K, val V) {
	c.data[key] = val
}
