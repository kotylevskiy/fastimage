# fastimage - fast image info for go

fastimage is a tiny Go helper that sniffs image headers to extract dimensions without
fully decoding the file, keeping it lightweight and suitable for hot paths like upload
validation, image proxies, or metadata indexing pipelines.

This is a fork of [github.com/rubenfonseca/fastimage](https://github.com/rubenfonseca/fastimage) with added AVIF support.

Big thanks and deep respect to [@rubenfonseca](https://github.com/rubenfonseca), author of the original repo.

## Features

* Zero Dependencies - core library has no imports
* High Performance - hand-written header parsing (no regex/wildcard)
* Widely Format - bmp/bpm/gif/jpeg/mng/pbm/pcx/pgm/png/ppm/psd/ras/tiff/webp/xbm/xpm/avif

### Getting Started

try on https://play.golang.org/p/8yHaCknD1Rm
```go
package main

import (
	"fmt"
	"github.com/phuslu/fastimage"
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

### Command Tool
```bash
$ go get github.com/phuslu/fastimage/cmd/fastimage
$ fastimage banner.png
png image/png 320 50
```
