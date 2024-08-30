package internal

type Format int
type Extension string

const (
	Unknown Format = iota
	Mp4
	Quicktime
)

const (
	Mp4Extension  Extension = ".MP4"
	MovExtension  Extension = ".MOV"
	NRAWExtension Extension = ".NEV"
)
