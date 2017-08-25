package AMF

import (
	"encoding/binary"
	"fmt"
	"math"
	"rtmpmate.com/util/AMF/Types"
)

type AMFValue struct {
	AMFType uint8
	Key     string
	Data    interface{}
	Cost    int
	Ended   bool
}

type AMFObject struct {
	Data  map[string]interface{}
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
	Data float64
	Cost int
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
	v.AMFType = val.AMFType
	v.Key = key.Data
	v.Data = val.Data
	v.Cost = key.Cost + val.Cost

	return &v, err
}

func DecodeObject(data []byte, offset int, size int) (*AMFObject, error) {
	if size < 3 {
		return nil, fmt.Errorf("Data not enough while decoding AMF object")
	}

	var obj = make(map[string]interface{})
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

		obj[key.Data] = val.Data
	}

	var v AMFObject
	v.Data = obj
	v.Cost = pos
	v.Ended = ended

	return &v, nil
}

func DecodeString(data []byte, offset int, size int) (*AMFString, error) {
	if size < 2 {
		return nil, fmt.Errorf("Data not enough while decoding AMF string")
	}

	var pos = 0
	var length = binary.BigEndian.Uint16(data[offset+pos : offset+pos+2])

	pos += 2

	var str string
	if length > 0 {
		str = string(data[offset+pos : offset+pos+int(length)])
		pos += int(length)
	}

	var v AMFString
	v.Data = str
	v.Cost = pos

	return &v, nil
}

func DecodeLongString(data []byte, offset int, size int) (*AMFLongString, error) {
	if size < 4 {
		return nil, fmt.Errorf("Data not enough while decoding AMF long string")
	}

	var pos = 0
	var length = binary.BigEndian.Uint32(data[offset+pos : offset+pos+4])

	pos += 4

	var str string
	if length > 0 {
		str = string(data[offset+pos : offset+pos+int(length)])
		pos += int(length)
	}

	var v AMFLongString
	v.Data = str
	v.Cost = pos

	return &v, nil
}

func DecodeDate(data []byte, offset int, size int) (*AMFDate, error) {
	if size < 4 {
		return nil, fmt.Errorf("Data not enough while decoding AMF date")
	}

	var pos = 0
	var bits = binary.BigEndian.Uint64(data[offset+pos : offset+pos+8])
	var timestamp = math.Float64frombits(bits)

	pos += 8

	var timeoffset = binary.BigEndian.Uint16(data[offset+pos : offset+pos+2])
	timestamp += float64(timeoffset) * 60 * 1000

	pos += 2

	var v AMFDate
	v.Data = timestamp
	v.Cost = pos

	return &v, nil
}

func DecodeValue(data []byte, offset int, size int) (*AMFValue, error) {
	if size < 1 {
		return nil, fmt.Errorf("Data not enough while decoding AMF value")
	}

	var pos = 0
	var valuetype = uint8(data[offset+pos : offset+pos+1][0])

	pos += 1

	var v AMFValue
	var ended = false

	switch valuetype {
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

		var arr = make([]interface{}, 0, 32)
		for i := uint32(0); i < length; i++ {
			val, err := DecodeValue(data, offset+pos, size-pos)
			if err != nil {
				return nil, err
			}

			arr = append(arr, val.Data)
			pos += val.Cost
		}

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
		return nil, fmt.Errorf("Skipping unsupported AMF value type(%x)", valuetype)
	}

	v.AMFType = valuetype
	v.Cost = pos
	v.Ended = ended

	return &v, nil
}
