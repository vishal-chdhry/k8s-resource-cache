package resourcecache

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

type ResourceCache interface {
	GetLister(resource schema.GroupVersionResource, namespace string) (cache.GenericNamespaceLister, error)
}

type resourceCache struct {
	sync.Mutex
	client *dynamic.DynamicClient
	cache  *ListerCache
}

func GetCacheSelector() (labels.Selector, error) {
	selector := labels.Everything()
	requirement, err := labels.NewRequirement("cache.kyverno.io/enabled", selection.Exists, nil)
	if err != nil {
		return nil, err
	}
	return selector.Add(*requirement), err
}

func NewResourceCache(d *dynamic.DynamicClient) (ResourceCache, error) {
	lcache, err := NewListerCache()
	if err != nil {
		return nil, err
	}
	return &resourceCache{
		client: d,
		cache:  lcache,
	}, nil
}

func (rc *resourceCache) GetLister(resource schema.GroupVersionResource, namespace string) (cache.GenericNamespaceLister, error) {
	key := rc.getKeyForEntry(resource, namespace)
	lister, ok := rc.cache.Get(key)
	if ok {
		return lister.Lister, nil
	}

	listerEntry, err := rc.createGenericListerForResource(resource, namespace)
	if err != nil {
		return nil, err
	}
	rc.Lock()
	defer rc.Unlock()

	ok = rc.cache.Add(key, listerEntry)
	if !ok {
		return nil, fmt.Errorf("failed to create a cache entry key=%s", key)
	}
	return listerEntry.Lister, nil
}

func (rc *resourceCache) createGenericListerForResource(resource schema.GroupVersionResource, namespace string) (*ListerCacheEntry, error) {
	selector, err := GetCacheSelector()
	if err != nil {
		return nil, err
	}
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(rc.client, 5*time.Second, namespace, func(opts *metav1.ListOptions) {
		opts.LabelSelector = selector.String()
	})
	informer := factory.ForResource(resource)

	stopCh := make(chan struct{}, 1)
	go informer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, informer.Informer().HasSynced) {
		return nil, errors.New("resource informer cache failed to sync")
	}

	return &ListerCacheEntry{
		Lister:         informer.Lister().ByNamespace(namespace),
		informerStopCh: &stopCh,
	}, nil
}

func (rc *resourceCache) getKeyForEntry(resource schema.GroupVersionResource, namespace string) string {
	return strings.Join([]string{resource.String(), ", Namespace=", namespace}, "")
}
