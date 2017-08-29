package AMF

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"fmt"
	"math"
	"rtmpmate.com/util/AMF/Types"
	"syscall"
)

const (
	AMF0 byte = 0
	AMF3 byte = 3
)

type AMFValue struct {
	Type   byte
	Key    string
	Data   interface{}
	Offset int16
	Cost   int
	Ended  bool
}

type AMFObject struct {
	Data  list.List
	Cost  int
	Ended bool
}

type AMFString struct {
	Data string
	Cost int
}

type AMFLongString struct {
	Data string
	Cost int
}

type AMFDate struct {
	Data   float64
	Offset int16
	Cost   int
}

func Decode(data []byte, offset int, size int) (*AMFValue, error) {
	key, err := DecodeString(data, offset, size)
	if err != nil {
		return nil, err
	}

	val, err := DecodeValue(data, offset+key.Cost, size-key.Cost)
	if err != nil {
		return nil, err
	}

	var v AMFValue
	v.Type = val.Type
	v.Key = key.Data
	v.Data = val.Data
	v.Cost = key.Cost + val.Cost

	return &v, err
}

func DecodeString(data []byte, offset int, size int) (*AMFString, error) {
	if size < 2 {
		return nil, fmt.Errorf("Data not enough while decoding AMF string")
	}

	var v AMFString

	var pos = 0
	var length = binary.BigEndian.Uint16(data[offset+pos : offset+pos+2])

	pos += 2

	if length > 0 {
		v.Data += string(data[offset+pos : offset+pos+int(length)])
		pos += int(length)
	}

	v.Cost = pos

	return &v, nil
}

func DecodeObject(data []byte, offset int, size int) (*AMFObject, error) {
	if size < 3 {
		return nil, fmt.Errorf("Data not enough while decoding AMF object")
	}

	var v AMFObject

	var pos = 0
	var ended = false

	for ended == false && pos < size {
		key, err := DecodeString(data, offset+pos, size-pos)
		if err != nil {
			return nil, err
		}

		pos += key.Cost

		val, err := DecodeValue(data, offset+pos, size-pos)
		if err != nil {
			return nil, err
		}

		pos += val.Cost
		ended = val.Ended

		val.Key = key.Data
		v.Data.PushBack(val)
	}

	v.Cost = pos
	v.Ended = ended

	return &v, nil
}

func DecodeDate(data []byte, offset int, size int) (*AMFDate, error) {
	if size < 10 {
		return nil, fmt.Errorf("Data not enough while decoding AMF date")
	}

	var v AMFDate

	var pos = 0
	var bits = binary.BigEndian.Uint64(data[offset+pos : offset+pos+8])
	v.Data = math.Float64frombits(bits)

	pos += 8

	var timeoffset = binary.BigEndian.Uint16(data[offset+pos : offset+pos+2])
	v.Data += float64(timeoffset) * 60 * 1000

	pos += 2

	v.Cost = pos

	return &v, nil
}

func DecodeLongString(data []byte, offset int, size int) (*AMFLongString, error) {
	if size < 4 {
		return nil, fmt.Errorf("Data not enough while decoding AMF long string")
	}

	var v AMFLongString

	var pos = 0
	var length = binary.BigEndian.Uint32(data[offset+pos : offset+pos+4])

	pos += 4

	if length > 0 {
		v.Data += string(data[offset+pos : offset+pos+int(length)])
		pos += int(length)
	}

	v.Cost = pos

	return &v, nil
}

func DecodeValue(data []byte, offset int, size int) (*AMFValue, error) {
	if size < 1 {
		return nil, fmt.Errorf("Data not enough while decoding AMF value")
	}

	var pos = 0
	var valueType = uint8(data[offset+pos : offset+pos+1][0])

	pos += 1

	var v AMFValue
	var ended = false

	switch valueType {
	case Types.DOUBLE:
		var bits = binary.BigEndian.Uint64(data[offset+pos : offset+pos+8])
		v.Data = math.Float64frombits(bits)
		pos += 8

	case Types.BOOLEAN:
		var bit = uint8(data[offset+pos : offset+pos+1][0])
		if bit > 0 {
			v.Data = true
		}
		pos += 1

	case Types.STRING:
		str, err := DecodeString(data, offset+pos, size-pos)
		if err != nil {
			return nil, err
		}

		v.Data = str.Data
		pos += str.Cost

	case Types.OBJECT:
		obj, err := DecodeObject(data, offset+pos, size-pos)
		if err != nil {
			return nil, err
		}

		v.Data = obj.Data
		pos += obj.Cost

	case Types.NULL:
	case Types.UNDEFINED:

	case Types.MIXED_ARRAY:
		arr, err := DecodeObject(data, offset+pos, size-pos)
		if err != nil {
			return nil, err
		}

		v.Data = arr.Data
		pos += arr.Cost

	case Types.END_OF_OBJECT:
		ended = true

	case Types.ARRAY:
		var length = binary.BigEndian.Uint32(data[offset+pos : offset+pos+4])
		pos += 4

		var arr list.List
		for i := uint32(0); i < length; i++ {
			val, err := DecodeValue(data, offset+pos, size-pos)
			if err != nil {
				return nil, err
			}

			arr.PushBack(val)
			pos += val.Cost
		}

		v.Data = arr

	case Types.DATE:
		date, err := DecodeDate(data, offset+pos, size-pos)
		if err != nil {
			return nil, err
		}

		v.Data = date.Data
		pos += date.Cost

	case Types.LONG_STRING:
		longstr, err := DecodeLongString(data, offset+pos, size-pos)
		if err != nil {
			return nil, err
		}

		v.Data = longstr.Data
		pos += longstr.Cost

	default:
		pos = size
		return nil, fmt.Errorf("Skipping unsupported AMF value type(%x)", valueType)
	}

	v.Type = valueType
	v.Cost = pos
	v.Ended = ended

	return &v, nil
}

type Encoder struct {
	buffer bytes.Buffer
}

func (this *Encoder) AppendInt8(n int8) error {
	return this.buffer.WriteByte(byte(n))
}

func (this *Encoder) AppendInt16(n int16, littleEndian bool) error {
	var order binary.ByteOrder

	if littleEndian {
		order = binary.LittleEndian
	} else {
		order = binary.BigEndian
	}

	err := binary.Write(&this.buffer, order, &n)
	if err != nil {
		return err
	}

	return nil
}

func (this *Encoder) AppendInt24(n int32, littleEndian bool) error {
	if littleEndian {
		this.buffer.WriteByte(byte(n & 0xFF))
		this.buffer.WriteByte(byte((n >> 8) & 0xFF))
		this.buffer.WriteByte(byte((n >> 16) & 0xFF))
	} else {
		this.buffer.WriteByte(byte((n >> 16) & 0xFF))
		this.buffer.WriteByte(byte((n >> 8) & 0xFF))
		this.buffer.WriteByte(byte(n & 0xFF))
	}

	return nil
}

func (this *Encoder) AppendInt32(n int32, littleEndian bool) error {
	var order binary.ByteOrder

	if littleEndian {
		order = binary.LittleEndian
	} else {
		order = binary.BigEndian
	}

	err := binary.Write(&this.buffer, order, &n)
	if err != nil {
		return err
	}

	return nil
}

func (this *Encoder) Encode(k string, o *AMFObject) error {
	err := this.EncodeString(k)
	if err != nil {
		return err
	}

	err = this.buffer.WriteByte(Types.OBJECT)
	if err != nil {
		return err
	}

	err = this.EncodeObject(o)
	if err != nil {
		return err
	}

	err = this.buffer.WriteByte(Types.END_OF_OBJECT)
	if err != nil {
		return err
	}

	return nil
}

func (this *Encoder) EncodeNumber(n float64) error {
	err := this.buffer.WriteByte(Types.DOUBLE)
	if err != nil {
		return err
	}

	err = binary.Write(&this.buffer, binary.BigEndian, &n)
	if err != nil {
		return err
	}

	return nil
}

func (this *Encoder) EncodeBoolean(b bool) error {
	err := this.buffer.WriteByte(Types.BOOLEAN)
	if err != nil {
		return err
	}

	var tmp byte
	if b {
		tmp = 1
	} else {
		tmp = 0
	}

	err = this.buffer.WriteByte(tmp)
	if err != nil {
		return err
	}

	return nil
}

func (this *Encoder) EncodeString(s string) error {
	size := len(s)
	if size == 0 {
		return syscall.EINVAL
	}

	if size >= 0xFFFF {
		return this.encodeLongString(s)
	}

	err := this.buffer.WriteByte(Types.STRING)
	if err != nil {
		return err
	}

	tmp := uint16(size)
	err = binary.Write(&this.buffer, binary.BigEndian, &tmp)
	if err != nil {
		return err
	}

	_, err = this.buffer.Write([]byte(s))
	if err != nil {
		return err
	}

	return nil
}

func (this *Encoder) EncodeObject(o *AMFObject) error {
	for item := o.Data.Front(); item != nil; item = item.Next() {
		v := item.Value.(*AMFValue)
		if len(v.Key) > 0 {
			err := this.EncodeString(v.Key)
			if err != nil {
				return err
			}
		}

		switch v.Type {
		case Types.DOUBLE:
			this.EncodeNumber(v.Data.(float64))

		case Types.BOOLEAN:
			this.EncodeBoolean(v.Data.(bool))

		case Types.STRING:
			this.EncodeString(v.Data.(string))

		case Types.OBJECT:
			err := this.buffer.WriteByte(Types.OBJECT)
			if err != nil {
				return err
			}

			obj := AMFObject{v.Data.(list.List), 0, true}
			this.EncodeObject(&obj)

			err = this.buffer.WriteByte(Types.END_OF_OBJECT)
			if err != nil {
				return err
			}

		case Types.NULL:
			this.EncodeNull()

		case Types.UNDEFINED:
			this.EncodeUndefined()

		case Types.MIXED_ARRAY:
			obj := AMFObject{v.Data.(list.List), 0, true}
			this.EncodeMixedArray(&obj)

		case Types.ARRAY:
			obj := AMFObject{v.Data.(list.List), 0, true}
			this.EncodeArray(&obj)

		case Types.DATE:
			this.EncodeDate(v.Data.(float64), v.Offset)

		case Types.LONG_STRING:
			this.encodeLongString(v.Data.(string))

		default:
		}
	}

	return nil
}

func (this *Encoder) EncodeNull() error {
	return this.buffer.WriteByte(byte(Types.NULL))
}

func (this *Encoder) EncodeUndefined() error {
	return this.buffer.WriteByte(byte(Types.UNDEFINED))
}

func (this *Encoder) EncodeMixedArray(o *AMFObject) error {
	err := this.buffer.WriteByte(Types.MIXED_ARRAY)
	if err != nil {
		return err
	}

	tmp := uint32(o.Data.Len())
	err = binary.Write(&this.buffer, binary.BigEndian, &tmp)
	if err != nil {
		return err
	}

	this.EncodeObject(o)

	err = this.buffer.WriteByte(Types.END_OF_OBJECT)
	if err != nil {
		return err
	}

	return nil
}

func (this *Encoder) EncodeArray(o *AMFObject) error {
	err := this.buffer.WriteByte(Types.ARRAY)
	if err != nil {
		return err
	}

	tmp := uint32(o.Data.Len())
	err = binary.Write(&this.buffer, binary.BigEndian, &tmp)
	if err != nil {
		return err
	}

	this.EncodeObject(o)

	err = this.buffer.WriteByte(Types.END_OF_OBJECT)
	if err != nil {
		return err
	}

	return nil
}

func (this *Encoder) EncodeDate(ts float64, off int16) error {
	err := this.buffer.WriteByte(byte(Types.DATE))
	if err != nil {
		return err
	}

	err = binary.Write(&this.buffer, binary.BigEndian, &ts)
	if err != nil {
		return err
	}

	err = binary.Write(&this.buffer, binary.BigEndian, &off)
	if err != nil {
		return err
	}

	return nil
}

func (this *Encoder) encodeLongString(ls string) error {
	size := len(ls)
	if size == 0 {
		return syscall.EINVAL
	}

	err := this.buffer.WriteByte(Types.LONG_STRING)
	if err != nil {
		return err
	}

	tmp := uint32(size)
	err = binary.Write(&this.buffer, binary.BigEndian, &tmp)
	if err != nil {
		return err
	}

	_, err = this.buffer.Write([]byte(ls))
	if err != nil {
		return err
	}

	return nil
}
