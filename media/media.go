package media

import (
	"encoding/binary"
	"github.com/fukco/media-meta-parser/manufacturer"
	"os"
)

type Context struct {
	Manufacturer     manufacturer.Manufacturer
	//ModelName        string
	MediaType        Type
	MajorBrand       string
	CompatibleBrands []string
}

// IsSupportMediaFile check file is support media file
func IsSupportMediaFile(file *os.File) (bool, *Context, error) {
	header := make([]byte, 8)
	ctx := &Context{}
	if _, err := file.Read(header); err != nil {
		return false, nil, err
	}
	if string(header[4:8]) != "ftyp" {
		return false, nil, nil
	}
	size := binary.BigEndian.Uint32(header[:4])

	body := make([]byte, size-8)
	_, err := file.ReadAt(body, 8)
	if err != nil {
		return false, nil, err
	}
	majorBrand := body[:4]
	if string(majorBrand) == string(SONYXAVC) || string(majorBrand) == string(QT) {
		ctx.MajorBrand = string(majorBrand)
		ctx.CompatibleBrands = make([]string, 0, 4)
		compatibleBytes := body[8:]
		for i := 0; i < len(compatibleBytes)/4; i++ {
			ctx.CompatibleBrands = append(ctx.CompatibleBrands, string(compatibleBytes[4*i:4*i+4]))
		}
		if ctx.MajorBrand == string(QT) {
			ctx.MediaType = MOV
			for i := range ctx.CompatibleBrands {
				if ctx.CompatibleBrands[i] == "pana" {
					ctx.Manufacturer = manufacturer.PANASONIC
				}
			}
		} else {
			ctx.MediaType = MP4
			if ctx.MajorBrand == string(SONYXAVC) {
				ctx.Manufacturer = manufacturer.SONY
			}
		}

		return true, ctx, nil
	}
	return false, nil, nil
}
