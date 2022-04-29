package exif

import "encoding/binary"

type TiffHeader struct {
	Order          binary.ByteOrder
	FirstIFDOffset uint32
}

type Group []string

type TiffDirectory struct {
	Group               Group
	NumOfDE             uint16
	Entries             []*TiffEntry
	NextDirectoryOffset uint32
}

type TiffEntry struct {
	TagId       uint16
	Type        DataType
	Count       uint32
	ValOrOffset uint32
	Val         []byte
	Order       binary.ByteOrder
	IntVals     []int64
	FloatVals   []float64
	RatVals     [][]int64
	StrVals     []string
	Format      Format
}

// Format specifies the Go type equivalent used to represent the basic
// tiff data types.
type Format int

const (
	IntVal Format = iota
	FloatVal
	RatVal
	StringVal
	UndefVal
	OtherVal
)

// DataType represents the basic tiff tag data types.
type DataType uint16

const (
	DTByte      DataType = 1
	DTAscii     DataType = 2
	DTShort     DataType = 3
	DTLong      DataType = 4
	DTRational  DataType = 5
	DTSByte     DataType = 6
	DTUndefined DataType = 7
	DTSShort    DataType = 8
	DTSLong     DataType = 9
	DTSRational DataType = 10
	DTFloat     DataType = 11
	DTDouble    DataType = 12
)

var typeNames = map[DataType]string{
	DTByte:      "byte",
	DTAscii:     "ascii",
	DTShort:     "short",
	DTLong:      "long",
	DTRational:  "rational",
	DTSByte:     "signed byte",
	DTUndefined: "undefined",
	DTSShort:    "signed short",
	DTSLong:     "signed long",
	DTSRational: "signed rational",
	DTFloat:     "float",
	DTDouble:    "double",
}

// TypeSize specifies the size in bytes of each type.
var TypeSize = map[DataType]uint32{
	DTByte:      1,
	DTAscii:     1,
	DTShort:     2,
	DTLong:      4,
	DTRational:  8,
	DTSByte:     1,
	DTUndefined: 1,
	DTSShort:    2,
	DTSLong:     4,
	DTSRational: 8,
	DTFloat:     4,
	DTDouble:    8,
}
