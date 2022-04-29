package main

import (
	"encoding/binary"
	"github.com/fukco/media-meta-parser/manufacturer"
	"github.com/fukco/media-meta-parser/media"
	"github.com/fukco/media-meta-parser/media/mp4"
	"github.com/fukco/media-meta-parser/media/quicktime"
	"os"
	"path/filepath"
	"strings"
)

func IsSupportExtension(file *os.File) bool {
	if strings.EqualFold(filepath.Ext(file.Name()), string(media.Mp4Extension)) ||
		strings.EqualFold(filepath.Ext(file.Name()), string(media.MovExtension)) ||
		strings.EqualFold(filepath.Ext(file.Name()), string(media.NRAWExtension)) {
		return true
	}
	return false
}

// IsSupportMediaFile check file is support media file
func IsSupportMediaFile(file *os.File) (bool, *media.Context, error) {
	if !IsSupportExtension(file) {
		return false, nil, nil
	}
	header := make([]byte, 8)
	ctx := &media.Context{}
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
	if string(majorBrand) == string(media.MP42) || string(majorBrand) == string(media.SONYXAVC) ||
		string(majorBrand) == string(media.QT) || string(majorBrand) == string(media.NIKO) {
		ctx.MajorBrand = string(majorBrand)
		ctx.CompatibleBrands = make([]string, 0, 4)
		compatibleBytes := body[8:]
		for i := 0; i < len(compatibleBytes)/4; i++ {
			ctx.CompatibleBrands = append(ctx.CompatibleBrands, string(compatibleBytes[4*i:4*i+4]))
		}
		if ctx.MajorBrand == string(media.QT) {
			ctx.MediaType = media.MOV
			for i := range ctx.CompatibleBrands {
				if ctx.CompatibleBrands[i] == "pana" {
					ctx.Manufacturer = manufacturer.PANASONIC
				} else if ctx.CompatibleBrands[i] == "niko" {
					ctx.Manufacturer = manufacturer.NIKON
				}
			}
		} else if ctx.MajorBrand == string(media.NIKO) {
			ctx.MediaType = media.MOV
			ctx.Manufacturer = manufacturer.NIKON
		} else {
			ctx.MediaType = media.MP4
			if ctx.MajorBrand == string(media.SONYXAVC) {
				ctx.Manufacturer = manufacturer.SONY
			} else if ctx.MajorBrand == string(media.NIKO) {
				ctx.Manufacturer = manufacturer.NIKON
			}
		}
		return true, ctx, nil
	}
	return false, nil, nil
}

func ExtractMeta(mediaFile *os.File, ctx *media.Context) *media.Meta {
	var meta *media.Meta
	if ctx.MediaType == media.MP4 {
		meta = mp4.ExtractMeta(mediaFile, ctx)
	} else if ctx.MediaType == media.MOV {
		meta = quicktime.ExtractMeta(mediaFile, ctx)
	}
	return meta
}
