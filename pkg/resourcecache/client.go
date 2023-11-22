package resourcecache

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	externalapi "github.com/vishal-chdhry/k8s-resource-cache/pkg/externalAPI"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

type ResourceCache interface {
	GetLister(resource schema.GroupVersionResource, namespace string) (cache.GenericNamespaceLister, error)
	DeleteResourceEntry(resource schema.GroupVersionResource, namespace string) error

	GetExternalData(url, caBundle string, interval int) (externalapi.Getter, error)
	DeleteExternalEntry(url, caBundle string, interval int) error

	// TODO: Performance testing when storing mutliple listers in the cache
}

type resourceCache struct {
	client *dynamic.DynamicClient
	cache  *Cache
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
	key := rc.getKeyForResourceEntry(resource, namespace)
	lister, ok := rc.cache.Get(key)
	if ok {
		return lister.Lister, nil
	}

	listerEntry, err := rc.createGenericListerForResource(resource, namespace)
	if err != nil {
		return nil, err
	}

	ok = rc.cache.Add(key, listerEntry)
	if !ok {
		return nil, fmt.Errorf("failed to create cache entry key=%s", key)
	}
	return listerEntry.Lister, nil
}

func (rc *resourceCache) DeleteResourceEntry(resource schema.GroupVersionResource, namespace string) error {
	key := rc.getKeyForResourceEntry(resource, namespace)
	ok := rc.cache.Delete(key)
	if !ok {
		return fmt.Errorf("failed to delete cache entry key=%s", key)
	}

	return nil
}

func (rc *resourceCache) DeleteExternalEntry(url, caBundle string, interval int) error {
	key := rc.getKeyForExternalEntry(url, caBundle, interval)
	entry, ok := rc.cache.Get(key)
	if !ok {
		return nil
	}
	entry.stop()

	ok = rc.cache.Delete(key)
	if !ok {
		return fmt.Errorf("failed to delete cache entry key=%s", key)
	}

	return nil
}

func (rc *resourceCache) createGenericListerForResource(resource schema.GroupVersionResource, namespace string) (*CacheEntry, error) {
	informer := dynamicinformer.NewFilteredDynamicInformer(rc.client, resource, namespace, 5*time.Second, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	go informer.Informer().Run(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), informer.Informer().HasSynced) {
		cancel()
		return nil, errors.New("resource informer cache failed to sync")
	}

	return &CacheEntry{
		Lister: informer.Lister().ByNamespace(namespace),
		stop:   cancel,
	}, nil
}

func (rc *resourceCache) GetExternalData(url, caBundle string, interval int) (externalapi.Getter, error) {
	key := rc.getKeyForExternalEntry(url, caBundle, interval)
	entry, ok := rc.cache.Get(key)
	if ok {
		return entry.ExternalGetter, nil
	}

	externalEntry, err := rc.createExternalDataGetter(url, caBundle, interval)
	if err != nil {
		return nil, err
	}

	ok = rc.cache.Add(key, externalEntry)
	if !ok {
		return nil, fmt.Errorf("failed to create cache entry key=%s", key)
	}
	return externalEntry.ExternalGetter, nil
}

func (rc *resourceCache) createExternalDataGetter(url, caBundle string, interval int) (*CacheEntry, error) {
	intervalSec := time.Duration(interval*int(time.Nanosecond)) * time.Second
	informer, err := externalapi.NewExternalAPIInformer(context.Background(), url, caBundle, intervalSec)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go informer.Poller().Run(ctx.Done())

	return &CacheEntry{
		ExternalGetter: informer.Getter(),
		stop:           cancel,
	}, nil
}

func (rc *resourceCache) executeHTTPRequest(client *http.Client, req *http.Request) (interface{}, error) {
	var data interface{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (rc *resourceCache) buildHTTPClient(url, caBundle string) (*http.Client, error) {
	if url == "" || caBundle == "" {
		return http.DefaultClient, nil
	}
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM([]byte(caBundle)); !ok {
		return nil, fmt.Errorf("failed to parse PEM CA bundle for APICall %s", url)
	}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:    caCertPool,
			MinVersion: tls.VersionTLS12,
		},
	}
	return &http.Client{
		Transport: transport,
	}, nil
}

func (rc *resourceCache) buildHTTPRequest(ctx context.Context, url string) (req *http.Request, err error) {
	// TODO: Add auth token

	return http.NewRequestWithContext(ctx, "GET", url, nil)
}

func (rc *resourceCache) getKeyForResourceEntry(resource schema.GroupVersionResource, namespace string) string {
	return strings.Join([]string{"Resource= ", resource.String(), ", Namespace=", namespace}, "")
}

func (rc *resourceCache) getKeyForExternalEntry(url, caBundle string, interval int) string {
	return strings.Join([]string{"External= ", url, ", Bundle=", caBundle, "Refresh= ", fmt.Sprint(interval)}, "")
}
