package iso8583

import (
	"testing"
)

func TestMessage_RoundTrip_0100(t *testing.T) {
	msg := NewMessage(DefaultSpec)
	msg.MTI = "0100"
	msg.SetField(2, "4532015112830366")
	msg.SetField(3, "000000")
	msg.SetField(4, "000000001000")
	msg.SetField(7, "0722143052")
	msg.SetField(11, "123456")
	msg.SetField(14, "2512")
	msg.SetField(22, "051")
	msg.SetField(25, "00")
	msg.SetField(37, "000000000001")
	msg.SetField(41, "TERM0001")
	msg.SetField(42, "MERCHANT0000001")
	msg.SetField(49, "840")

	encoded, err := msg.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	decoded, err := Decode(encoded, DefaultSpec)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if decoded.MTI != msg.MTI {
		t.Errorf("MTI = %q, want %q", decoded.MTI, msg.MTI)
	}

	for field, want := range msg.Fields {
		got, ok := decoded.GetField(field)
		if !ok {
			t.Errorf("field %d missing after decode", field)
			continue
		}

		spec := DefaultSpec[field]
		if spec.LenType == Fixed && !isNumericField(spec) {
			enc := NewFieldEncoder(spec)
			padded, _ := enc.Encode(want)
			want = string(padded)
		} else if spec.LenType == Fixed && isNumericField(spec) {
			enc := NewFieldEncoder(spec)
			padded, _ := enc.Encode(want)
			want = string(padded)
		}

		if got != want {
			t.Errorf("field %d = %q, want %q", field, got, want)
		}
	}
}

func TestMessage_RoundTrip_0110(t *testing.T) {
	msg := NewMessage(DefaultSpec)
	msg.MTI = "0110"
	msg.SetField(2, "4532015112830366")
	msg.SetField(3, "000000")
	msg.SetField(4, "000000001000")
	msg.SetField(7, "0722143052")
	msg.SetField(11, "123456")
	msg.SetField(37, "000000000001")
	msg.SetField(38, "ABC123")
	msg.SetField(39, "00")
	msg.SetField(41, "TERM0001")
	msg.SetField(49, "840")

	encoded, err := msg.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	decoded, err := Decode(encoded, DefaultSpec)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if decoded.MTI != "0110" {
		t.Errorf("MTI = %q, want 0110", decoded.MTI)
	}

	rc, ok := decoded.GetField(39)
	if !ok {
		t.Fatal("field 39 (response code) missing")
	}
	if rc != "00" {
		t.Errorf("response code = %q, want %q", rc, "00")
	}

	authID, ok := decoded.GetField(38)
	if !ok {
		t.Fatal("field 38 (auth ID) missing")
	}
	if authID != "ABC123" {
		t.Errorf("auth ID = %q, want %q", authID, "ABC123")
	}
}

func TestMessage_RoundTrip_0800(t *testing.T) {
	msg := NewMessage(DefaultSpec)
	msg.MTI = "0800"
	msg.SetField(7, "0722143052")
	msg.SetField(11, "000001")
	msg.SetField(70, "301")

	encoded, err := msg.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	decoded, err := Decode(encoded, DefaultSpec)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if decoded.MTI != "0800" {
		t.Errorf("MTI = %q, want 0800", decoded.MTI)
	}

	code, ok := decoded.GetField(70)
	if !ok {
		t.Fatal("field 70 missing")
	}
	if code != "301" {
		t.Errorf("network info code = %q, want %q", code, "301")
	}
}

func TestMessage_RoundTrip_WithLLVAR(t *testing.T) {
	msg := NewMessage(DefaultSpec)
	msg.MTI = "0100"
	msg.SetField(2, "4532015112830366")
	msg.SetField(3, "000000")
	msg.SetField(4, "000000005000")
	msg.SetField(11, "654321")
	msg.SetField(32, "12345678")
	msg.SetField(37, "000000000099")
	msg.SetField(41, "TERM0002")
	msg.SetField(49, "840")

	encoded, err := msg.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	decoded, err := Decode(encoded, DefaultSpec)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	pan, _ := decoded.GetField(2)
	if pan != "4532015112830366" {
		t.Errorf("PAN = %q, want 4532015112830366", pan)
	}

	acq, _ := decoded.GetField(32)
	if acq != "12345678" {
		t.Errorf("acquiring ID = %q, want 12345678", acq)
	}
}

func TestMessage_RoundTrip_WithLLLVAR(t *testing.T) {
	msg := NewMessage(DefaultSpec)
	msg.MTI = "0100"
	msg.SetField(3, "000000")
	msg.SetField(4, "000000002500")
	msg.SetField(11, "111111")
	msg.SetField(37, "000000000042")
	msg.SetField(41, "TERM0003")
	msg.SetField(49, "840")
	msg.SetField(54, "1001840C000000010000")

	encoded, err := msg.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	decoded, err := Decode(encoded, DefaultSpec)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	amt, _ := decoded.GetField(54)
	if amt != "1001840C000000010000" {
		t.Errorf("additional amounts = %q, want 1001840C000000010000", amt)
	}
}

func TestMessage_EncodeError_NoSpec(t *testing.T) {
	msg := NewMessage(DefaultSpec)
	msg.MTI = "0100"
	msg.SetField(999, "test")

	_, err := msg.Encode()
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestMessage_EncodeError_BadMTI(t *testing.T) {
	msg := NewMessage(DefaultSpec)
	msg.MTI = "01"
	msg.SetField(3, "000000")

	_, err := msg.Encode()
	if err == nil {
		t.Fatal("expected error for short MTI")
	}
}

func TestDecode_TooShort(t *testing.T) {
	_, err := Decode([]byte("01"), DefaultSpec)
	if err == nil {
		t.Fatal("expected error for data shorter than MTI")
	}
}

func TestMessage_FieldCount(t *testing.T) {
	msg := NewMessage(DefaultSpec)
	msg.MTI = "0100"
	msg.SetField(2, "4532015112830366")
	msg.SetField(3, "000000")
	msg.SetField(11, "123456")

	encoded, err := msg.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	decoded, err := Decode(encoded, DefaultSpec)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if len(decoded.Fields) != 3 {
		t.Errorf("field count = %d, want 3", len(decoded.Fields))
	}
}
