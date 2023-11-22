package externalapi

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type Poller interface {
	Run(stopCh <-chan struct{})
}

type Getter interface {
	Get() interface{}
}

type externalAPIInformer struct {
	sync.Mutex
	ticker *time.Ticker
	client *http.Client
	req    *http.Request
	data   interface{}
}

type ExternalAPIInformer interface {
	Poller() Poller
	Getter() Getter
}

func NewExternalAPIInformer(ctx context.Context, url, caBundle string, interval time.Duration) (ExternalAPIInformer, error) {
	client, err := buildHTTPClient(url, caBundle)
	if err != nil {
		return nil, err
	}

	req, err := buildHTTPRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(interval)
	e := &externalAPIInformer{
		ticker: ticker,
		client: client,
		req:    req,
	}

	data, err := e.executeHTTPRequest()
	if err != nil {
		return nil, err
	}
	e.data = data

	return e, nil
}

func (e *externalAPIInformer) Getter() Getter {
	return e
}

func (e *externalAPIInformer) Poller() Poller {
	return e
}

func (e *externalAPIInformer) Get() interface{} {
	e.Lock()
	defer e.Unlock()
	return e.data
}

func (e *externalAPIInformer) Run(stopCh <-chan struct{}) {
	go func() {
		for {
			select {
			case <-e.ticker.C:
				data, err := e.executeHTTPRequest()
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					return
				}

				e.Lock()
				e.data = data
				e.Unlock()
				fmt.Fprintln(os.Stdout, "fetched from source url")
			case <-stopCh:
				fmt.Fprintln(os.Stdout, "exited")
				return
			}
		}
	}()
}

func (e *externalAPIInformer) executeHTTPRequest() (interface{}, error) {
	var data interface{}

	resp, err := e.client.Do(e.req)
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

func buildHTTPClient(url, caBundle string) (*http.Client, error) {
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

func buildHTTPRequest(ctx context.Context, url string) (req *http.Request, err error) {
	// TODO: Add auth token

	return http.NewRequestWithContext(ctx, "GET", url, nil)
}
