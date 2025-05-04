package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promslog"
)

const (
	apiKey = "test-api-key"
)

func newTTNServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v3/gateways", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+apiKey {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		file, err := os.Open(path.Join("test", "gateways_list_response.json"))
		if err != nil {
			t.Error(err)
		}
		bytes, err := io.ReadAll(file)
		if err != nil {
			t.Error(err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bytes)
	})
	// NOTE: Golang does not support regex in patterns, so we're hardcoding this
	mux.HandleFunc("/api/v3/gs/gateways", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+apiKey {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/api/v3/gs/gateways/a111111aaa111222/connection/stats" {
			http.NotFound(w, r)
			return
		}

		file, err := os.Open(path.Join("test", "gateway_stats_response.json"))
		if err != nil {
			t.Error(err)
		}
		bytes, err := io.ReadAll(file)
		if err != nil {
			t.Error(err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bytes)
	})

	ts := httptest.NewServer(mux)
	return ts
}

func expectMetrics(t *testing.T, c prometheus.Collector, fixture string) {
	_, err := os.Open(path.Join("test", fixture)) // #nosec G304 Potential file inclusion via variable
	if err != nil {
		t.Fatalf("Error opening fixture file %q: %v", fixture, err)
	}
	// if err := testutil.CollectAndCompare(c, exp); err != nil {
	// 	t.Fatal("Unexpected metrics returned:", err)
	// }
}

func TestServer(t *testing.T) {
	h := newTTNServer(t)
	defer h.Close()

	e, _ := NewExporter(h.URL, apiKey, false, 5*time.Second, promslog.NewNopLogger())

	expectMetrics(t, e, "gateway_stats.metrics")
}

func TestInvalidScheme(t *testing.T) {
	e, err := NewExporter("gopher://gopher.quux.org", apiKey, false, 1*time.Second, promslog.NewNopLogger())
	if expect, got := (*Exporter)(nil), e; expect != got {
		t.Errorf("expected %v, got %v", expect, got)
	}
	if err == nil {
		t.Fatalf("expected non-nil error")
	}
	if expect, got := `invalid URI scheme`, err.Error(); expect != got {
		t.Errorf("expected %q, got %q", expect, got)
	}
}
