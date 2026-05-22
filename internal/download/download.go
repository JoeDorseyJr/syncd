package download

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Result holds the outcome of a single download attempt.
type Result struct {
	Package PackageURL
	Err     error
	Bytes   int64
	Cached  bool
}

// Options configures the parallel download.
type Options struct {
	Concurrency int
	CacheDir    string
}

// PreDownload downloads all packages in parallel to brew's cache.
// All failures are per-package in Result.Err; the function never fails overall.
func PreDownload(packages []PackageURL, opts Options) []Result {
	concurrency := opts.Concurrency
	if concurrency < 1 {
		concurrency = 1
	}

	results := make([]Result, len(packages))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	client := &http.Client{Timeout: 5 * time.Minute}

	for i, pkg := range packages {
		results[i].Package = pkg
		dest := filepath.Join(opts.CacheDir, pkg.Filename)

		// Skip if already cached
		if info, err := os.Stat(dest); err == nil && info.Size() > 0 {
			results[i].Cached = true
			results[i].Bytes = info.Size()
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, p PackageURL, destPath string) {
			defer wg.Done()
			defer func() { <-sem }()

			bytes, err := downloadFile(client, p.URL, destPath)
			results[idx].Bytes = bytes
			results[idx].Err = err
		}(i, pkg, dest)
	}

	wg.Wait()
	return results
}

func downloadFile(client *http.Client, url, dest string) (int64, error) {
	tmp := dest + ".downloading"

	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, &httpError{StatusCode: resp.StatusCode, URL: url}
	}

	f, err := os.Create(tmp)
	if err != nil {
		return 0, err
	}

	n, err := io.Copy(f, resp.Body)
	f.Close()
	if err != nil {
		os.Remove(tmp)
		return 0, err
	}

	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return 0, err
	}
	return n, nil
}

type httpError struct {
	StatusCode int
	URL        string
}

func (e *httpError) Error() string {
	return "HTTP " + http.StatusText(e.StatusCode) + " for " + e.URL
}
