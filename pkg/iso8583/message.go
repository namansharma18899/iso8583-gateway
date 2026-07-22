package iso8583

import (
	"fmt"
	"sort"
)

const mtiLength = 4

type Message struct {
	MTI    string
	Fields map[int]string
	Spec   map[int]FieldSpec
}

func NewMessage(spec map[int]FieldSpec) *Message {
	return &Message{
		Fields: make(map[int]string),
		Spec:   spec,
	}
}

func (m *Message) SetField(field int, value string) {
	m.Fields[field] = value
}

func (m *Message) GetField(field int) (string, bool) {
	v, ok := m.Fields[field]
	return v, ok
}

func (m *Message) Encode() ([]byte, error) {
	if len(m.MTI) != mtiLength {
		return nil, fmt.Errorf("encode: MTI must be %d characters, got %d", mtiLength, len(m.MTI))
	}

	bm := NewBitmap()
	var fieldNums []int
	for num := range m.Fields {
		bm.Set(num)
		fieldNums = append(fieldNums, num)
	}
	sort.Ints(fieldNums)

	var result []byte
	result = append(result, []byte(m.MTI)...)
	result = append(result, bm.Encode()...)

	for _, num := range fieldNums {
		spec, ok := m.Spec[num]
		if !ok {
			return nil, fmt.Errorf("encode: no spec for field %d", num)
		}
		encoder := NewFieldEncoder(spec)
		encoded, err := encoder.Encode(m.Fields[num])
		if err != nil {
			return nil, fmt.Errorf("encode field %d: %w", num, err)
		}
		result = append(result, encoded...)
	}

	return result, nil
}

func Decode(data []byte, spec map[int]FieldSpec) (*Message, error) {
	if len(data) < mtiLength {
		return nil, fmt.Errorf("decode: message too short for MTI, need %d bytes got %d", mtiLength, len(data))
	}

	msg := NewMessage(spec)
	msg.MTI = string(data[:mtiLength])
	offset := mtiLength

	bm, bmLen, err := DecodeBitmap(data[offset:])
	if err != nil {
		return nil, fmt.Errorf("decode bitmap: %w", err)
	}
	offset += bmLen

	for _, fieldNum := range bm.Fields() {
		fieldSpec, ok := spec[fieldNum]
		if !ok {
			return nil, fmt.Errorf("decode: no spec for field %d", fieldNum)
		}

		encoder := NewFieldEncoder(fieldSpec)
		value, bytesRead, err := encoder.Decode(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("decode field %d: %w", fieldNum, err)
		}

		msg.Fields[fieldNum] = value
		offset += bytesRead
	}

	return msg, nil
}
