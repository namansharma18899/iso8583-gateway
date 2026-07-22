package iso8583

import (
	"testing"
)

func TestFixedLengthEncoder(t *testing.T) {
	tests := []struct {
		name    string
		spec    FieldSpec
		input   string
		wantEnc string
		wantErr bool
	}{
		{
			name:    "numeric field zero-padded",
			spec:    FieldSpec{Number: 3, Name: "Processing Code", Length: 6, LenType: Fixed},
			input:   "000000",
			wantEnc: "000000",
		},
		{
			name:    "numeric field needs padding",
			spec:    FieldSpec{Number: 11, Name: "STAN", Length: 6, LenType: Fixed},
			input:   "123",
			wantEnc: "000123",
		},
		{
			name:    "alpha field right-padded",
			spec:    FieldSpec{Number: 41, Name: "Terminal ID", Length: 8, LenType: Fixed},
			input:   "TERM001",
			wantEnc: "TERM001 ",
		},
		{
			name:    "exact length",
			spec:    FieldSpec{Number: 39, Name: "Response Code", Length: 2, LenType: Fixed},
			input:   "00",
			wantEnc: "00",
		},
		{
			name:    "too long",
			spec:    FieldSpec{Number: 39, Name: "Response Code", Length: 2, LenType: Fixed},
			input:   "000",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := NewFieldEncoder(tt.spec)
			encoded, err := enc.Encode(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Encode: %v", err)
			}
			if string(encoded) != tt.wantEnc {
				t.Errorf("Encode(%q) = %q, want %q", tt.input, encoded, tt.wantEnc)
			}

			decoded, bytesRead, err := enc.Decode(encoded)
			if err != nil {
				t.Fatalf("Decode: %v", err)
			}
			if bytesRead != tt.spec.Length {
				t.Errorf("bytesRead = %d, want %d", bytesRead, tt.spec.Length)
			}
			if decoded != tt.wantEnc {
				t.Errorf("Decode = %q, want %q", decoded, tt.wantEnc)
			}
		})
	}
}

func TestLLVAREncoder(t *testing.T) {
	tests := []struct {
		name    string
		spec    FieldSpec
		input   string
		wantEnc string
		wantErr bool
	}{
		{
			name:    "PAN 16 digits",
			spec:    FieldSpec{Number: 2, Name: "PAN", Length: 19, LenType: LLVAR},
			input:   "4532015112830366",
			wantEnc: "164532015112830366",
		},
		{
			name:    "PAN 19 digits",
			spec:    FieldSpec{Number: 2, Name: "PAN", Length: 19, LenType: LLVAR},
			input:   "4532015112830366123",
			wantEnc: "194532015112830366123",
		},
		{
			name:    "acquiring institution",
			spec:    FieldSpec{Number: 32, Name: "Acquiring Institution ID", Length: 11, LenType: LLVAR},
			input:   "123456",
			wantEnc: "06123456",
		},
		{
			name:    "too long",
			spec:    FieldSpec{Number: 2, Name: "PAN", Length: 19, LenType: LLVAR},
			input:   "12345678901234567890",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := NewFieldEncoder(tt.spec)
			encoded, err := enc.Encode(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Encode: %v", err)
			}
			if string(encoded) != tt.wantEnc {
				t.Errorf("Encode(%q) = %q, want %q", tt.input, encoded, tt.wantEnc)
			}

			decoded, bytesRead, err := enc.Decode(encoded)
			if err != nil {
				t.Fatalf("Decode: %v", err)
			}
			if bytesRead != len(tt.wantEnc) {
				t.Errorf("bytesRead = %d, want %d", bytesRead, len(tt.wantEnc))
			}
			if decoded != tt.input {
				t.Errorf("Decode = %q, want %q", decoded, tt.input)
			}
		})
	}
}

func TestLLLVAREncoder(t *testing.T) {
	tests := []struct {
		name    string
		spec    FieldSpec
		input   string
		wantEnc string
	}{
		{
			name:    "additional amounts",
			spec:    FieldSpec{Number: 54, Name: "Additional Amounts", Length: 120, LenType: LLLVAR},
			input:   "1001840C000000010000",
			wantEnc: "0201001840C000000010000",
		},
		{
			name:    "short value",
			spec:    FieldSpec{Number: 55, Name: "ICC/EMV Data", Length: 999, LenType: LLLVAR},
			input:   "9F0206",
			wantEnc: "0069F0206",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := NewFieldEncoder(tt.spec)
			encoded, err := enc.Encode(tt.input)
			if err != nil {
				t.Fatalf("Encode: %v", err)
			}
			if string(encoded) != tt.wantEnc {
				t.Errorf("Encode(%q) = %q, want %q", tt.input, encoded, tt.wantEnc)
			}

			decoded, bytesRead, err := enc.Decode(encoded)
			if err != nil {
				t.Fatalf("Decode: %v", err)
			}
			if bytesRead != len(tt.wantEnc) {
				t.Errorf("bytesRead = %d, want %d", bytesRead, len(tt.wantEnc))
			}
			if decoded != tt.input {
				t.Errorf("Decode = %q, want %q", decoded, tt.input)
			}
		})
	}
}

func TestFixedLengthDecoder_TooShort(t *testing.T) {
	enc := NewFieldEncoder(FieldSpec{Number: 3, Length: 6, LenType: Fixed})
	_, _, err := enc.Decode([]byte("123"))
	if err == nil {
		t.Fatal("expected error for short data")
	}
}

func TestLLVARDecoder_InvalidPrefix(t *testing.T) {
	enc := NewFieldEncoder(FieldSpec{Number: 2, Length: 19, LenType: LLVAR})
	_, _, err := enc.Decode([]byte("XX1234"))
	if err == nil {
		t.Fatal("expected error for non-numeric length prefix")
	}
}
