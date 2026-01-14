# fastimage - fast image info for go
[![Go Reference](https://pkg.go.dev/badge/github.com/kotylevskiy/fastimage.svg)](https://pkg.go.dev/github.com/kotylevskiy/fastimage) 
[![Go Report Card](https://goreportcard.com/badge/github.com/kotylevskiy/fastimage)](https://goreportcard.com/report/github.com/kotylevskiy/fastimage)
[![License](https://img.shields.io/github/license/kotylevskiy/go-sitemap-fetcher)](LICENSE)

fastimage is a tiny Go helper that sniffs image headers to extract dimensions without
fully decoding the file, keeping it lightweight and suitable for hot paths like upload
validation, image proxies, or metadata indexing pipelines. It also includes HTTP
helpers for remote image probing via progressive range requests.

Typical use-cases:

* Validate user image uploads (type + size) without decoding the full image
* Implement lightweight image proxies / thumbnailers
* Crawl and index image metadata at scale
* Quickly check remote banner images / avatars in microservices


## Differences from upstream

This repo started as a fork of [rubenfonseca/fastimage](https://github.com/rubenfonseca/fastimage) and adds:

* AVIF support
* HTTP helpers for concurrent, range-based remote image probing
* Stream-aware `GetInfoReader` API for working with `io.Reader`

Big thanks and deep respect to [@rubenfonseca](https://github.com/rubenfonseca), author of the original repo.

## Features

* Zero Dependencies - stdlib only
* High Performance - hand-written header parsing (no regex/wildcard)
* Wide format support – BMP, GIF, JPEG, MNG, PBM, PCX, PGM, PNG, PPM, PSD, RAS, TIFF, WebP, XBM, XPM, AVIF
* HTTP range helpers – progressive range fetching for remote images
* Reader API - stream-aware `GetInfoReader` for files or network responses

## Install

Library (modules will fetch it automatically when you import):

```go
import "github.com/kotylevskiy/fastimage"
```
CLI:

```bash
go install github.com/kotylevskiy/fastimage/cmd/fastimage@latest
```

### Getting Started

```go
package main

import (
	"fmt"

	"github.com/kotylevskiy/fastimage"
)

var data = []byte("RIFF,-\x00\x00WEBPVP8X\n\x00\x00\x00" +
    "\x10\x00\x00\x00\x8f\x01\x00,\x01\x00VP8X\n\x00\x00\x00\x10\xb2" +
    "\x01\x00\x00WEB\x01\x00VP8X\n\x00\x00\x00\x10\xb2\x01\x00" +
    "\x00WEB\x01\x00VP8X\n\x00\x00\x00\x10\xb2\x01\x00\x00W" +
    "EB\x01\x00VP8X\n\x00\x00\x00\x10\xb2\x01\x00\x00WEB"")

func main() {
	fmt.Printf("%+v\n", fastimage.GetInfo(data))
}

// Output: {Type:webp Width:400 Height:301}
```

### Reader API
```go
resp, err := http.Get("https://example.com/image.jpg")
if err != nil {
    // handle error
}
defer resp.Body.Close()

info, err := fastimage.GetInfoReader(resp.Body)
if err != nil {
    // handle error
}
fmt.Printf("%+v\n", info)
```

### HTTP Range Helper
The HTTP helper is multithreaded and probes URLs concurrently (bounded by the concurrency options below).
```go
urls := []string{
    "https://example.com/a.jpg",
    "https://example.com/b.png",
}
results := fastimage.GetHTTPImageInfo(context.Background(), urls)
for _, result := range results {
    fmt.Printf("%s %+v err=%v\n", result.URL, result.Info, result.Error)
}
```

### HTTP Concurrency Defaults
`GetHTTPImageInfo` uses these defaults:
- `CONCURRENT_REQUESTS_FOR_REUSABLE_CONNECTIONS_DEFAULT = 20` (per-origin when range is supported)
- `CONCURRENT_REQUESTS_FOR_NON_REUSABLE_CONNECTIONS_DEFAULT = 5` (per-origin when range is not supported)
- `MAX_CONCURRENT_CONNECTIONS_GLOBAL_DEFAULT = 50` (global cap across origins)

These limits cap concurrent requests; the HTTP transport also sets `MaxConnsPerHost` to `ConcurrentRequestsReusable` to avoid excessive connections per origin.

“Reusable” means the server supports HTTP range requests (`206 Partial Content`), so the client can reuse keep‑alive connections efficiently while fetching multiple byte ranges. “Not reusable” means range is unsupported; each probe may need a full response, so concurrency is reduced to avoid overloading the origin and wasting bandwidth.

For “browser‑like” behavior, the defaults are reasonable. The code uses 20 when range is supported by the remote host (requests multiplexed to a fewer number of connections) and falls back to 5 when it’s not.

To override them, use `GetHTTPImageDataWithOptions`:

```go
options := fastimage.GetHTTPImageOptions{
    ConcurrentRequestsReusable:    8,
    ConcurrentRequestsNonReusable: 2,
    MaxConcurrentConnections:      16,
}
results := fastimage.GetHTTPImageDataWithOptions(context.Background(), urls, options)
```

### Command Tool
```bash
$ go get github.com/kotylevskiy/fastimage/cmd/fastimage
$ fastimage banner.png
png image/png 320 50
$ fastimage https://example.com/banner.png
png image/png 320 50
```
