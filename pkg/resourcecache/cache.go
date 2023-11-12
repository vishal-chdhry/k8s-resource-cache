package resourcecache

import (
	"context"
	"sync"

	"github.com/dgraph-io/ristretto"
	"k8s.io/client-go/tools/cache"
)

type ListerCacheEntry struct {
	Lister       cache.GenericNamespaceLister
	informerStop context.CancelFunc
}

type ListerCache struct {
	sync.Mutex
	store *ristretto.Cache
}

func NewListerCache() (*ListerCache, error) {
	config := ristretto.Config{
		MaxCost:     100 * 1000 * 1000, // 100 MB
		NumCounters: 10 * 100,          // 100 entries
		BufferItems: 64,
		OnExit:      ristrettoOnExit,
	}

	rcache, err := ristretto.NewCache(&config)
	if err != nil {
		return nil, err
	}

	return &ListerCache{
		store: rcache,
	}, nil
}

func (l *ListerCache) Add(key string, val *ListerCacheEntry) bool {
	l.Lock()
	defer l.Unlock()
	return l.store.Set(key, val, 0)
}
func (l *ListerCache) Get(key string) (*ListerCacheEntry, bool) {
	l.Lock()
	defer l.Unlock()
	val, ok := l.store.Get(key)
	if !ok {
		return nil, ok
	}

	entry, ok := val.(*ListerCacheEntry)
	return entry, ok
}

func ristrettoOnExit(val interface{}) {
	if entry, ok := val.(ListerCacheEntry); ok {
		entry.informerStop()
	}
}
