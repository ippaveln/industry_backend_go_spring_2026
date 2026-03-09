package main

import (
	"container/list"
	"sync"
)

type Entry[K comparable, V any] struct {
	key   K
	value V
}

func NewEntry[K comparable, V any](key K, value V) Entry[K, V] {
	return Entry[K, V]{key: key, value: value}
}

type LRU[K comparable, V any] interface {
	Get(key K) (value V, ok bool)
	Set(key K, value V)
}

type LRUCache[K comparable, V any] struct {
	capacity              int
	linkedListElementsMap map[K]*list.Element
	cachedLinkedList      *list.List
	mutex                 sync.Mutex
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		capacity:              capacity,
		linkedListElementsMap: make(map[K]*list.Element),
		cachedLinkedList:      list.New(),
	}
}

func (L *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	L.mutex.Lock()
	defer L.mutex.Unlock()

	var zeroValue V

	if L.capacity == 0 {
		return zeroValue, false
	}

	foundValue, isExists := L.linkedListElementsMap[key]
	if isExists {
		L.makeMRU(key)

		return foundValue.Value.(Entry[K, V]).value, true
	}
	return zeroValue, false
}

func (L *LRUCache[K, V]) Set(key K, value V) {
	L.mutex.Lock()
	defer L.mutex.Unlock()

	if L.capacity == 0 {
		return
	}
	_, isExists := L.linkedListElementsMap[key]

	if isExists {
		L.updateExistingPair(key, value)
		L.makeMRU(key)

		return
	}
	L.processCacheMiss(key, value)
	L.removeLRU()
}

func (L *LRUCache[K, V]) updateExistingPair(key K, value V) {
	oldListEl := L.linkedListElementsMap[key]
	newPair := NewEntry(key, value)
	oldListEl.Value = newPair
}

func (L *LRUCache[K, V]) processCacheMiss(key K, value V) {
	newPair := NewEntry(key, value)

	L.cachedLinkedList.PushFront(newPair)
	L.linkedListElementsMap[key] = L.cachedLinkedList.Front()
}

func (L *LRUCache[K, V]) removeLRU() {
	if tail := L.cachedLinkedList.Back(); L.cachedLinkedList.Len() > L.capacity && tail != nil {
		delete(L.linkedListElementsMap, tail.Value.(Entry[K, V]).key)

		L.cachedLinkedList.Remove(tail)
	}
}

func (L *LRUCache[K, V]) makeMRU(key K) {
	accessedListElement := L.linkedListElementsMap[key]
	L.cachedLinkedList.MoveToFront(accessedListElement)
}

// Space complexity - O(n)

// Set time complexity - amortized O(1)
// Get time complexity - amortized O(1)
