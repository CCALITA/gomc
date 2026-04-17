package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// WriteString writes a UTF-8 string with a uint16 big-endian length prefix.
func WriteString(w io.Writer, s string) error {
	length := uint16(len(s))
	if err := binary.Write(w, binary.BigEndian, length); err != nil {
		return fmt.Errorf("failed to write string length: %w", err)
	}
	if _, err := w.Write([]byte(s)); err != nil {
		return fmt.Errorf("failed to write string data: %w", err)
	}
	return nil
}

// ReadString reads a UTF-8 string with a uint16 big-endian length prefix.
func ReadString(r io.Reader) (string, error) {
	var length uint16
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return "", fmt.Errorf("failed to read string length: %w", err)
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", fmt.Errorf("failed to read string data: %w", err)
	}
	return string(buf), nil
}

// WriteFloat32 writes a float32 in big-endian byte order.
func WriteFloat32(w io.Writer, v float32) error {
	return binary.Write(w, binary.BigEndian, math.Float32bits(v))
}

// ReadFloat32 reads a float32 in big-endian byte order.
func ReadFloat32(r io.Reader) (float32, error) {
	var bits uint32
	if err := binary.Read(r, binary.BigEndian, &bits); err != nil {
		return 0, err
	}
	return math.Float32frombits(bits), nil
}

// WriteFloat64 writes a float64 in big-endian byte order.
func WriteFloat64(w io.Writer, v float64) error {
	return binary.Write(w, binary.BigEndian, math.Float64bits(v))
}

// ReadFloat64 reads a float64 in big-endian byte order.
func ReadFloat64(r io.Reader) (float64, error) {
	var bits uint64
	if err := binary.Read(r, binary.BigEndian, &bits); err != nil {
		return 0, err
	}
	return math.Float64frombits(bits), nil
}

// WriteInt32 writes an int32 in big-endian byte order.
func WriteInt32(w io.Writer, v int32) error {
	return binary.Write(w, binary.BigEndian, v)
}

// ReadInt32 reads an int32 in big-endian byte order.
func ReadInt32(r io.Reader) (int32, error) {
	var v int32
	if err := binary.Read(r, binary.BigEndian, &v); err != nil {
		return 0, err
	}
	return v, nil
}

// WriteUint16 writes a uint16 in big-endian byte order.
func WriteUint16(w io.Writer, v uint16) error {
	return binary.Write(w, binary.BigEndian, v)
}

// ReadUint16 reads a uint16 in big-endian byte order.
func ReadUint16(r io.Reader) (uint16, error) {
	var v uint16
	if err := binary.Read(r, binary.BigEndian, &v); err != nil {
		return 0, err
	}
	return v, nil
}

// WriteUint32 writes a uint32 in big-endian byte order.
func WriteUint32(w io.Writer, v uint32) error {
	return binary.Write(w, binary.BigEndian, v)
}

// ReadUint32 reads a uint32 in big-endian byte order.
func ReadUint32(r io.Reader) (uint32, error) {
	var v uint32
	if err := binary.Read(r, binary.BigEndian, &v); err != nil {
		return 0, err
	}
	return v, nil
}

// WriteByte_ writes a single byte. Named with trailing underscore to avoid
// conflict with the io.ByteWriter interface method.
func WriteByte_(w io.Writer, v byte) error {
	return binary.Write(w, binary.BigEndian, v)
}

// ReadByte_ reads a single byte. Named with trailing underscore to avoid
// conflict with the io.ByteReader interface method.
func ReadByte_(r io.Reader) (byte, error) {
	var v byte
	if err := binary.Read(r, binary.BigEndian, &v); err != nil {
		return 0, err
	}
	return v, nil
}

// WriteBool writes a boolean as a single byte (1 = true, 0 = false).
func WriteBool(w io.Writer, v bool) error {
	var b byte
	if v {
		b = 1
	}
	return binary.Write(w, binary.BigEndian, b)
}

// ReadBool reads a boolean from a single byte (non-zero = true).
func ReadBool(r io.Reader) (bool, error) {
	var b byte
	if err := binary.Read(r, binary.BigEndian, &b); err != nil {
		return false, err
	}
	return b != 0, nil
}

// WriteInt64 writes an int64 in big-endian byte order.
func WriteInt64(w io.Writer, v int64) error {
	return binary.Write(w, binary.BigEndian, v)
}

// ReadInt64 reads an int64 in big-endian byte order.
func ReadInt64(r io.Reader) (int64, error) {
	var v int64
	if err := binary.Read(r, binary.BigEndian, &v); err != nil {
		return 0, err
	}
	return v, nil
}
