package tcp

import (
	"EIM"
	"EIM/wire/endian"
	"io"
)

func WriteFrame(w io.Writer, code EIM.OpCode, payload []byte) error {
	if err := endian.WriteUint8(w, unit8(code)); err != nil {
		return err
	}
	if err := endian.WriteBytes(w, payload); err != nil {
		return err
	}
	return nil
}
