package exif

type ExifMeta struct {
	Tags map[string][]*ExifTag
}

type Base struct {
	TiffHeader  *TiffHeader
	Directories []*TiffDirectory
	MakerIFD    *MakerDirectory
}

type MakerDirectory struct {
	Directories []*TiffDirectory
	Maker       Maker
}

type Maker string

const (
	Canon     Maker = "Canon"
	Panasonic Maker = "Panasonic"
	FUJIFILM  Maker = "FUJIFILM"
)

type DirectoryType string

const (
	IFD0     = DirectoryType(GroupIFD0)
	ExifIFD  = DirectoryType(GroupExif)
	IFD1     = DirectoryType(GroupIFD1)
	MakerIFD = DirectoryType(GroupMakerNotes)
)

type GroupName string

const (
	GroupIFD0                GroupName = "IFD0"
	GroupIFD1                GroupName = "IFD1"
	GroupExif                GroupName = "ExifIFD"
	GroupMakerNotes          GroupName = "Maker"
	GroupCanonCameraSettings GroupName = "Canon CameraSettings"
	GroupCanonShotInfo       GroupName = "Canon ShotInfo"
	GroupCanonProcessingInfo GroupName = "Processing Info"
)
