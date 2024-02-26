package media

type Type string
type Brand string
type Extension string

const (
	MP4 Type = "MP4"
	MOV Type = "MOV(QUICKTIME)"
)

const (
	MP42     Brand = "mp42"
	SONYXAVC Brand = "XAVC"
	QT       Brand = "qt  "
	NIKO     Brand = "niko"
	PANA     Brand = "pana"
)

const (
	Mp4Extension  Extension = ".MP4"
	MovExtension  Extension = ".MOV"
	NRAWExtension Extension = ".NEV"
)
