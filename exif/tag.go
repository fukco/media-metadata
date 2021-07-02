package exif

import (
	"bytes"
	"encoding/binary"
	"strings"
)

type ExifTag struct {
	ID            uint16
	Name          string
	Value         string
	Undefined     bool        `json:"-"`
	OriginalValue interface{} `json:"-"`
	Group         []string    `json:"-"`
}

type TagDefinition struct {
	Name             string
	Fn               func(interface{}) string
	SubTagDefinition *SubTagDefinition
}

type SubTagDefinition struct {
	tagDefinitionType   tagDefinitionType
	subTagDefinitionMap map[interface{}]*TagDefinition
}

type tagDefinitionType int

const (
	byId tagDefinitionType = iota
	byIndex
)

func (t *TiffEntry) convertVals() error {
	r := bytes.NewReader(t.Val)
	switch t.Type {
	case DTAscii:
		if t.Count <= 0 {
			break
		}
		t.StrVals = strings.Split(string(t.Val[:t.Count-1]), string(byte(0)))
	case DTByte:
		var v uint8
		t.IntVals = make([]int64, int(t.Count))
		for i := range t.IntVals {
			err := binary.Read(r, t.Order, &v)
			if err != nil {
				return err
			}
			t.IntVals[i] = int64(v)
		}
	case DTShort:
		var v uint16
		t.IntVals = make([]int64, int(t.Count))
		for i := range t.IntVals {
			err := binary.Read(r, t.Order, &v)
			if err != nil {
				return err
			}
			t.IntVals[i] = int64(v)
		}
	case DTLong:
		var v uint32
		t.IntVals = make([]int64, int(t.Count))
		for i := range t.IntVals {
			err := binary.Read(r, t.Order, &v)
			if err != nil {
				return err
			}
			t.IntVals[i] = int64(v)
		}
	case DTSByte:
		var v int8
		t.IntVals = make([]int64, int(t.Count))
		for i := range t.IntVals {
			err := binary.Read(r, t.Order, &v)
			if err != nil {
				return err
			}
			t.IntVals[i] = int64(v)
		}
	case DTSShort:
		var v int16
		t.IntVals = make([]int64, int(t.Count))
		for i := range t.IntVals {
			err := binary.Read(r, t.Order, &v)
			if err != nil {
				return err
			}
			t.IntVals[i] = int64(v)
		}
	case DTSLong:
		var v int32
		t.IntVals = make([]int64, int(t.Count))
		for i := range t.IntVals {
			err := binary.Read(r, t.Order, &v)
			if err != nil {
				return err
			}
			t.IntVals[i] = int64(v)
		}
	case DTRational:
		t.RatVals = make([][]int64, int(t.Count))
		for i := range t.RatVals {
			var n, d uint32
			err := binary.Read(r, t.Order, &n)
			if err != nil {
				return err
			}
			err = binary.Read(r, t.Order, &d)
			if err != nil {
				return err
			}
			t.RatVals[i] = []int64{int64(n), int64(d)}
		}
	case DTSRational:
		t.RatVals = make([][]int64, int(t.Count))
		for i := range t.RatVals {
			var n, d int32
			err := binary.Read(r, t.Order, &n)
			if err != nil {
				return err
			}
			err = binary.Read(r, t.Order, &d)
			if err != nil {
				return err
			}
			t.RatVals[i] = []int64{int64(n), int64(d)}
		}
	case DTFloat: // float32
		t.FloatVals = make([]float64, int(t.Count))
		for i := range t.FloatVals {
			var v float32
			err := binary.Read(r, t.Order, &v)
			if err != nil {
				return err
			}
			t.FloatVals[i] = float64(v)
		}
	case DTDouble:
		t.FloatVals = make([]float64, int(t.Count))
		for i := range t.FloatVals {
			var u float64
			err := binary.Read(r, t.Order, &u)
			if err != nil {
				return err
			}
			t.FloatVals[i] = u
		}
	}

	switch t.Type {
	case DTByte, DTShort, DTLong, DTSByte, DTSShort, DTSLong:
		t.Format = IntVal
	case DTRational, DTSRational:
		t.Format = RatVal
	case DTFloat, DTDouble:
		t.Format = FloatVal
	case DTAscii:
		t.Format = StringVal
	case DTUndefined:
		t.Format = UndefVal
	default:
		t.Format = OtherVal
	}
	return nil
}
