package resourcecache

import (
	"errors"
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

type ResourceCache struct {
	sync.Mutex
	client *dynamic.DynamicClient
	cache  map[string]cache.GenericNamespaceLister
}

func GetCacheSelector() (labels.Selector, error) {
	selector := labels.Everything()
	requirement, err := labels.NewRequirement("cache.kyverno.io/enabled", selection.Exists, nil)
	if err != nil {
		return nil, err
	}
	return selector.Add(*requirement), err
}

func NewResourceCache(d *dynamic.DynamicClient) *ResourceCache {
	return &ResourceCache{
		client: d,
		cache:  make(map[string]cache.GenericNamespaceLister),
	}
}

func (rc *ResourceCache) GetLister(resource schema.GroupVersionResource, namespace string) (cache.GenericNamespaceLister, error) {
	key := rc.getKeyForEntry(resource, namespace)
	lister, ok := rc.cache[key]
	if ok {
		return lister, nil
	}

	lister, err := rc.createGenericListerForResource(resource, namespace)
	if err != nil {
		return nil, err
	}
	rc.Lock()
	defer rc.Unlock()

	rc.cache[key] = lister
	return lister, nil
}

func (rc *ResourceCache) createGenericListerForResource(resource schema.GroupVersionResource, namespace string) (cache.GenericNamespaceLister, error) {
	selector, err := GetCacheSelector()
	if err != nil {
		return nil, err
	}
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(rc.client, time.Minute, namespace, func(opts *metav1.ListOptions) {
		opts.LabelSelector = selector.String()
	})
	informer := factory.ForResource(resource)

	stopCh := make(chan struct{}, 1)
	go informer.Informer().Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, informer.Informer().HasSynced) {
		return nil, errors.New("resource informer cache failed to sync")
	}

	return informer.Lister().ByNamespace(namespace), nil
}

func (rc *ResourceCache) getKeyForEntry(resource schema.GroupVersionResource, namespace string) string {
	return strings.Join([]string{resource.String(), ", Namespace=", namespace}, "")
}
