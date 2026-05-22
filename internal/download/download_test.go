package download

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestPreDownload_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("file-content-here"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	dest := filepath.Join(dir, "abc--test--1.0.bottle.tar.gz")
	pkgs := []PackageURL{{Name: "test", Version: "1.0", URL: srv.URL + "/test.tar.gz", CachePath: dest}}

	results := PreDownload(pkgs, Options{Concurrency: 4})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}
	if results[0].Cached {
		t.Fatal("expected Cached=false")
	}
	if results[0].Bytes != 17 {
		t.Fatalf("expected 17 bytes, got %d", results[0].Bytes)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("file not found: %v", err)
	}
	if string(data) != "file-content-here" {
		t.Fatalf("unexpected content: %s", data)
	}
}

func TestPreDownload_AlreadyCached(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "abc--cached--1.0.bottle.tar.gz")
	os.WriteFile(dest, []byte("cached"), 0644)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("server should not be called for cached file")
	}))
	defer srv.Close()

	pkgs := []PackageURL{{Name: "cached", Version: "1.0", URL: srv.URL + "/x", CachePath: dest}}
	results := PreDownload(pkgs, Options{Concurrency: 4})

	if !results[0].Cached {
		t.Fatal("expected Cached=true")
	}
	if results[0].Bytes != 6 {
		t.Fatalf("expected 6 bytes, got %d", results[0].Bytes)
	}
}

func TestPreDownload_FailureCleansUp(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	dir := t.TempDir()
	dest := filepath.Join(dir, "abc--bad--1.0.bottle.tar.gz")
	pkgs := []PackageURL{{Name: "bad", Version: "1.0", URL: srv.URL + "/missing", CachePath: dest}}

	results := PreDownload(pkgs, Options{Concurrency: 4})

	if results[0].Err == nil {
		t.Fatal("expected error for 404")
	}
	if _, err := os.Stat(dest + ".downloading"); !os.IsNotExist(err) {
		t.Fatal(".downloading file should be cleaned up")
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatal("final file should not exist on failure")
	}
}

func TestPreDownload_RedirectFollowed(t *testing.T) {
	final := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("redirected-content"))
	}))
	defer final.Close()

	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, final.URL+"/file", http.StatusFound)
	}))
	defer redirect.Close()

	dir := t.TempDir()
	dest := filepath.Join(dir, "abc--redir--2.0.bottle.tar.gz")
	pkgs := []PackageURL{{Name: "redir", Version: "2.0", URL: redirect.URL + "/start", CachePath: dest}}

	results := PreDownload(pkgs, Options{Concurrency: 4})

	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}
	data, _ := os.ReadFile(dest)
	if string(data) != "redirected-content" {
		t.Fatalf("unexpected content: %s", data)
	}
}

func TestPreDownload_ConcurrencyLimit(t *testing.T) {
	var active int64
	var maxActive int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := atomic.AddInt64(&active, 1)
		for {
			old := atomic.LoadInt64(&maxActive)
			if cur <= old {
				break
			}
			if atomic.CompareAndSwapInt64(&maxActive, old, cur) {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt64(&active, -1)
		w.Write([]byte("x"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	var pkgs []PackageURL
	for i := 0; i < 8; i++ {
		pkgs = append(pkgs, PackageURL{
			Name:      "pkg",
			Version:   "1.0",
			URL:       srv.URL + "/" + string(rune('a'+i)),
			CachePath: filepath.Join(dir, "file-"+string(rune('a'+i))+".tar.gz"),
		})
	}

	results := PreDownload(pkgs, Options{Concurrency: 2})

	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("unexpected error: %v", r.Err)
		}
	}
	if atomic.LoadInt64(&maxActive) > 2 {
		t.Fatalf("concurrency exceeded: max active was %d, expected <= 2", maxActive)
	}
}

func TestPreDownload_MixedSuccessAndFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/fail" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	pkgs := []PackageURL{
		{Name: "good", Version: "1.0", URL: srv.URL + "/good", CachePath: filepath.Join(dir, "good.tar.gz")},
		{Name: "bad", Version: "1.0", URL: srv.URL + "/fail", CachePath: filepath.Join(dir, "bad.tar.gz")},
		{Name: "good2", Version: "1.0", URL: srv.URL + "/good2", CachePath: filepath.Join(dir, "good2.tar.gz")},
	}

	results := PreDownload(pkgs, Options{Concurrency: 4})

	if results[0].Err != nil {
		t.Fatalf("expected success for good: %v", results[0].Err)
	}
	if results[1].Err == nil {
		t.Fatal("expected error for bad")
	}
	if results[2].Err != nil {
		t.Fatalf("expected success for good2: %v", results[2].Err)
	}

	if _, err := os.Stat(filepath.Join(dir, "good.tar.gz")); err != nil {
		t.Fatal("good.tar.gz should exist")
	}
	if _, err := os.Stat(filepath.Join(dir, "bad.tar.gz")); !os.IsNotExist(err) {
		t.Fatal("bad.tar.gz should not exist")
	}
}

func TestPreDownload_ZeroConcurrency(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("data"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	dest := filepath.Join(dir, "test.tar.gz")
	pkgs := []PackageURL{{Name: "test", Version: "1.0", URL: srv.URL + "/x", CachePath: dest}}

	results := PreDownload(pkgs, Options{Concurrency: 0})

	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}
}
