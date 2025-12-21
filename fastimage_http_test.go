package fastimage

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
)

func httpImageTestCases() []httpImageTestCase {
	return []httpImageTestCase{
		{Path: "/letter_T.jpg", File: "testdata/letter_T.jpg", Info: Info{JPEG, 52, 54}},
		{Path: "/4.sm.webp", File: "testdata/4.sm.webp", Info: Info{WEBP, 320, 241}},
		{Path: "/2_webp_a.webp", File: "testdata/2_webp_a.webp", Info: Info{WEBP, 386, 395}},
		{Path: "/2_webp_ll.webp", File: "testdata/2_webp_ll.webp", Info: Info{WEBP, 386, 395}},
		{Path: "/4_webp_ll.webp", File: "testdata/4_webp_ll.webp", Info: Info{WEBP, 421, 163}},
		{Path: "/pass-1_s.png", File: "testdata/pass-1_s.png", Info: Info{PNG, 90, 60}},
		{Path: "/pak38.gif", File: "testdata/pak38.gif", Info: Info{GIF, 333, 194}},
		{Path: "/test.gif", File: "testdata/test.gif", Info: Info{GIF, 60, 40}},
		{Path: "/xterm.bmp", File: "testdata/xterm.bmp", Info: Info{BMP, 64, 38}},
		{Path: "/letter_N.ppm", File: "testdata/letter_N.ppm", Info: Info{PPM, 66, 57}},
		{Path: "/spacer50.xbm", File: "testdata/spacer50.xbm", Info: Info{XBM, 50, 10}},
		{Path: "/xterm.xpm", File: "testdata/xterm.xpm", Info: Info{XPM, 64, 38}},
		{Path: "/bexjdic.tif", File: "testdata/bexjdic.tif", Info: Info{TIFF, 35, 32}},
		{Path: "/lexjdic.tif", File: "testdata/lexjdic.tif", Info: Info{TIFF, 35, 32}},
		{Path: "/letter_T.psd", File: "testdata/letter_T.psd", Info: Info{PSD, 52, 54}},
		{Path: "/468x60.psd", File: "testdata/468x60.psd", Info: Info{PSD, 468, 60}},
		{Path: "/letter_T.mng", File: "testdata/letter_T.mng", Info: Info{MNG, 52, 54}},
		{Path: "/letter_T.ras", File: "testdata/letter_T.ras", Info: Info{RAS, 52, 54}},
		{Path: "/letter_T.pcx", File: "testdata/letter_T.pcx", Info: Info{PCX, 52, 54}},
		{Path: "/bridge.avif", File: "testdata/bridge.avif", Info: Info{AVIF, 1000, 666}},
		{Path: "/cow.avif", File: "testdata/cow.avif", Info: Info{AVIF, 500, 300}},
		{Path: "/parrot.avif", File: "testdata/parrot.avif", Info: Info{AVIF, 1000, 667}},
	}
}

func TestGetHTTPImageDataWithRangeServer(t *testing.T) {
	server := newTestImageServer(t, true)
	defer server.Close()

	cases := httpImageTestCases()
	urls := make([]string, 0, len(cases))
	for _, c := range cases {
		urls = append(urls, server.URL+c.Path)
	}
	results := GetHTTPImageDataWithOptions(context.Background(), urls, GetHTTPImageOptions{
		ConcurrentRequestsReusable:    2,
		ConcurrentRequestsNonReusable: 1,
		MaxConcurrentConnections:      2,
	})

	if len(results) != len(urls) {
		t.Fatalf("unexpected results length: got %d want %d", len(results), len(urls))
	}
	for i, result := range results {
		if result.Error != nil {
			t.Fatalf("unexpected error for %s: %v", urls[i], result.Error)
		}
		if got, expected := result.Info, cases[i].Info; got != expected {
			t.Fatalf("unexpected info for %s: got %+v want %+v", urls[i], got, expected)
		}
	}
}

func TestGetHTTPImageDataWithoutRangeServer(t *testing.T) {
	server := newTestImageServer(t, false)
	defer server.Close()

	cases := httpImageTestCases()
	urls := make([]string, 0, len(cases))
	for _, c := range cases {
		urls = append(urls, server.URL+c.Path)
	}
	results := GetHTTPImageDataWithOptions(context.Background(), urls, GetHTTPImageOptions{
		ConcurrentRequestsReusable:    2,
		ConcurrentRequestsNonReusable: 1,
		MaxConcurrentConnections:      2,
	})

	if len(results) != len(urls) {
		t.Fatalf("unexpected results length: got %d want %d", len(results), len(urls))
	}
	for i, result := range results {
		if result.Error != nil {
			t.Fatalf("unexpected error for %s: %v", urls[i], result.Error)
		}
		if got, expected := result.Info, cases[i].Info; got != expected {
			t.Fatalf("unexpected info for %s: got %+v want %+v", urls[i], got, expected)
		}
	}
}

func newTestImageServer(t *testing.T, supportRange bool) *httptest.Server {
	t.Helper()

	files := make(map[string]string)
	for _, c := range httpImageTestCases() {
		files[c.Path] = c.File
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		filePath, ok := files[path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		data, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if supportRange {
			if start, end, ok := parseRangeHeader(r.Header.Get("Range"), len(data)); ok {
				w.Header().Set("Accept-Ranges", "bytes")
				w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(data)))
				w.WriteHeader(http.StatusPartialContent)
				_, _ = w.Write(data[start : end+1])
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})

	return httptest.NewServer(handler)
}

func parseRangeHeader(value string, size int) (int, int, bool) {
	if !strings.HasPrefix(value, "bytes=") {
		return 0, 0, false
	}
	parts := strings.SplitN(strings.TrimPrefix(value, "bytes="), "-", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return 0, 0, false
	}
	start, err := strconv.Atoi(parts[0])
	if err != nil || start < 0 {
		return 0, 0, false
	}
	end, err := strconv.Atoi(parts[1])
	if err != nil || end < start {
		return 0, 0, false
	}
	if start >= size {
		return 0, 0, false
	}
	if end >= size {
		end = size - 1
	}
	return start, end, true
}

type httpImageTestCase struct {
	Path string
	File string
	Info Info
}
