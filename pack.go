package pack

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
)

const (
	Bin8      = 0xC4
	Bin16     = 0xC5
	Bin32     = 0xC6
	Nil       = 0xC0
	False     = 0xC2
	True      = 0XC3
	Float32   = 0xCA
	Float64   = 0xCB
	UintFix   = 0x00
	Uint8     = 0xCC
	Uint16    = 0xCD
	Uint32    = 0xCE
	Uint64    = 0XCF
	IntFix    = 0xE0
	Int8      = 0xD0
	Int16     = 0xD1
	Int32     = 0xD2
	Int64     = 0xD3
	MapFix    = 0x80
	Map16     = 0xDE
	Map32     = 0xDF
	SliceFix  = 0x90
	Slice16   = 0xDC
	Slice32   = 0xDD
	StringFix = 0xA0
	String8   = 0xD9
	String16  = 0xDA
	String32  = 0xDB
)

const (
	Len4  = (1 << 4) - 1
	Len5  = (1 << 5) - 1
	Len8  = (1 << 8) - 1
	Len16 = (1 << 16) - 1
	Len32 = (1 << 32) - 1
)

type InvalidTagErr byte

func (i InvalidTagErr) Error() string {
	return fmt.Sprintf("pack: invalid tag found %#02x", byte(i))
}

type UnsupportedTypeErr string

func (u UnsupportedTypeErr) Error() string {
	return fmt.Sprintf("pack: unsupported type: %s", string(u))
}

var (
	TooManyValuesErr = errors.New("pack: too many values")
	TooFewValuesErr  = errors.New("pack: too few values")
)

func MarshalCompact(d interface{}) ([]byte, error) {
	return runMarshal(reflect.ValueOf(d), true)
}

func Marshal(d interface{}) ([]byte, error) {
	return runMarshal(reflect.ValueOf(d), false)
}

func runMarshal(v reflect.Value, compact bool) ([]byte, error) {
	var buf bytes.Buffer
	if err := marshal(v, &buf, compact); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Unmarshal(data []byte, d interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(data)

	if err := unmarshal(reflect.ValueOf(d).Elem(), buf, false); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func UnmarshalCompact(data []byte, d interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(data)

	if err := unmarshal(reflect.ValueOf(d).Elem(), buf, true); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
