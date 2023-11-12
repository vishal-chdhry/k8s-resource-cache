package resourcecache

import (
	"fmt"
	"os"

	"github.com/dgraph-io/ristretto"
	"k8s.io/client-go/tools/cache"
)

type ListerCacheEntry struct {
	Lister         cache.GenericNamespaceLister
	informerStopCh *chan struct{}
}

type ListerCache struct {
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
	return l.store.Set(key, val, 1)
}
func (l *ListerCache) Get(key string) (*ListerCacheEntry, bool) {
	val, ok := l.store.Get(key)
	if !ok {
		return nil, ok
	}

	fmt.Fprintln(os.Stdout, "found in cache")
	entry, ok := val.(*ListerCacheEntry)
	return entry, ok
}

func ristrettoOnExit(val interface{}) {
	if entry, ok := val.(ListerCacheEntry); ok {
		*entry.informerStopCh <- struct{}{}
	}
}
