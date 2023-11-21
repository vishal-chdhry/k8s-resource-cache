package resourcecache

import (
	"context"
	"sync"

	"github.com/dgraph-io/ristretto"
	"k8s.io/client-go/tools/cache"
)

type CacheEntry struct {
	Lister cache.GenericNamespaceLister
	data   interface{}
	stop   context.CancelFunc
}

type Cache struct {
	sync.Mutex
	store *ristretto.Cache
}

func NewListerCache() (*Cache, error) {
	config := ristretto.Config{
		MaxCost:     100 * 1000 * 1000, // 100 MB
		NumCounters: 10 * 100,          // 100 entries
		BufferItems: 64,
		OnExit:      ristrettoOnExit,
		OnEvict:     ristrettoOnEvict,
	}

	rcache, err := ristretto.NewCache(&config)
	if err != nil {
		return nil, err
	}

	return &Cache{
		store: rcache,
	}, nil
}

func (l *Cache) Add(key string, val *CacheEntry) bool {
	l.Lock()
	defer l.Unlock()
	if val.Lister != nil && val.data != nil {
		return false
	}
	return l.store.Set(key, val, 0)
}

func (l *Cache) Get(key string) (*CacheEntry, bool) {
	l.Lock()
	defer l.Unlock()
	val, ok := l.store.Get(key)
	if !ok {
		return nil, ok
	}

	entry, ok := val.(*CacheEntry)
	return entry, ok
}

func (l *Cache) Put(key string, val *CacheEntry) bool {
	l.Lock()
	defer l.Unlock()
	if val.Lister != nil && val.data != nil {
		return false
	}
	return l.store.Set(key, val, 0)
}

func (l *Cache) Delete(key string) bool {
	l.Lock()
	defer l.Unlock()

	l.store.Del(key)
	_, ok := l.store.Get(key)
	return ok
}

func ristrettoOnExit(val interface{}) {
	if entry, ok := val.(*CacheEntry); ok {
		if entry.Lister != nil {
			entry.stop()
		}
	}
}

func ristrettoOnEvict(item *ristretto.Item) {
	if entry, ok := item.Value.(*CacheEntry); ok {
		entry.stop()
	}
}
