package media

import (
	"encoding/binary"
	"fmt"
	"github.com/fukco/media-meta-parser/manufacturer"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type Media struct {
	*os.File
	Type
	Ftyp *Ftyp
	manufacturer.Manufacturer
}

func NewMedia(file *os.File) *Media {
	m := &Media{
		File: file,
		Ftyp: &Ftyp{},
	}

	if t := getMediaType(file); t == "" {
		fmt.Printf("%s not support filename extension\n", file.Name())
		return nil
	} else {
		m.Type = t
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil
	}

	header := make([]byte, 8)
	if _, err := file.Read(header); err != nil {
		return nil
	}
	if string(header[4:8]) != "ftyp" {
		return nil
	}

	size := binary.BigEndian.Uint32(header[:4])

	m.Ftyp.Size = size

	body := make([]byte, size-8)
	_, err := file.ReadAt(body, 8)
	if err != nil {
		return nil
	}
	majorBrand := body[:4]
	if string(majorBrand) != string(MP42) && string(majorBrand) != string(SONYXAVC) &&
		string(majorBrand) != string(QT) && string(majorBrand) != string(NIKO) {
		return nil
	}
	m.Ftyp.MajorBrand = string(majorBrand)
	m.Ftyp.MinorVersion = binary.BigEndian.Uint32(body[4:8])
	m.Ftyp.CompatibleBrands = make([]string, 0, 4)
	compatibleBytes := body[8:]
	for i := 0; i < len(compatibleBytes)/4; i++ {
		m.Ftyp.CompatibleBrands = append(m.Ftyp.CompatibleBrands, string(compatibleBytes[4*i:4*i+4]))
	}

	m.Manufacturer = getManufacturer(m.Ftyp)

	return m
}

func (m *Media) Close() {
	if m != nil && m.File != nil {
		_ = m.File.Close()
	}
}

func getMediaType(file *os.File) Type {
	if strings.EqualFold(filepath.Ext(file.Name()), string(Mp4Extension)) ||
		strings.EqualFold(filepath.Ext(file.Name()), string(NRAWExtension)) {
		return MP4
	}
	if strings.EqualFold(filepath.Ext(file.Name()), string(MovExtension)) {
		return MOV
	}
	return ""
}

func getManufacturer(ftyp *Ftyp) manufacturer.Manufacturer {
	if ftyp.MajorBrand == string(SONYXAVC) {
		return manufacturer.SONY
	} else if ftyp.MajorBrand == string(NIKO) {
		return manufacturer.NIKON
	}
	var m manufacturer.Manufacturer
	for i := range ftyp.CompatibleBrands {
		if ftyp.CompatibleBrands[i] == string(PANA) {
			m = manufacturer.PANASONIC
		} else if ftyp.CompatibleBrands[i] == string(NIKO) {
			m = manufacturer.NIKON
		}
	}
	return m
}

// BoxType is mpeg box type
type BoxType [4]byte

func StrToType(code string) BoxType {
	if len(code) != 4 {
		panic(fmt.Errorf("invalid box type id length: [%s]", code))
	}
	return BoxType{code[0], code[1], code[2], code[3]}
}

func isASCIIPrintableCharacter(c byte) bool {
	return c >= 0x20 && c <= 0x7e
}

func isPrintable(c byte) bool {
	return isASCIIPrintableCharacter(c) || c == 0xa9
}

func (boxType BoxType) String() string {
	if isPrintable(boxType[0]) && isPrintable(boxType[1]) && isPrintable(boxType[2]) && isPrintable(boxType[3]) {
		s := string(boxType[:])
		s = strings.ReplaceAll(s, string([]byte{0xa9}), "(c)")
		return s
	}
	return fmt.Sprintf("0x%02x%02x%02x%02x", boxType[0], boxType[1], boxType[2], boxType[3])
}

//
//func IsSupportExtension(file *os.File) bool {
//	if strings.EqualFold(filepath.Ext(file.Name()), string(Mp4Extension)) ||
//		strings.EqualFold(filepath.Ext(file.Name()), string(MovExtension)) ||
//		strings.EqualFold(filepath.Ext(file.Name()), string(NRAWExtension)) {
//		return true
//	}
//	return false
//}
//
//// IsSupportMediaFile check file is support media file
//func IsSupportMediaFile(file *os.File) (bool, *Context, error) {
//	if !IsSupportExtension(file) {
//		return false, nil, nil
//	}
//	header := make([]byte, 8)
//	ctx := &Context{}
//	if _, err := file.Read(header); err != nil {
//		return false, nil, err
//	}
//	if string(header[4:8]) != "ftyp" {
//		return false, nil, nil
//	}
//	size := binary.BigEndian.Uint32(header[:4])
//
//	body := make([]byte, size-8)
//	_, err := file.ReadAt(body, 8)
//	if err != nil {
//		return false, nil, err
//	}
//	majorBrand := body[:4]
//	if string(majorBrand) == string(MP42) || string(majorBrand) == string(SONYXAVC) ||
//		string(majorBrand) == string(QT) || string(majorBrand) == string(NIKO) {
//		ctx.MajorBrand = string(majorBrand)
//		ctx.CompatibleBrands = make([]string, 0, 4)
//		compatibleBytes := body[8:]
//		for i := 0; i < len(compatibleBytes)/4; i++ {
//			ctx.CompatibleBrands = append(ctx.CompatibleBrands, string(compatibleBytes[4*i:4*i+4]))
//		}
//		if ctx.MajorBrand == string(QT) {
//			ctx.MediaType = MOV
//			for i := range ctx.CompatibleBrands {
//				if ctx.CompatibleBrands[i] == "pana" {
//					ctx.Manufacturer = manufacturer.PANASONIC
//				} else if ctx.CompatibleBrands[i] == "niko" {
//					ctx.Manufacturer = manufacturer.NIKON
//				}
//			}
//		} else if ctx.MajorBrand == string(NIKO) {
//			ctx.MediaType = MOV
//			ctx.Manufacturer = manufacturer.NIKON
//		} else {
//			ctx.MediaType = MP4
//			if ctx.MajorBrand == string(SONYXAVC) {
//				ctx.Manufacturer = manufacturer.SONY
//			} else if ctx.MajorBrand == string(NIKO) {
//				ctx.Manufacturer = manufacturer.NIKON
//			}
//		}
//		return true, ctx, nil
//	}
//	return false, nil, nil
//}

//func ExtractMeta(mediaFile *os.File, ctx *Context) *Meta {
//	var meta *Meta
//	if ctx.MediaType == MP4 {
//		meta = mp4.ExtractMeta(mediaFile, ctx)
//	} else if ctx.MediaType == MOV {
//		meta = quicktime.ExtractMeta(mediaFile, ctx)
//	}
//	return meta
//}

func (m *Media) ExtractMeta() *Meta {
	var boxMap map[BoxType]interface{}
	if m.Type == MP4 {
		boxMap = Mp4Map
	} else {
		boxMap = QuicktimeMap
	}
	boxes, _ := GetMetaBoxes(m.File, boxMap)
	meta := &Meta{
		//Context: ctx,
		Items: make([]interface{}, 0, 8),
		Temp:  make(map[string]interface{}),
	}
	for i := range boxes {
		funcValue := reflect.ValueOf(boxMap[boxes[i].Type]).MethodByName("GetMeta")
		if funcValue.IsValid() {
			result := funcValue.Call([]reflect.Value{reflect.ValueOf(m), reflect.ValueOf(boxes[i]), reflect.ValueOf(meta)})
			err := result[0].Interface()
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return meta
}
