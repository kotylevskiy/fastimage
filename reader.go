package fastimage

import (
	"io"
)

// GetInfoReader reads from r until it can determine the image info or EOF.
func GetInfoReader(r io.Reader) (Info, error) {
	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 4096)

	for {
		n, err := r.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
			info := GetInfo(buf)
			if info.Type != Unknown && info.Width != 0 && info.Height != 0 {
				return info, nil
			}
		}
		if err != nil {
			if err == io.EOF {
				return GetInfo(buf), nil
			}
			return Info{}, err
		}
	}
}
