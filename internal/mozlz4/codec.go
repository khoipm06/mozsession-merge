package mozlz4

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/pierrec/lz4/v4"
)

// MagicHeader is "mozLz40\0" (8 bytes: 109, 111, 122, 76, 122, 52, 48, 0)
var MagicHeader = []byte{'m', 'o', 'z', 'L', 'z', '4', '0', 0}

const (
	HeaderSize    = 8
	SizeBufSize   = 4
	PayloadOffset = HeaderSize + SizeBufSize // 12 bytes
)

var (
	ErrInvalidHeader = errors.New("mozlz4: invalid magic header")
	ErrFileTooShort  = errors.New("mozlz4: file too short")
	ErrDecompress    = errors.New("mozlz4: decompression failed")
	ErrCompress      = errors.New("mozlz4: compression failed")
)

// Decode decompresses a mozLz4-compressed byte slice.
// Format matches the dejsonlz4 reference implementation:
// 1. Checks minimum size (8 bytes magic + 4 bytes uncompressed size = 12 bytes).
// 2. Validates magic bytes "mozLz40\0".
// 3. Reads 4-byte little-endian uncompressed size.
// 4. Decompresses LZ4 block from offset 12 into allocated output buffer.
func Decode(data []byte) ([]byte, error) {
	if len(data) < PayloadOffset {
		return nil, ErrFileTooShort
	}

	for i := 0; i < HeaderSize; i++ {
		if data[i] != MagicHeader[i] {
			return nil, ErrInvalidHeader
		}
	}

	uncompressedSize := binary.LittleEndian.Uint32(data[HeaderSize:PayloadOffset])
	if uncompressedSize == 0 {
		return []byte{}, nil
	}

	out := make([]byte, uncompressedSize)
	n, err := lz4.UncompressBlock(data[PayloadOffset:], out)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecompress, err)
	}
	if uint32(n) != uncompressedSize {
		out = out[:n]
	}

	return out, nil
}

// Encode compresses uncompressed data into mozLz4 format.
// Format matches the jsonlz4 reference implementation:
// 1. Writes magic bytes "mozLz40\0" (8 bytes).
// 2. Writes uncompressed length in 4 bytes little-endian.
// 3. Compresses uncompressed data into LZ4 block starting at offset 12.
func Encode(data []byte) ([]byte, error) {
	uncompressedSize := len(data)
	if uncompressedSize == 0 {
		out := make([]byte, PayloadOffset)
		copy(out[:HeaderSize], MagicHeader)
		binary.LittleEndian.PutUint32(out[HeaderSize:PayloadOffset], 0)
		return out, nil
	}

	var compressor lz4.Compressor
	maxCompressLen := lz4.CompressBlockBound(uncompressedSize)
	out := make([]byte, PayloadOffset+maxCompressLen)

	copy(out[:HeaderSize], MagicHeader)
	binary.LittleEndian.PutUint32(out[HeaderSize:PayloadOffset], uint32(uncompressedSize))

	csize, err := compressor.CompressBlock(data, out[PayloadOffset:])
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCompress, err)
	}
	if csize == 0 {
		return nil, ErrCompress
	}

	return out[:PayloadOffset+csize], nil
}
