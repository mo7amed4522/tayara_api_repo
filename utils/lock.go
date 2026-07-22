package utils

import "sync"

type KeyLocker struct {
	mu    sync.Mutex // mutex for the map operations
	locks sync.Map   // map[string]*sync.Mutex
}

func (kl *KeyLocker) Lock(key string) {
	actual, _ := kl.locks.LoadOrStore(key, &sync.Mutex{})
	m := actual.(*sync.Mutex)
	m.Lock()
}

func (kl *KeyLocker) Unlock(key string) {
	actual, ok := kl.locks.Load(key)
	if !ok {
		return
	}
	m := actual.(*sync.Mutex)
	m.Unlock()
	kl.locks.Delete(key)
}
