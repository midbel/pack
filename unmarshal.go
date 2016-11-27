package pack

import (
	"bytes"
	"encoding/binary"
	"math"
	"reflect"
)

func unmarshal(v reflect.Value, buf *bytes.Buffer, compact bool) error {
	b, err := buf.ReadByte()
	if err != nil {
		return err
	}
	if b == Nil {
		return nil
	}
	switch t := v.Kind(); t {
	case reflect.Struct:
		return unmarshalStruct(v, b, buf, compact)
	case reflect.Slice:
		return unmarshalSlice(v, b, buf, compact)
	case reflect.Map:
		return unmarshalMap(v, b, buf, compact)
	case reflect.Interface:
		buf.UnreadByte()
		return unmarshalInterface(v, b, buf, compact)
	case reflect.Ptr:
		buf.UnreadByte()
		return unmarshal(v.Elem(), buf, compact)
	default:
		return decode(v, b, buf)
	}
	return nil
}

func unmarshalInterface(v reflect.Value, b byte, buf *bytes.Buffer, compact bool) error {
	var val reflect.Value
	switch {
	case b == Int8:
		val = reflect.ValueOf(new(int8))
	case b == Int16:
		val = reflect.ValueOf(new(int16))
	case b == Int32:
		val = reflect.ValueOf(new(int32))
	case b == Int64:
		val = reflect.ValueOf(new(int64))
	case b == Uint8:
		val = reflect.ValueOf(new(uint8))
	case b == Uint16:
		val = reflect.ValueOf(new(uint16))
	case b == Uint32:
		val = reflect.ValueOf(new(uint32))
	case b == Uint64:
		val = reflect.ValueOf(new(uint64))
	case b == Float32:
		val = reflect.ValueOf(new(float32))
	case b == Float64:
		val = reflect.ValueOf(new(float64))
	case StringFix>>5 == b>>5 || b == String8 || b == String16 || b == String32:
		val = reflect.ValueOf(new(string))
	case b == Bin8 || b == Bin16 || b == Bin32:
		val = reflect.ValueOf(new(string))
	case b == True || b == False:
		val = reflect.ValueOf(new(bool))
	case b == Nil:
		return nil
	default:
		return nil
	}
	val = val.Elem()
	if err := unmarshal(val, buf, compact); err != nil {
		return err
	}
	v.Set(val)
	return nil
}

func unmarshalStruct(v reflect.Value, b byte, buf *bytes.Buffer, compact bool) error {
	var length int
	switch {
	case b>>4 == MapFix>>4:
		length = int(b & 0x0F)
	case b == Map16:
		length = int(binary.BigEndian.Uint16(buf.Next(2)))
	case b == Map32:
		length = int(binary.BigEndian.Uint32(buf.Next(4)))
	default:
		return InvalidTagErr(b)
	}
	count := v.NumField()
	if count > length {
		return TooFewValuesErr
	}

	for i := 0; i < count; i++ {
		f := v.Field(i)
		if !f.CanSet() {
			continue
		}
		if !compact {
			b, err := buf.ReadByte()
			if err != nil {
				return err
			}
			var name string
			val := reflect.ValueOf(&name).Elem()
			if err := decode(val, b, buf); err != nil {
				return err
			}
		}
		if err := unmarshal(f, buf, compact); err != nil {
			return err
		}
	}
	return nil
}

func unmarshalMap(v reflect.Value, b byte, buf *bytes.Buffer, compact bool) error {
	var length int
	switch {
	case b>>4 == MapFix>>4:
		length = int(b & 0x0F)
	case b == Map16:
		length = int(binary.BigEndian.Uint16(buf.Next(2)))
	case b == Map32:
		length = int(binary.BigEndian.Uint32(buf.Next(4)))
	default:
		return InvalidTagErr(b)
	}
	for i := 0; i < length; i++ {
		key := reflect.New(v.Type().Key()).Elem()
		if err := unmarshal(key, buf, compact); err != nil {
			return err
		}
		value := reflect.New(v.Type().Elem()).Elem()
		if err := unmarshal(value, buf, compact); err != nil {
			return err
		}
		v.SetMapIndex(key, value)
	}
	return nil
}

func unmarshalSlice(v reflect.Value, b byte, buf *bytes.Buffer, compact bool) error {
	var length int
	switch {
	case b>>4 == SliceFix>>4:
		length = int(b & 0x0F)
	case b == Slice16:
		length = int(binary.BigEndian.Uint16(buf.Next(2)))
	case b == Slice32:
		length = int(binary.BigEndian.Uint32(buf.Next(4)))
	default:
		return InvalidTagErr(b)
	}
	for i := 0; i < length; i++ {
		f := reflect.New(v.Type().Elem()).Elem()
		if err := unmarshal(f, buf, compact); err != nil {
			return err
		}
		v.Set(reflect.Append(v, f))
	}
	return nil
}

func decode(v reflect.Value, b byte, buf *bytes.Buffer) error {
	switch t := v.Kind(); t {
	case reflect.Bool:
		if b == True {
			v.SetBool(true)
		} else {
			v.SetBool(false)
		}
	case reflect.String:
		var length int
		switch {
		case b>>5 == StringFix>>5:
			length = int(b & 0X1F)
		case b == String8 || b == Bin8:
			if l, err := buf.ReadByte(); err != nil {
				return err
			} else {
				length = int(l)
			}
		case b == String16 || b == Bin16:
			length = int(binary.BigEndian.Uint16(buf.Next(2)))
		case b == String32 || b == Bin32:
			length = int(binary.BigEndian.Uint16(buf.Next(4)))
		default:
			return InvalidTagErr(b)
		}
		v.SetString(string(buf.Next(length)))
	case reflect.Float32:
		if b != Float32 {
			return InvalidTagErr(b)
		}
		i := binary.BigEndian.Uint32(buf.Next(4))
		f := math.Float32frombits(i)

		v.SetFloat(float64(f))
	case reflect.Float64:
		if b != Float64 {
			return InvalidTagErr(b)
		}
		i := binary.BigEndian.Uint64(buf.Next(8))
		f := math.Float64frombits(i)

		v.SetFloat(f)
	case reflect.Int8:
		if b == Int8 {
			i, err := buf.ReadByte()
			if err != nil {
				return err
			}
			v.SetInt(int64(i))
		} else {
			return InvalidTagErr(b)
		}
	case reflect.Int16:
		if b == Int16 {
			i := binary.BigEndian.Uint16(buf.Next(2))
			v.SetInt(int64(i))
		} else {
			return InvalidTagErr(b)
		}
	case reflect.Int32:
		if b == Int32 {
			i := binary.BigEndian.Uint32(buf.Next(4))
			v.SetInt(int64(i))
		} else {
			return InvalidTagErr(b)
		}
	case reflect.Int64, reflect.Int:
		if b == Int64 {
			i := binary.BigEndian.Uint64(buf.Next(8))
			v.SetInt(int64(i))
		} else {
			return InvalidTagErr(b)
		}
	case reflect.Uint8:
		if b == Uint8 {
			i, err := buf.ReadByte()
			if err != nil {
				return err
			}
			v.SetUint(uint64(i))
		} else {
			return InvalidTagErr(b)
		}
	case reflect.Uint16:
		if b == Uint16 {
			i := binary.BigEndian.Uint16(buf.Next(2))
			v.SetUint(uint64(i))
		} else {
			return InvalidTagErr(b)
		}
	case reflect.Uint32:
		if b == Uint32 {
			i := binary.BigEndian.Uint32(buf.Next(4))
			v.SetUint(uint64(i))
		} else {
			return InvalidTagErr(b)
		}
	case reflect.Uint64, reflect.Uint:
		if b == Uint64 {
			i := binary.BigEndian.Uint64(buf.Next(8))
			v.SetUint(i)
		} else {
			return InvalidTagErr(b)
		}
	default:
		return UnsupportedTypeErr(t.String())
	}
	return nil
}
