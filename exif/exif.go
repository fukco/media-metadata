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
	IFD0     = DirectoryType(Group_IFD0)
	ExifIFD  = DirectoryType(Group_Exif)
	IFD1     = DirectoryType(Group_IFD1)
	MakerIFD = DirectoryType(Group_MakerNotes)
)

type GroupName string

const (
	Group_IFD0                  GroupName = "IFD0"
	Group_IFD1                  GroupName = "IFD1"
	Group_Exif                  GroupName = "ExifIFD"
	Group_MakerNotes            GroupName = "Maker"
	Group_Canon_Camera_Settings GroupName = "Canon CameraSettings"
	Group_Canon_Shot_Info       GroupName = "Canon ShotInfo"
	Group_Canon_Processing_Info GroupName = "Processing Info"
)
