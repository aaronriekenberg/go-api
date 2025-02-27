package utils

import (
	"sync"
)

type GenericSyncMap[K comparable, V any] struct {
	syncMap sync.Map
}

func (gsm *GenericSyncMap[K, V]) Store(
	key K,
	value V,
) {

	gsm.syncMap.Store(key, value)
}

func (gsm *GenericSyncMap[K, V]) Load(
	key K,
) (value V, ok bool) {

	anyValue, ok := gsm.syncMap.Load(key)
	if !ok {
		return
	}

	value = anyValue.(V)

	return
}

func (gsm *GenericSyncMap[K, V]) LoadAndDelete(
	key K,
) (value V, loaded bool) {

	anyValue, loaded := gsm.syncMap.LoadAndDelete(key)

	if !loaded {
		return
	}

	value = anyValue.(V)
	return
}

func (gsm *GenericSyncMap[K, V]) Range(yield func(key K, value V) bool) {

	for anyKey, anyValue := range gsm.syncMap.Range {
		key := anyKey.(K)
		value := anyValue.(V)

		if !yield(key, value) {
			break
		}
	}
}

func (gsm *GenericSyncMap[K, V]) ValueRange(yield func(value V) bool) {

	for _, anyValue := range gsm.syncMap.Range {
		value := anyValue.(V)

		if !yield(value) {
			break
		}
	}
}
