package mp4

import (
	"fmt"
	"github.com/fukco/media-meta-parser/media"
	"os"
	"path/filepath"
	"strings"
)

// BoxType is mpeg box type
type BoxType [4]byte

func StrToBoxType(code string) BoxType {
	if len(code) != 4 {
		panic(fmt.Errorf("invalid box type id length: [%s]", code))
	}
	return BoxType{code[0], code[1], code[2], code[3]}
}

func (boxType BoxType) String() string {
	if isPrintable(boxType[0]) && isPrintable(boxType[1]) && isPrintable(boxType[2]) && isPrintable(boxType[3]) {
		s := string(boxType[:])
		s = strings.ReplaceAll(s, string([]byte{0xa9}), "(c)")
		return s
	}
	return fmt.Sprintf("0x%02x%02x%02x%02x", boxType[0], boxType[1], boxType[2], boxType[3])
}

func isASCII(c byte) bool {
	return c >= 0x20 && c <= 0x7e
}

func isPrintable(c byte) bool {
	return isASCII(c) || c == 0xa9
}

var boxMap = make(map[BoxType]interface{}, 64)

func AppendBoxMap(boxType BoxType, i interface{}) {
	boxMap[boxType] = i
}

func ExtractMeta(file *os.File) *media.Meta {
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)
	is, ctx, err := media.IsSupportMediaFile(file)
	if err != nil {
		fmt.Println(err)
		return nil
	} else if !is {
		return nil
	}
	boxes, _ := GetMetaBoxes(file)
	path, _ := filepath.Abs(file.Name())
	meta := &media.Meta{
		MediaPath: path,
		Context:   ctx,
		Items:     make([]interface{}, 0, 8),
	}
	for i := range boxes {
		switch boxes[i].Type {
		case StrToBoxType("xml "):
			xml := Xml{}
			if err := xml.getMeta(file, boxes[i], ctx, meta); err != nil {
				fmt.Println(err)
			}
		case StrToBoxType("mdat"):
			mdat := Mdat{}
			if err := mdat.getMeta(file, boxes[i], ctx, meta); err != nil {
				fmt.Println(err)
			}

		}
	}
	return meta
}
