package iso8583

import (
	"encoding/hex"
	"fmt"
)

const (
	primaryBitmapBytes   = 8
	secondaryBitmapBytes = 8
)

// Bitmap represents the primary (fields 1-64) and optional secondary (fields 65-128) bitmap.
type Bitmap struct {
	bits [128]bool
}

func NewBitmap() *Bitmap {
	return &Bitmap{}
}

func (b *Bitmap) Set(field int) {
	if field >= 1 && field <= 128 {
		b.bits[field-1] = true
	}
}

func (b *Bitmap) IsSet(field int) bool {
	if field >= 1 && field <= 128 {
		return b.bits[field-1]
	}
	return false
}

func (b *Bitmap) HasSecondary() bool {
	return b.bits[0]
}

func (b *Bitmap) Fields() []int {
	var fields []int
	limit := 64
	if b.HasSecondary() {
		limit = 128
	}
	for i := 2; i <= limit; i++ {
		if b.bits[i-1] {
			fields = append(fields, i)
		}
	}
	return fields
}

func DecodeBitmap(data []byte) (*Bitmap, int, error) {
	if len(data) < primaryBitmapBytes {
		return nil, 0, fmt.Errorf("bitmap: need at least %d bytes, got %d", primaryBitmapBytes, len(data))
	}

	bm := NewBitmap()
	bytesRead := primaryBitmapBytes

	for i := 0; i < primaryBitmapBytes; i++ {
		for bit := 7; bit >= 0; bit-- {
			fieldNum := i*8 + (7 - bit) + 1
			if data[i]&(1<<uint(bit)) != 0 {
				bm.bits[fieldNum-1] = true
			}
		}
	}

	if bm.HasSecondary() {
		if len(data) < primaryBitmapBytes+secondaryBitmapBytes {
			return nil, 0, fmt.Errorf("bitmap: secondary indicated but only %d bytes available", len(data))
		}
		for i := 0; i < secondaryBitmapBytes; i++ {
			b := data[primaryBitmapBytes+i]
			for bit := 7; bit >= 0; bit-- {
				fieldNum := 64 + i*8 + (7 - bit) + 1
				if b&(1<<uint(bit)) != 0 {
					bm.bits[fieldNum-1] = true
				}
			}
		}
		bytesRead += secondaryBitmapBytes
	}

	return bm, bytesRead, nil
}

func DecodeBitmapHex(hexStr string) (*Bitmap, error) {
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("bitmap: invalid hex: %w", err)
	}
	bm, _, err := DecodeBitmap(data)
	return bm, err
}

func (b *Bitmap) Encode() []byte {
	hasSecondary := false
	for i := 64; i < 128; i++ {
		if b.bits[i] {
			hasSecondary = true
			break
		}
	}

	if hasSecondary {
		b.bits[0] = true
	}

	size := primaryBitmapBytes
	if hasSecondary {
		size += secondaryBitmapBytes
	}

	result := make([]byte, size)
	for i := 0; i < size; i++ {
		var byteVal byte
		for bit := 7; bit >= 0; bit-- {
			fieldNum := i*8 + (7 - bit)
			if fieldNum < len(b.bits) && b.bits[fieldNum] {
				byteVal |= 1 << uint(bit)
			}
		}
		result[i] = byteVal
	}

	return result
}
