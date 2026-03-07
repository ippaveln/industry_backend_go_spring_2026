package main

import (
	"container/list"
	"sync"
)

type LRU[K comparable, V any] interface {
	Get(key K) (value V, ok bool)
	Set(key K, value V)
}

type LRUCache[K comparable, V any] struct {
	queue *list.List
	cap   int
	data  map[K]*list.Element
	mu    sync.Mutex
}

type item[K comparable, V any] struct {
	key   K
	value V
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		queue: list.New(),
		cap:   capacity,
		data:  make(map[K]*list.Element, capacity),
	}
}

func (lru *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if elem, ok := lru.data[key]; ok {
		lru.queue.MoveToFront(elem)
		return elem.Value.(*item[K, V]).value, true
	}
	return value, false
}

func (lru *LRUCache[K, V]) Set(key K, value V) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if lru.cap <= 0 {
		return
	}

	if e, ok := lru.data[key]; ok {
		e.Value.(*item[K, V]).value = value
		lru.queue.MoveToFront(e)
		return
	}

	if len(lru.data) >= lru.cap {
		oldElement := lru.queue.Back()
		oldItem := lru.queue.Remove(oldElement).(*item[K, V])
		delete(lru.data, oldItem.key)
	}

	newEntry := &item[K, V]{
		key:   key,
		value: value,
	}
	newElement := lru.queue.PushFront(newEntry)
	lru.data[key] = newElement
}
