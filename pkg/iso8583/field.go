package iso8583

import (
	"fmt"
	"strconv"
)

type Encoding int

const (
	ASCII Encoding = iota
	BCD
	Binary
)

type LengthType int

const (
	Fixed LengthType = iota
	LLVAR
	LLLVAR
)

type FieldSpec struct {
	Number    int
	Name      string
	Length    int
	LenType  LengthType
	Encoding Encoding
}

type FieldEncoder interface {
	Encode(value string) ([]byte, error)
	Decode(data []byte) (string, int, error)
}

func NewFieldEncoder(spec FieldSpec) FieldEncoder {
	switch spec.LenType {
	case LLVAR:
		return &varLengthEncoder{spec: spec, prefixDigits: 2}
	case LLLVAR:
		return &varLengthEncoder{spec: spec, prefixDigits: 3}
	default:
		return &fixedLengthEncoder{spec: spec}
	}
}

type fixedLengthEncoder struct {
	spec FieldSpec
}

func (e *fixedLengthEncoder) Encode(value string) ([]byte, error) {
	if e.spec.Encoding == Binary {
		if len(value) != e.spec.Length {
			return nil, fmt.Errorf("field %d: binary value length %d != spec length %d", e.spec.Number, len(value), e.spec.Length)
		}
		return []byte(value), nil
	}

	if len(value) > e.spec.Length {
		return nil, fmt.Errorf("field %d: value length %d exceeds max %d", e.spec.Number, len(value), e.spec.Length)
	}

	padded := fmt.Sprintf("%-*s", e.spec.Length, value)
	if isNumericField(e.spec) {
		padded = fmt.Sprintf("%0*s", e.spec.Length, value)
	}
	return []byte(padded), nil
}

func (e *fixedLengthEncoder) Decode(data []byte) (string, int, error) {
	if len(data) < e.spec.Length {
		return "", 0, fmt.Errorf("field %d: need %d bytes, got %d", e.spec.Number, e.spec.Length, len(data))
	}
	return string(data[:e.spec.Length]), e.spec.Length, nil
}

type varLengthEncoder struct {
	spec         FieldSpec
	prefixDigits int
}

func (e *varLengthEncoder) Encode(value string) ([]byte, error) {
	if len(value) > e.spec.Length {
		return nil, fmt.Errorf("field %d: value length %d exceeds max %d", e.spec.Number, len(value), e.spec.Length)
	}

	prefix := fmt.Sprintf("%0*d", e.prefixDigits, len(value))
	return []byte(prefix + value), nil
}

func (e *varLengthEncoder) Decode(data []byte) (string, int, error) {
	if len(data) < e.prefixDigits {
		return "", 0, fmt.Errorf("field %d: need at least %d bytes for length prefix", e.spec.Number, e.prefixDigits)
	}

	lenStr := string(data[:e.prefixDigits])
	fieldLen, err := strconv.Atoi(lenStr)
	if err != nil {
		return "", 0, fmt.Errorf("field %d: invalid length prefix %q: %w", e.spec.Number, lenStr, err)
	}

	if fieldLen > e.spec.Length {
		return "", 0, fmt.Errorf("field %d: declared length %d exceeds max %d", e.spec.Number, fieldLen, e.spec.Length)
	}

	totalLen := e.prefixDigits + fieldLen
	if len(data) < totalLen {
		return "", 0, fmt.Errorf("field %d: need %d bytes, got %d", e.spec.Number, totalLen, len(data))
	}

	return string(data[e.prefixDigits:totalLen]), totalLen, nil
}

func isNumericField(spec FieldSpec) bool {
	switch spec.Number {
	case 3, 4, 7, 11, 12, 13, 14, 22, 23, 25, 37, 38, 39, 49, 70:
		return true
	}
	return false
}
