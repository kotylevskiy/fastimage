package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/kotylevskiy/fastimage"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage: %s <file>\n", filepath.Base(os.Args[0]))
		return
	}

	name := os.Args[1]
	info, err := getInfo(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read error: %+v", err)
		os.Exit(1)
	}
	if info.Type == fastimage.Unknown {
		os.Exit(1)
	}

	fmt.Printf("%s %s %d %d\n", info.Type, info.Type.Mime(), info.Width, info.Height)
}

func getInfo(name string) (fastimage.Info, error) {
	if isHTTPURL(name) {
		resp, err := http.Get(name)
		if err != nil {
			return fastimage.Info{}, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fastimage.Info{}, fmt.Errorf("unexpected status %s", resp.Status)
		}
		return fastimage.GetInfoReader(resp.Body)
	}

	file, err := os.Open(name)
	if err != nil {
		return fastimage.Info{}, err
	}
	defer file.Close()

	return fastimage.GetInfoReader(file)
}

func isHTTPURL(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}
