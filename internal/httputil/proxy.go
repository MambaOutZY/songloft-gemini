package httputil

import (
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var globalProxy proxyConfig

type proxyConfig struct {
	mu       sync.RWMutex
	proxyURL *url.URL
}

func (pc *proxyConfig) set(rawURL string) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	if rawURL == "" {
		pc.proxyURL = nil
		return nil
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	pc.proxyURL = u
	return nil
}

func (pc *proxyConfig) get() string {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	if pc.proxyURL == nil {
		return ""
	}
	return pc.proxyURL.String()
}

func (pc *proxyConfig) proxyFunc(req *http.Request) (*url.URL, error) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	if pc.proxyURL == nil {
		return nil, nil
	}
	host := req.URL.Hostname()
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return nil, nil
	}
	return pc.proxyURL, nil
}

func newProxyTransport(responseHeaderTimeout time.Duration) *http.Transport {
	return &http.Transport{
		Proxy: globalProxy.proxyFunc,
		DialContext: (&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: responseHeaderTimeout,
	}
}

var sharedTransport = newProxyTransport(0)

var streamingTransport = newProxyTransport(15 * time.Second)

// SetGlobalProxy sets the global HTTP proxy used by all clients created via NewClient.
// Pass an empty string to clear the proxy (direct connection).
func SetGlobalProxy(rawURL string) error {
	if err := globalProxy.set(rawURL); err != nil {
		return err
	}
	sharedTransport.CloseIdleConnections()
	streamingTransport.CloseIdleConnections()
	return nil
}

// GetGlobalProxy returns the current global HTTP proxy URL, or "" if not set.
func GetGlobalProxy() string {
	return globalProxy.get()
}

// NewClient creates an http.Client that uses the global HTTP proxy.
// Requests to loopback addresses bypass the proxy automatically.
func NewClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: sharedTransport,
		Timeout:   timeout,
	}
}

// NewStreamingClient creates an http.Client for long-lived streaming responses.
// It has no whole-request timeout, but it does time out while waiting for
// response headers so dead radio/HLS endpoints do not hang forever.
func NewStreamingClient() *http.Client {
	return &http.Client{
		Transport: streamingTransport,
	}
}
