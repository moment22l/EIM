package endian

import "io"

// WriteBytes 写一个 buf []byte 到 writer 中
func WriteBytes(w io.Writer, buf []byte) error {
	bufLen := len(buf)

	if err := WriteUnit32(w, unit32(bufLen)); err != nil {
		return err
	}
	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
}
