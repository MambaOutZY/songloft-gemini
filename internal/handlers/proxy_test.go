package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestServeRemoteResourceWithOptionsTimeout(t *testing.T) {
	release := make(chan struct{})
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	defer func() {
		close(release)
		upstream.Close()
	}()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/songs/1/cover", nil)
	rr := httptest.NewRecorder()

	start := time.Now()
	ServeRemoteResourceWithOptions(rr, req, upstream.URL, RemoteResourceOptions{
		Timeout:      20 * time.Millisecond,
		ErrorStatus:  http.StatusNotFound,
		ErrorMessage: "cover fetch failed",
	})

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
	if !strings.Contains(rr.Body.String(), "cover fetch failed") {
		t.Fatalf("body = %q, want cover fetch failed", rr.Body.String())
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("timeout response took %s, want under 1s", elapsed)
	}
}
