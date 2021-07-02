package mp4

import (
	"fmt"
	"github.com/fukco/media-meta-parser/media"
	"os"
	"path/filepath"
	"reflect"
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

func isASCIIPrintableCharacter(c byte) bool {
	return c >= 0x20 && c <= 0x7e
}

func isPrintable(c byte) bool {
	return isASCIIPrintableCharacter(c) || c == 0xa9
}

var boxMap = make(map[BoxType]interface{}, 64)

func AppendBoxMap(boxType BoxType, i interface{}) {
	boxMap[boxType] = i
}

func ExtractMeta(file *os.File, ctx *media.Context) *media.Meta {
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)
	boxes, _ := GetMetaBoxes(file)
	path, _ := filepath.Abs(file.Name())
	meta := &media.Meta{
		MediaPath: path,
		Context:   ctx,
		Items:     make([]interface{}, 0, 8),
	}
	for i := range boxes {
		funcValue := reflect.ValueOf(boxMap[boxes[i].Type]).MethodByName("GetMeta")
		if funcValue.IsValid() {
			result := funcValue.Call([]reflect.Value{reflect.ValueOf(file), reflect.ValueOf(boxes[i]), reflect.ValueOf(ctx), reflect.ValueOf(meta)})
			err := result[0].Interface()
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return meta
}
