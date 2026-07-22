package iso8583

import (
	"encoding/hex"
	"testing"
)

func TestBitmap_SetAndIsSet(t *testing.T) {
	bm := NewBitmap()
	bm.Set(2)
	bm.Set(3)
	bm.Set(11)
	bm.Set(64)

	tests := []struct {
		field int
		want  bool
	}{
		{1, false},
		{2, true},
		{3, true},
		{4, false},
		{11, true},
		{64, true},
		{65, false},
	}

	for _, tt := range tests {
		if got := bm.IsSet(tt.field); got != tt.want {
			t.Errorf("IsSet(%d) = %v, want %v", tt.field, got, tt.want)
		}
	}
}

func TestBitmap_Fields(t *testing.T) {
	bm := NewBitmap()
	bm.Set(2)
	bm.Set(3)
	bm.Set(4)
	bm.Set(11)

	fields := bm.Fields()
	want := []int{2, 3, 4, 11}
	if len(fields) != len(want) {
		t.Fatalf("Fields() returned %d fields, want %d", len(fields), len(want))
	}
	for i, f := range fields {
		if f != want[i] {
			t.Errorf("Fields()[%d] = %d, want %d", i, f, want[i])
		}
	}
}

func TestBitmap_EncodeDecodePrimary(t *testing.T) {
	bm := NewBitmap()
	bm.Set(2)
	bm.Set(3)
	bm.Set(4)
	bm.Set(11)
	bm.Set(14)
	bm.Set(22)
	bm.Set(25)
	bm.Set(37)
	bm.Set(41)
	bm.Set(42)
	bm.Set(43)
	bm.Set(49)

	encoded := bm.Encode()
	if len(encoded) != 8 {
		t.Fatalf("primary-only bitmap should be 8 bytes, got %d", len(encoded))
	}

	decoded, bytesRead, err := DecodeBitmap(encoded)
	if err != nil {
		t.Fatalf("DecodeBitmap: %v", err)
	}
	if bytesRead != 8 {
		t.Errorf("bytesRead = %d, want 8", bytesRead)
	}

	for _, f := range []int{2, 3, 4, 11, 14, 22, 25, 37, 41, 42, 43, 49} {
		if !decoded.IsSet(f) {
			t.Errorf("field %d should be set after decode", f)
		}
	}
}

func TestBitmap_EncodeDecodeSecondary(t *testing.T) {
	bm := NewBitmap()
	bm.Set(2)
	bm.Set(70)

	encoded := bm.Encode()
	if len(encoded) != 16 {
		t.Fatalf("secondary bitmap should be 16 bytes, got %d", len(encoded))
	}

	if !bm.HasSecondary() {
		t.Error("HasSecondary() should be true after setting field 70")
	}

	decoded, bytesRead, err := DecodeBitmap(encoded)
	if err != nil {
		t.Fatalf("DecodeBitmap: %v", err)
	}
	if bytesRead != 16 {
		t.Errorf("bytesRead = %d, want 16", bytesRead)
	}
	if !decoded.IsSet(2) {
		t.Error("field 2 should be set")
	}
	if !decoded.IsSet(70) {
		t.Error("field 70 should be set")
	}
}

func TestDecodeBitmapHex(t *testing.T) {
	// Bitmap with fields 2, 3, 4, 7, 11, 12, 13, 14, 22, 25, 37, 41, 42, 43, 49
	hexStr := "723C048008E08000"
	bm, err := DecodeBitmapHex(hexStr)
	if err != nil {
		t.Fatalf("DecodeBitmapHex: %v", err)
	}

	raw, _ := hex.DecodeString(hexStr)
	t.Logf("bitmap hex: %s (bytes: %v)", hexStr, raw)

	expected := []int{2, 3, 4, 7, 11, 12, 13, 14, 22, 25, 37, 41, 42, 43, 49}
	for _, f := range expected {
		if !bm.IsSet(f) {
			t.Errorf("field %d should be set in bitmap %s", f, hexStr)
		}
	}
}

func TestBitmap_RoundTrip(t *testing.T) {
	fields := []int{2, 3, 4, 7, 11, 12, 13, 14, 22, 25, 32, 37, 41, 42, 43, 49}

	bm := NewBitmap()
	for _, f := range fields {
		bm.Set(f)
	}

	encoded := bm.Encode()
	decoded, _, err := DecodeBitmap(encoded)
	if err != nil {
		t.Fatalf("round trip decode: %v", err)
	}

	for _, f := range fields {
		if !decoded.IsSet(f) {
			t.Errorf("field %d lost in round trip", f)
		}
	}

	for i := 2; i <= 64; i++ {
		if decoded.IsSet(i) != bm.IsSet(i) {
			t.Errorf("field %d mismatch after round trip", i)
		}
	}
}
