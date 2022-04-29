package media

type Type string
type SupportMajorBrand string
type Extension string

const (
	MP4 Type = "MP4"
	MOV Type = "MOV(QUICKTIME)"
)

const (
	MP42     SupportMajorBrand = "mp42"
	SONYXAVC SupportMajorBrand = "XAVC"
	QT       SupportMajorBrand = "qt  "
	NIKO     SupportMajorBrand = "niko"
)

const (
	Mp4Extension  Extension = ".MP4"
	MovExtension  Extension = ".MOV"
	NRAWExtension Extension = ".NEV"
)
