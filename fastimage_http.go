package fastimage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// CONCURRENT_REQUESTS_FOR_REUSABLE_CONNECTIONS_DEFAULT is the default per-origin concurrency when range requests are supported.
const (
	CONCURRENT_REQUESTS_FOR_REUSABLE_CONNECTIONS_DEFAULT = 20
	// CONCURRENT_REQUESTS_FOR_NON_REUSABLE_CONNECTIONS_DEFAULT is the default per-origin concurrency when range requests are not supported.
	CONCURRENT_REQUESTS_FOR_NON_REUSABLE_CONNECTIONS_DEFAULT = 5
	// MAX_CONCURRENT_CONNECTIONS_GLOBAL_DEFAULT is the default global concurrency cap across origins.
	MAX_CONCURRENT_CONNECTIONS_GLOBAL_DEFAULT = 50
)

// HTTPImageInfo holds the URL and detected image metadata.
type HTTPImageInfo struct {
	// URL is the original image URL.
	URL string `json:"url"`
	Info
}

// GetHTTPImageResult contains image metadata or an error for a given URL.
type GetHTTPImageResult struct {
	HTTPImageInfo
	Error error `json:"error,omitempty"`
}

// GetHTTPImageOptions controls concurrency behavior for HTTP image probing.
type GetHTTPImageOptions struct {
	// ConcurrentRequestsReusable is the per-origin limit when range requests are supported.
	ConcurrentRequestsReusable int
	// ConcurrentRequestsNonReusable is the per-origin limit when range requests are not supported.
	ConcurrentRequestsNonReusable int
	// MaxConcurrentConnections is the global limit across all origins.
	MaxConcurrentConnections int
}

// GetHTTPImageInfo fetches basic image metadata for a list of URLs using default options.
//
// Errors:
//   - context.Canceled or context.DeadlineExceeded if the context ends.
//   - *url.Error from url.Parse or for invalid URLs.
//   - http.Client transport errors from http.Client.Do.
//   - io.ReadAll errors while reading the response body.
//   - *HTTPStatusError for non-200/206 responses.
//   - *RetryAfterError for 429/503 responses with parseable Retry-After.
//   - *InsufficientBytesError when there is not enough data to detect image info.
func GetHTTPImageInfo(ctx context.Context, urls []string) []GetHTTPImageResult {
	return GetHTTPImageDataWithOptions(ctx, urls, GetHTTPImageOptions{})
}

// GetHTTPImageDataWithOptions fetches basic image metadata for a list of URLs using custom options.
//
// Errors:
//   - context.Canceled or context.DeadlineExceeded if the context ends.
//   - *url.Error from url.Parse or for invalid URLs.
//   - http.Client transport errors from http.Client.Do.
//   - io.ReadAll errors while reading the response body.
//   - *HTTPStatusError for non-200/206 responses.
//   - *RetryAfterError for 429/503 responses with parseable Retry-After.
//   - *InsufficientBytesError when there is not enough data to detect image info.
func GetHTTPImageDataWithOptions(ctx context.Context, urls []string, options GetHTTPImageOptions) []GetHTTPImageResult {
	sizes := []int64{1024, 4096, 16384, 65536, 262144}

	if ctx == nil {
		ctx = context.Background()
	}

	options = normalizeHTTPImageOptions(options)

	results := make([]GetHTTPImageResult, len(urls))
	if len(urls) == 0 {
		return results
	}

	type item struct {
		index  int
		rawURL string
	}

	originGroups := make(map[string][]item)
	origins := make([]string, 0)

	for i, rawURL := range urls {
		results[i].URL = rawURL
		parsed, err := url.Parse(rawURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			if err == nil {
				err = &url.Error{Op: "parse", URL: rawURL, Err: fmt.Errorf("invalid URL")}
			}
			results[i].Error = err
			continue
		}
		host := normalizeOriginHost(parsed)
		origin := parsed.Scheme + "://" + host
		if _, ok := originGroups[origin]; !ok {
			origins = append(origins, origin)
		}
		originGroups[origin] = append(originGroups[origin], item{index: i, rawURL: rawURL})
	}

	sort.Strings(origins)

	type originWorker struct {
		client  *http.Client
		limiter *originLimiter
	}
	originWorkers := make(map[string]originWorker, len(origins))
	for _, origin := range origins {
		transport := &http.Transport{
			ForceAttemptHTTP2: true,
			MaxConnsPerHost:   options.ConcurrentRequestsReusable,
			Proxy:             http.ProxyFromEnvironment,
		}
		originWorkers[origin] = originWorker{
			client:  &http.Client{Transport: transport},
			limiter: newOriginLimiter(options.ConcurrentRequestsNonReusable, options.ConcurrentRequestsReusable),
		}
	}

	globalLimiter := make(chan struct{}, options.MaxConcurrentConnections)
	var wg sync.WaitGroup

	for _, origin := range origins {
		worker := originWorkers[origin]
		for _, it := range originGroups[origin] {
			if results[it.index].Error != nil {
				continue
			}
			wg.Add(1)
			go func(it item, worker originWorker) {
				defer wg.Done()
				info, err := fetchImageInfo(ctx, worker.client, it.rawURL, globalLimiter, worker.limiter, sizes)
				if err != nil {
					results[it.index].Error = err
					return
				}
				results[it.index].HTTPImageInfo = HTTPImageInfo{
					URL:  it.rawURL,
					Info: info,
				}
			}(it, worker)
		}
	}
	wg.Wait()

	for _, worker := range originWorkers {
		worker.client.CloseIdleConnections()
	}

	return results
}

func normalizeHTTPImageOptions(options GetHTTPImageOptions) GetHTTPImageOptions {
	if options.ConcurrentRequestsReusable < 1 {
		options.ConcurrentRequestsReusable = CONCURRENT_REQUESTS_FOR_REUSABLE_CONNECTIONS_DEFAULT
	}
	if options.ConcurrentRequestsNonReusable < 1 {
		options.ConcurrentRequestsNonReusable = CONCURRENT_REQUESTS_FOR_NON_REUSABLE_CONNECTIONS_DEFAULT
	}
	if options.MaxConcurrentConnections < 1 {
		options.MaxConcurrentConnections = MAX_CONCURRENT_CONNECTIONS_GLOBAL_DEFAULT
	}
	if options.ConcurrentRequestsReusable < options.ConcurrentRequestsNonReusable {
		options.ConcurrentRequestsReusable = options.ConcurrentRequestsNonReusable
	}
	return options
}

func acquire(ctx context.Context, limiter chan struct{}) error {
	select {
	case limiter <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func release(limiter chan struct{}) {
	<-limiter
}

func fetchImageInfo(
	ctx context.Context,
	client *http.Client,
	rawURL string,
	globalLimiter chan struct{},
	originLimiter *originLimiter,
	sizes []int64,
) (Info, error) {
	var info Info
	if err := acquire(ctx, globalLimiter); err != nil {
		return info, err
	}
	releaseOrigin, err := originLimiter.acquire(ctx)
	if err != nil {
		release(globalLimiter)
		return info, err
	}
	defer releaseOrigin()
	defer release(globalLimiter)

	return fetchImageInfoWithRetry(ctx, client, rawURL, sizes, originLimiter)
}

func fetchImageInfoWithRetry(
	ctx context.Context,
	client *http.Client,
	rawURL string,
	sizes []int64,
	originLimiter *originLimiter,
) (Info, error) {
	var info Info
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		var retryAfter time.Duration
		info, retryAfter, lastErr = fetchImageInfoProgressive(ctx, client, rawURL, sizes, originLimiter)
		if lastErr == nil {
			return info, nil
		}
		if retryAfter <= 0 || attempt == 1 {
			break
		}
		if err := sleepWithContext(ctx, retryAfter); err != nil {
			return info, err
		}
	}
	return info, lastErr
}

func fetchImageInfoProgressive(
	ctx context.Context,
	client *http.Client,
	rawURL string,
	sizes []int64,
	originLimiter *originLimiter,
) (Info, time.Duration, error) {
	var info Info
	var lastErr error
	lastRead := 0

	for _, size := range sizes {
		if size < 80 {
			continue
		}
		var retryAfter time.Duration
		var needMore bool
		var readBytes int
		info, retryAfter, lastErr, needMore, readBytes = fetchImageInfoOnce(ctx, client, rawURL, size, originLimiter)
		if readBytes > 0 {
			lastRead = readBytes
		}
		if retryAfter > 0 {
			return info, retryAfter, lastErr
		}
		if lastErr != nil {
			return info, 0, lastErr
		}
		if !needMore {
			return info, 0, nil
		}
	}

	if lastErr == nil {
		// We tried all sizes but still couldn't detect enough header/dimensions.
		// Treat as insufficient bytes for detection.
		lastErr = &InsufficientBytesError{Got: lastRead, Min: 80}
	}
	return info, 0, lastErr
}

func fetchImageInfoOnce(
	ctx context.Context,
	client *http.Client,
	rawURL string,
	minBytes int64,
	originLimiter *originLimiter,
) (Info, time.Duration, error, bool, int) {
	var info Info

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return info, 0, err, false, 0
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=0-%d", minBytes-1))

	resp, err := client.Do(req)
	if err != nil {
		return info, 0, err, false, 0
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
		if retryAfter, ok := parseRetryAfter(resp.Header.Get("Retry-After")); ok {
			return info, retryAfter, &RetryAfterError{
				URL:        rawURL,
				StatusCode: resp.StatusCode,
				Status:     resp.Status,
				RetryAfter: retryAfter,
			}, false, 0
		}
	}

	if resp.StatusCode == http.StatusPartialContent {
		originLimiter.enableReusable()
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return info, 0, &HTTPStatusError{
			URL:        rawURL,
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
		}, false, 0
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, minBytes))
	if err != nil {
		return info, 0, err, false, 0
	}

	readBytes := len(data)
	if readBytes < 80 {
		return info, 0, &InsufficientBytesError{Got: readBytes, Min: 80}, false, readBytes
	}

	info = GetInfo(data)
	if info.Type == Unknown || info.Width == 0 || info.Height == 0 {
		return info, 0, nil, true, readBytes
	}

	return info, 0, nil, false, readBytes
}

func parseRetryAfter(value string) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds <= 0 {
			return 0, false
		}
		return time.Duration(seconds) * time.Second, true
	}
	if retryAt, err := http.ParseTime(value); err == nil {
		wait := time.Until(retryAt)
		if wait <= 0 {
			return 0, false
		}
		return wait, true
	}
	return 0, false
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func normalizeOriginHost(u *url.URL) string {
	if u == nil {
		return ""
	}
	host := u.Host
	port := u.Port()
	if port == "" {
		return host
	}
	if (u.Scheme == "https" && port == "443") || (u.Scheme == "http" && port == "80") {
		if h := u.Hostname(); h != "" {
			return h
		}
	}
	return host
}

type originLimiter struct {
	base           chan struct{}
	extra          chan struct{}
	rangeSupported atomic.Bool
}

func newOriginLimiter(nonReusableLimit int, reusableLimit int) *originLimiter {
	if nonReusableLimit < 1 {
		nonReusableLimit = 1
	}
	if reusableLimit < nonReusableLimit {
		reusableLimit = nonReusableLimit
	}
	extra := reusableLimit - nonReusableLimit
	var extraChan chan struct{}
	if extra > 0 {
		extraChan = make(chan struct{}, extra)
	}
	return &originLimiter{
		base:  make(chan struct{}, nonReusableLimit),
		extra: extraChan,
	}
}

func (l *originLimiter) enableReusable() {
	l.rangeSupported.Store(true)
}

func (l *originLimiter) acquire(ctx context.Context) (func(), error) {
	if l.rangeSupported.Load() && l.extra != nil {
		select {
		case l.base <- struct{}{}:
			return func() { <-l.base }, nil
		case l.extra <- struct{}{}:
			return func() { <-l.extra }, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	select {
	case l.base <- struct{}{}:
		return func() { <-l.base }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
