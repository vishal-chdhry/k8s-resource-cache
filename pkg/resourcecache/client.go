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
	"os"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

type ResourceCache interface {
	GetLister(resource schema.GroupVersionResource, namespace string) (cache.GenericNamespaceLister, error)
	DeleteResourceEntry(resource schema.GroupVersionResource, namespace string) error
	GetExternalData(url, caBundle string, interval int) (interface{}, error)
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
	ok := rc.cache.Delete(key)
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

func (rc *resourceCache) GetExternalData(url, caBundle string, interval int) (interface{}, error) {
	key := rc.getKeyForExternalEntry(url, caBundle, interval)
	entry, ok := rc.cache.Get(key)
	if ok {
		return entry.data, nil
	}

	client, err := rc.buildHTTPClient(url, caBundle)
	if err != nil {
		return nil, err
	}

	intervalSec := time.Duration(interval*int(time.Nanosecond)) * time.Second
	ticker := time.NewTicker(intervalSec)

	ctx, cancel := context.WithCancel(context.Background())
	req, err := rc.buildHTTPRequest(ctx, url)
	if err != nil {
		cancel()
		return nil, err
	}

	cacheCtx, cacheCancel := context.WithCancel(context.Background())

	entryChan := make(chan *CacheEntry, 1)
	go func() {
		for {
			select {
			case <-ticker.C:
				data, err := rc.executeHTTPRequest(client, req)
				if err != nil {
					cancel()
					cacheCancel()
					fmt.Fprintln(os.Stderr, err.Error())
					os.Exit(1)
				}

				entry := &CacheEntry{
					data: data,
					stop: cacheCancel,
				}

				ok := rc.cache.Put(key, entry)
				if !ok {
					os.Exit(1)
				}
				fmt.Fprintln(os.Stdout, "Fetched data from external url")
				select {
				case entryChan <- entry:
				default:
					continue
				}

			case <-cacheCtx.Done():
				fmt.Fprintln(os.Stdout, "exited")
				return
			}
		}
	}()

	entry = <-entryChan
	return entry.data, nil
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
