package meta

import (
	"github.com/fukco/media-metadata/internal/exif"
	"github.com/fukco/media-metadata/internal/manufacturer"
	"github.com/fukco/media-metadata/internal/manufacturer/nikon"
	"github.com/fukco/media-metadata/internal/manufacturer/panasonic"
	"github.com/fukco/media-metadata/internal/manufacturer/sony/nrtmd"
	"github.com/fukco/media-metadata/internal/manufacturer/sony/rtmd"
	"time"
)

type Metadata struct {
	FileName string
	FilePath string
	manufacturer.Manufacturer
	*Mp4Meta
	MetaItemKeyValues map[string]any
	*exif.ExifMeta
	*MakerMeta
}

type Mp4Meta struct {
	CreationTime     *time.Time
	ModificationTime *time.Time
	*VideoProfile
}

type VideoProfile struct {
	VideoAvgBitrate  string
	PixelAspectRatio string
}

type MakerMeta struct {
	*Atomos
	*Canon
	*Fujifilm
	*Nikon
	*Panasonic
	*Sony
}

type Atomos struct{}
type Canon struct{}
type Fujifilm struct{}
type Nikon struct {
	*nikon.NCTG
}
type Panasonic struct {
	*panasonic.ClipMain
}
type Sony struct {
	*nrtmd.NonRealTimeMeta
	*rtmd.RTMD
}
