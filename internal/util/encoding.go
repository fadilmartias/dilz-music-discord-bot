package util

import (
	"bufio"
	"encoding/binary"
)

func ReadPCM16(r *bufio.Reader, buffer []int16) error {
	for i := range buffer {
		b1, err := r.ReadByte()
		if err != nil {
			return err
		}
		b2, err := r.ReadByte()
		if err != nil {
			return err
		}
		buffer[i] = int16(binary.LittleEndian.Uint16([]byte{b1, b2}))
	}
	return nil
}
