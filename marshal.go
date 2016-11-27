package pack

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strings"
	"unicode/utf8"
)

func marshal(v reflect.Value, buf *bytes.Buffer, compact bool) error {
	switch t := v.Kind(); t {
	case reflect.Struct:
		return marshalStruct(v, buf, compact)
	case reflect.Map:
		if v.IsNil() {
			buf.WriteByte(Nil)
		} else {
			return marshalMap(v, buf, compact)
		}
	case reflect.Slice:
		if v.IsNil() {
			buf.WriteByte(Nil)
		} else {
			return marshalSlice(v, buf, compact)
		}
	case reflect.Ptr:
		if v.IsNil() {
			buf.WriteByte(Nil)
		} else {
			return marshal(v.Elem(), buf, compact)
		}
	case reflect.Interface:
		if v.IsNil() {
			buf.WriteByte(Nil)
		} else {
			return marshal(reflect.ValueOf(v.Interface()), buf, compact)
		}
	default:
		return encode(v, buf)
	}
	return nil
}

func marshalStruct(v reflect.Value, buf *bytes.Buffer, compact bool) error {
	length := v.NumField()

	var (
		tag   byte
		chunk interface{}
	)
	switch {
	case length <= Len4:
		tag = byte(MapFix | length)
	case length <= Len16:
		tag = Map16
		chunk = uint16(length)
	case length <= Len32:
		tag = Map32
		chunk = uint32(length)
	default:
		return TooManyValuesErr
	}
	buf.WriteByte(tag)
	if chunk != nil {
		binary.Write(buf, binary.BigEndian, chunk)
	}

	t := v.Type()
	for i := 0; i < length; i++ {
		f := v.Field(i)
		if !f.CanSet() {
			continue
		}
		if !compact {
			name := strings.ToLower(t.Field(i).Name)
			if err := encode(reflect.ValueOf(name), buf); err != nil {
				return err
			}
		}
		if err := marshal(f, buf, compact); err != nil {
			return err
		}
	}
	return nil
}

func marshalMap(v reflect.Value, buf *bytes.Buffer, compact bool) error {
	var (
		tag   byte
		chunk interface{}
	)
	length := v.Len()
	switch {
	case length <= Len4:
		tag = byte(MapFix | length)
	case length <= Len16:
		tag = Map16
		chunk = uint16(length)
	case length <= Len32:
		tag = Map32
		chunk = uint32(length)
	default:
		return TooManyValuesErr
	}
	buf.WriteByte(tag)
	if chunk != nil {
		binary.Write(buf, binary.BigEndian, chunk)
	}
	for _, k := range v.MapKeys() {
		if err := marshal(k, buf, compact); err != nil {
			return err
		}
		if err := marshal(v.MapIndex(k), buf, compact); err != nil {
			return err
		}
	}
	return nil
}

func marshalSlice(v reflect.Value, buf *bytes.Buffer, compact bool) error {
	length := v.Len()

	var (
		tag   byte
		chunk interface{}
	)
	switch {
	case length <= Len4:
		tag = byte(SliceFix | length)
	case length <= Len16:
		tag = Slice16
		chunk = uint16(length)
	case length <= Len32:
		tag = Slice32
		chunk = uint32(length)
	default:
		return TooManyValuesErr
	}
	buf.WriteByte(tag)
	if chunk != nil {
		binary.Write(buf, binary.BigEndian, chunk)
	}
	for i := 0; i < length; i++ {
		f := v.Index(i)
		if err := marshal(f, buf, compact); err != nil {
			return err
		}
	}
	return nil
}

func encode(v reflect.Value, buf *bytes.Buffer) error {
	switch t := v.Kind(); t {
	case reflect.String:
		var (
			tag   byte
			chunk interface{}
		)
		count := v.Len()
		if utf8.ValidString(v.String()) {
			switch {
			case count <= Len5:
				tag = byte(StringFix | count)
			case count <= Len8:
				tag = String8
				chunk = uint8(count)
			case count <= Len16:
				tag = String16
				chunk = uint16(count)
			case count <= Len32:
				tag = String32
				chunk = uint32(count)
			default:
				return fmt.Errorf("string too long")
			}
		} else {
			switch {
			case count <= Len8:
				tag = Bin8
				chunk = uint8(count)
			case count <= Len16:
				tag = Bin16
				chunk = uint16(count)
			case count <= Len32:
				tag = Bin32
				chunk = uint32(count)
			}
		}
		buf.WriteByte(tag)
		if chunk != nil {
			binary.Write(buf, binary.BigEndian, chunk)
		}
		buf.Write([]byte(v.String()))
	case reflect.Float32:
		buf.WriteByte(Float32)
		binary.Write(buf, binary.BigEndian, math.Float32bits(float32(v.Float())))
	case reflect.Float64:
		buf.WriteByte(Float64)
		binary.Write(buf, binary.BigEndian, math.Float64bits(float64(v.Float())))
	case reflect.Int8:
		buf.WriteByte(Int8)
		binary.Write(buf, binary.BigEndian, int8(v.Int()))
	case reflect.Int16:
		buf.WriteByte(Int16)
		binary.Write(buf, binary.BigEndian, int16(v.Int()))
	case reflect.Int32:
		buf.WriteByte(Int32)
		binary.Write(buf, binary.BigEndian, int32(v.Int()))
	case reflect.Int64, reflect.Int:
		buf.WriteByte(Int64)
		binary.Write(buf, binary.BigEndian, v.Int())
	case reflect.Uint8:
		buf.WriteByte(Uint8)
		binary.Write(buf, binary.BigEndian, uint8(v.Int()))
	case reflect.Uint16:
		buf.WriteByte(Uint16)
		binary.Write(buf, binary.BigEndian, uint16(v.Int()))
	case reflect.Uint32:
		buf.WriteByte(Uint32)
		binary.Write(buf, binary.BigEndian, uint32(v.Int()))
	case reflect.Uint64, reflect.Uint:
		buf.WriteByte(Uint64)
		binary.Write(buf, binary.BigEndian, v.Uint())
	case reflect.Bool:
		if v.Bool() {
			buf.WriteByte(True)
		} else {
			buf.WriteByte(False)
		}
	default:
		return UnsupportedTypeErr(t.String())
	}
	return nil
}
