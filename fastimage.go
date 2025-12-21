package fastimage

// Type represents the type of the image detected, or `Unknown`.
type Type uint64

const (
	// Unknown represents an unknown image type
	Unknown Type = iota
	// BMP represendts a BMP image
	BMP
	// BPM represendts a BPM image
	BPM
	// GIF represendts a GIF image
	GIF
	// JPEG represendts a JPEG image
	JPEG
	// MNG represendts a MNG image
	MNG
	// PBM represendts a PBM image
	PBM
	// PCX represendts a PCX image
	PCX
	// PGM represendts a PGM image
	PGM
	// PNG represendts a PNG image
	PNG
	// PPM represendts a PPM image
	PPM
	// PSD represendts a PSD image
	PSD
	// RAS represendts a RAS image
	RAS
	// RGB represendts a RGB image
	RGB
	// TIFF represendts a TIFF image
	TIFF
	// WEBP represendts a WEBP image
	WEBP
	// XBM represendts a XBM image
	XBM
	// XPM represendts a XPM image
	XPM
	// XV represendts a XV image
	XV
	// AVIF represendts a AVIF image
	AVIF
)

// String return a lower name of image type
func (t Type) String() string {
	switch t {
	case BMP:
		return "bmp"
	case BPM:
		return "bpm"
	case GIF:
		return "gif"
	case JPEG:
		return "jpeg"
	case MNG:
		return "mng"
	case PBM:
		return "pbm"
	case PCX:
		return "pcx"
	case PGM:
		return "pgm"
	case PNG:
		return "png"
	case PPM:
		return "ppm"
	case PSD:
		return "psd"
	case RAS:
		return "ras"
	case RGB:
		return "rgb"
	case TIFF:
		return "tiff"
	case WEBP:
		return "webp"
	case XBM:
		return "xbm"
	case XPM:
		return "xpm"
	case XV:
		return "xv"
	case AVIF:
		return "avif"
	}
	return ""
}

// Mime return mime type of image type
func (t Type) Mime() string {
	switch t {
	case BMP:
		return "image/bmp"
	case BPM:
		return "image/x-portable-pixmap"
	case GIF:
		return "image/gif"
	case JPEG:
		return "image/jpeg"
	case MNG:
		return "video/x-mng"
	case PBM:
		return "image/x-portable-bitmap"
	case PCX:
		return "image/x-pcx"
	case PGM:
		return "image/x-portable-graymap"
	case PNG:
		return "image/png"
	case PPM:
		return "image/x-portable-pixmap"
	case PSD:
		return "image/vnd.adobe.photoshop"
	case RAS:
		return "image/x-cmu-raster"
	case RGB:
		return "image/x-rgb"
	case TIFF:
		return "image/tiff"
	case WEBP:
		return "image/webp"
	case XBM:
		return "image/x-xbitmap"
	case XPM:
		return "image/x-xpixmap"
	case XV:
		return "image/x-portable-pixmap"
	case AVIF:
		return "image/avif"
	}
	return ""
}

// Info holds the type and dismissons of an image
type Info struct {
	Type   Type   `json:"type"`
	Width  uint32 `json:"width"`
	Height uint32 `json:"height"`
}

// GetType detects a image info of data (minimum 80 bytes required).
func GetType(p []byte) Type {
	const minOffset = 80 // 1 pixel gif
	if len(p) < minOffset {
		return Unknown
	}
	_ = p[minOffset-1]

	switch {
	case hasJPEG(p):
		return JPEG
	case hasPNG(p):
		return PNG
	case hasWEBP(p):
		return WEBP
	case hasGIF(p):
		return GIF
	case hasBMP(p):
		return BMP
	case hasPPM(p):
		return PPM
	case hasXBM(p):
		return XBM
	case hasXPM(p):
		return XPM
	case hasTIFFBig(p):
		return TIFF
	case hasTIFFLittle(p):
		return TIFF
	case hasPSD(p):
		return PSD
	case hasMNG(p):
		return MNG
	case hasRGB(p):
		return RGB
	case hasRAS(p):
		return RAS
	case hasPCX(p):
		return PCX
	case hasAVIFFtyp(p):
		return AVIF
	}

	return Unknown
}

// GetInfo detects a image info of data (minimum 80 bytes required).
func GetInfo(p []byte) (info Info) {
	const minOffset = 80 // 1 pixel gif
	if len(p) < minOffset {
		return
	}
	_ = p[minOffset-1]

	switch {
	case hasJPEG(p):
		jpeg(p, &info)
	case hasPNG(p):
		png(p, &info)
	case hasWEBP(p):
		webp(p, &info)
	case hasGIF(p):
		gif(p, &info)
	case hasBMP(p):
		bmp(p, &info)
	case hasPPM(p):
		ppm(p, &info)
	case hasXBM(p):
		xbm(p, &info)
	case hasXPM(p):
		xpm(p, &info)
	case hasTIFFBig(p):
		tiff(p, &info, bigEndian)
	case hasTIFFLittle(p):
		tiff(p, &info, littleEndian)
	case hasPSD(p):
		psd(p, &info)
	case hasMNG(p):
		mng(p, &info)
	case hasRGB(p):
		rgb(p, &info)
	case hasRAS(p):
		ras(p, &info)
	case hasPCX(p):
		pcx(p, &info)
	case hasAVIFFtyp(p):
		avif(p, &info)
	}

	return
}

func hasJPEG(b []byte) bool {
	return len(b) >= 2 && b[0] == '\xff' && b[1] == '\xd8'
}

func hasPNG(b []byte) bool {
	return len(b) >= 8 &&
		b[0] == '\x89' &&
		b[1] == 'P' &&
		b[2] == 'N' &&
		b[3] == 'G' &&
		b[4] == '\x0d' &&
		b[5] == '\x0a' &&
		b[6] == '\x1a' &&
		b[7] == '\x0a'
}

func hasWEBP(b []byte) bool {
	return len(b) >= 12 &&
		b[0] == 'R' &&
		b[1] == 'I' &&
		b[2] == 'F' &&
		b[3] == 'F' &&
		b[8] == 'W' &&
		b[9] == 'E' &&
		b[10] == 'B' &&
		b[11] == 'P'
}

func hasGIF(b []byte) bool {
	return len(b) >= 6 &&
		b[0] == 'G' &&
		b[1] == 'I' &&
		b[2] == 'F' &&
		b[3] == '8' &&
		(b[4] == '7' || b[4] == ',' || b[4] == '9') &&
		b[5] == 'a'
}

func hasBMP(b []byte) bool {
	return len(b) >= 2 && b[0] == 'B' && b[1] == 'M'
}

func hasPPM(b []byte) bool {
	if len(b) < 2 || b[0] != 'P' {
		return false
	}
	switch b[1] {
	case '1', '2', '3', '4', '5', '6', '7':
		return true
	}
	return false
}

func hasXBM(b []byte) bool {
	return len(b) >= 8 &&
		b[0] == '#' &&
		b[1] == 'd' &&
		b[2] == 'e' &&
		b[3] == 'f' &&
		b[4] == 'i' &&
		b[5] == 'n' &&
		b[6] == 'e' &&
		(b[7] == ' ' || b[7] == '\t')
}

func hasXPM(b []byte) bool {
	return len(b) >= 9 &&
		b[0] == '/' &&
		b[1] == '*' &&
		b[2] == ' ' &&
		b[3] == 'X' &&
		b[4] == 'P' &&
		b[5] == 'M' &&
		b[6] == ' ' &&
		b[7] == '*' &&
		b[8] == '/'
}

func hasTIFFBig(b []byte) bool {
	return len(b) >= 4 && b[0] == 'M' && b[1] == 'M' && b[2] == '\x00' && b[3] == '\x2a'
}

func hasTIFFLittle(b []byte) bool {
	return len(b) >= 4 && b[0] == 'I' && b[1] == 'I' && b[2] == '\x2a' && b[3] == '\x00'
}

func hasPSD(b []byte) bool {
	return len(b) >= 4 && b[0] == '8' && b[1] == 'B' && b[2] == 'P' && b[3] == 'S'
}

func hasMNG(b []byte) bool {
	return len(b) >= 8 &&
		b[0] == '\x8a' &&
		b[1] == 'M' &&
		b[2] == 'N' &&
		b[3] == 'G' &&
		b[4] == '\x0d' &&
		b[5] == '\x0a' &&
		b[6] == '\x1a' &&
		b[7] == '\x0a'
}

func hasRGB(b []byte) bool {
	return len(b) >= 6 &&
		b[0] == '\x01' &&
		b[1] == '\xda' &&
		b[2] == '[' &&
		b[3] == '\x01' &&
		b[4] == '\x00' &&
		b[5] == ']'
}

func hasRAS(b []byte) bool {
	return len(b) >= 4 && b[0] == '\x59' && b[1] == '\xa6' && b[2] == '\x6a' && b[3] == '\x95'
}

func hasPCX(b []byte) bool {
	return len(b) >= 3 && b[0] == '\x0a' && b[2] == '\x01'
}

func hasAVIFFtyp(b []byte) bool {
	for i := 0; i+8 <= len(b); {
		size32 := bigEndian.Uint32(b[i : i+4])
		size := int(size32)
		header := 8
		switch size32 {
		case 1:
			if i+16 > len(b) {
				return false
			}
			size64 := readUint64(b[i+8 : i+16])
			if size64 < 16 || size64 > uint64(len(b)-i) {
				return false
			}
			size = int(size64)
			header = 16
		case 0:
			size = len(b) - i
		}
		if size < header {
			return false
		}
		if i+size > len(b) {
			return false
		}
		if b[i+4] == 'f' &&
			b[i+5] == 't' &&
			b[i+6] == 'y' &&
			b[i+7] == 'p' {
			return ftypHasAVIF(b[i:i+size], header)
		}
		i += size
	}
	return false
}

func ftypHasAVIF(b []byte, header int) bool {
	if len(b) < header+8 {
		return false
	}
	if isAVIFBrand(b[header : header+4]) {
		return true
	}
	for i := header + 8; i+4 <= len(b); i += 4 {
		if isAVIFBrand(b[i : i+4]) {
			return true
		}
	}
	return false
}

func isAVIFBrand(b []byte) bool {
	return len(b) >= 4 &&
		b[0] == 'a' &&
		b[1] == 'v' &&
		b[2] == 'i' &&
		(b[3] == 'f' || b[3] == 's')
}

func jpeg(b []byte, info *Info) {
	i := 2
	for {
		length := int(b[i+3]) | int(b[i+2])<<8
		code := b[i+1]
		marker := b[i]
		i += 4
		switch {
		case marker != 0xff:
			return
		case code >= 0xc0 && code <= 0xc3:
			info.Type = JPEG
			info.Width = uint32(b[i+4]) | uint32(b[i+3])<<8
			info.Height = uint32(b[i+2]) | uint32(b[i+1])<<8
			return
		default:
			i += int(length) - 2
		}
	}
}

func webp(b []byte, info *Info) {
	if len(b) < 30 {
		return
	}
	_ = b[29]

	if !(b[12] == 'V' && b[13] == 'P' && b[14] == '8') {
		return
	}

	switch b[15] {
	case ' ': // VP8
		info.Width = (uint32(b[27])&0x3f)<<8 | uint32(b[26])
		info.Height = (uint32(b[29])&0x3f)<<8 | uint32(b[28])
	case 'L': // VP8L
		info.Width = (uint32(b[22])<<8|uint32(b[21]))&16383 + 1
		info.Height = (uint32(b[23])<<2|uint32(b[22]>>6))&16383 + 1
	case 'X': // VP8X
		info.Width = (uint32(b[24]) | uint32(b[25])<<8 | uint32(b[26])<<16) + 1
		info.Height = (uint32(b[27]) | uint32(b[28])<<8 | uint32(b[29])<<16) + 1
	}

	if info.Width != 0 && info.Height != 0 {
		info.Type = WEBP
	}
}

func avif(b []byte, info *Info) {
	info.Width, info.Height = avifDimensions(b)
	if info.Width != 0 && info.Height != 0 {
		info.Type = AVIF
	}
}

func avifDimensions(b []byte) (uint32, uint32) {
	for i := 4; i+16 <= len(b); i++ {
		if b[i] != 'i' ||
			b[i+1] != 's' ||
			b[i+2] != 'p' ||
			b[i+3] != 'e' {
			continue
		}
		size := int(bigEndian.Uint32(b[i-4 : i]))
		if size < 20 {
			continue
		}
		if i-4+size > len(b) {
			continue
		}
		width := bigEndian.Uint32(b[i+8 : i+12])
		height := bigEndian.Uint32(b[i+12 : i+16])
		if width != 0 && height != 0 {
			return width, height
		}
	}
	return 0, 0
}

func png(b []byte, info *Info) {
	if len(b) < 24 {
		return
	}
	_ = b[23]

	// IHDR
	if b[12] == 'I' && b[13] == 'H' && b[14] == 'D' && b[15] == 'R' {
		info.Width = uint32(b[16])<<24 |
			uint32(b[17])<<16 |
			uint32(b[18])<<8 |
			uint32(b[19])
		info.Height = uint32(b[20])<<24 |
			uint32(b[21])<<16 |
			uint32(b[22])<<8 |
			uint32(b[23])
	}

	if info.Width != 0 && info.Height != 0 {
		info.Type = PNG
	}
}

func gif(b []byte, info *Info) {
	if len(b) < 12 {
		return
	}
	_ = b[11]

	info.Width = uint32(b[7])<<8 | uint32(b[6])
	info.Height = uint32(b[9])<<8 | uint32(b[8])

	if info.Width != 0 && info.Height != 0 {
		info.Type = GIF
	}
}

func bmp(b []byte, info *Info) {
	if len(b) < 26 {
		return
	}
	_ = b[25]

	info.Width = uint32(b[21])<<24 |
		uint32(b[20])<<16 |
		uint32(b[19])<<8 |
		uint32(b[18])
	info.Height = uint32(b[25])<<24 |
		uint32(b[24])<<16 |
		uint32(b[23])<<8 |
		uint32(b[22])

	if info.Width != 0 && info.Height != 0 {
		info.Type = BMP
	}
}

func ppm(b []byte, info *Info) {
	switch b[1] {
	case '1':
		info.Type = PBM
	case '2', '5':
		info.Type = PGM
	case '3', '6':
		info.Type = PPM
	case '4':
		info.Type = BPM
	case '7':
		info.Type = XV
	}

	i := skipSpace(b, 2)
	info.Width, i = parseUint32(b, i)
	i = skipSpace(b, i)
	info.Height, _ = parseUint32(b, i)

	if info.Width == 0 || info.Height == 0 {
		info.Type = Unknown
	}
}

func xbm(b []byte, info *Info) {
	var p []byte
	var i int

	_, i = readNonSpace(b, i)
	i = skipSpace(b, i)
	_, i = readNonSpace(b, i)
	i = skipSpace(b, i)
	info.Width, i = parseUint32(b, i)

	i = skipSpace(b, i)
	p, i = readNonSpace(b, i)
	if !(len(p) == 7 &&
		p[6] == 'e' &&
		p[0] == '#' &&
		p[1] == 'd' &&
		p[2] == 'e' &&
		p[3] == 'f' &&
		p[4] == 'i' &&
		p[5] == 'n') {
		return
	}
	i = skipSpace(b, i)
	_, i = readNonSpace(b, i)
	i = skipSpace(b, i)
	info.Height, i = parseUint32(b, i)

	if info.Width != 0 && info.Height != 0 {
		info.Type = XBM
	}
}

func xpm(b []byte, info *Info) {
	var line []byte
	var i, j int

	for {
		line, i = readLine(b, i)
		if len(line) == 0 {
			break
		}
		j = skipSpace(line, 0)
		if line[j] != '"' {
			continue
		}
		info.Width, j = parseUint32(line, j+1)
		j = skipSpace(line, j)
		info.Height, j = parseUint32(line, j)
		break
	}

	if info.Width != 0 && info.Height != 0 {
		info.Type = XPM
	}
}

func tiff(b []byte, info *Info, order byteOrder) {
	i := int(order.Uint32(b[4:8]))
	n := int(order.Uint16(b[i+2 : i+4]))
	i += 2

	for ; i < n*12; i += 12 {
		tag := order.Uint16(b[i : i+2])
		datatype := order.Uint16(b[i+2 : i+4])

		var value uint32
		switch datatype {
		case 1, 6:
			value = uint32(b[i+9])
		case 3, 8:
			value = uint32(order.Uint16(b[i+8 : i+10]))
		case 4, 9:
			value = order.Uint32(b[i+8 : i+12])
		default:
			return
		}

		switch tag {
		case 256:
			info.Width = value
		case 257:
			info.Height = value
		}

		if info.Width > 0 && info.Height > 0 {
			info.Type = TIFF
			return
		}
	}
}

func psd(b []byte, info *Info) {
	if len(b) < 22 {
		return
	}
	_ = b[21]

	info.Width = uint32(b[18])<<24 |
		uint32(b[19])<<16 |
		uint32(b[20])<<8 |
		uint32(b[21])
	info.Height = uint32(b[14])<<24 |
		uint32(b[15])<<16 |
		uint32(b[16])<<8 |
		uint32(b[17])

	if info.Width != 0 && info.Height != 0 {
		info.Type = PSD
	}
}

func mng(b []byte, info *Info) {
	if len(b) < 24 {
		return
	}
	_ = b[23]

	if !(b[12] == 'M' && b[13] == 'H' && b[14] == 'D' && b[15] == 'R') {
		return
	}

	info.Width = uint32(b[16])<<24 |
		uint32(b[17])<<16 |
		uint32(b[18])<<8 |
		uint32(b[19])
	info.Height = uint32(b[20])<<24 |
		uint32(b[21])<<16 |
		uint32(b[22])<<8 |
		uint32(b[23])

	if info.Width != 0 && info.Height != 0 {
		info.Type = MNG
	}
}

func rgb(b []byte, info *Info) {
	if len(b) < 10 {
		return
	}
	_ = b[9]

	info.Width = uint32(b[6])<<8 |
		uint32(b[7])
	info.Height = uint32(b[8])<<8 |
		uint32(b[9])

	if info.Width != 0 && info.Height != 0 {
		info.Type = RGB
	}
}

func ras(b []byte, info *Info) {
	if len(b) < 12 {
		return
	}
	_ = b[11]

	info.Width = uint32(b[4])<<24 |
		uint32(b[5])<<16 |
		uint32(b[6])<<8 |
		uint32(b[7])
	info.Height = uint32(b[8])<<24 |
		uint32(b[9])<<16 |
		uint32(b[10])<<8 |
		uint32(b[11])

	if info.Width != 0 && info.Height != 0 {
		info.Type = RAS
	}
}

func pcx(b []byte, info *Info) {
	if len(b) < 12 {
		return
	}
	_ = b[11]

	info.Width = 1 +
		(uint32(b[9])<<8 | uint32(b[8])) -
		(uint32(b[5])<<8 | uint32(b[4]))
	info.Height = 1 +
		(uint32(b[11])<<8 | uint32(b[10])) -
		(uint32(b[7])<<8 | uint32(b[6]))

	if info.Width != 0 && info.Height != 0 {
		info.Type = PCX
	}
}

func skipSpace(b []byte, i int) (j int) {
	_ = b[len(b)-1]
	for j = i; j < len(b); j++ {
		if b[j] != ' ' && b[j] != '\t' && b[j] != '\r' && b[j] != '\n' {
			break
		}
	}
	return
}

func readNonSpace(b []byte, i int) (p []byte, j int) {
	_ = b[len(b)-1]
	for j = i; j < len(b); j++ {
		if b[j] == ' ' || b[j] == '\t' || b[j] == '\r' || b[j] == '\n' {
			break
		}
	}
	p = b[i:j]
	return
}

func readLine(b []byte, i int) (p []byte, j int) {
	_ = b[len(b)-1]
	for j = i; j < len(b); j++ {
		if b[j] == '\n' {
			break
		}
	}
	j++
	p = b[i:j]
	return
}

func parseUint32(b []byte, i int) (n uint32, j int) {
	_ = b[len(b)-1]
	for j = i; j < len(b); j++ {
		x := uint32(b[j] - '0')
		if x > 9 {
			break
		}
		n = n*10 + x
	}
	return
}

func readUint64(b []byte) uint64 {
	_ = b[7]
	return uint64(b[0])<<56 |
		uint64(b[1])<<48 |
		uint64(b[2])<<40 |
		uint64(b[3])<<32 |
		uint64(b[4])<<24 |
		uint64(b[5])<<16 |
		uint64(b[6])<<8 |
		uint64(b[7])
}

type byteOrder interface {
	Uint16([]byte) uint16
	Uint32([]byte) uint32
}

var littleEndian littleOrder

type littleOrder struct{}

func (littleOrder) Uint16(b []byte) uint16 {
	_ = b[1]
	return uint16(b[0]) | uint16(b[1])<<8
}

func (littleOrder) Uint32(b []byte) uint32 {
	_ = b[3]
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

var bigEndian bigOrder

type bigOrder struct{}

func (bigOrder) Uint16(b []byte) uint16 {
	_ = b[1]
	return uint16(b[1]) | uint16(b[0])<<8
}

func (bigOrder) Uint32(b []byte) uint32 {
	_ = b[3]
	return uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24
}
