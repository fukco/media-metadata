package mp4

import (
	"fmt"
	"github.com/fukco/media-meta-parser/media"
	"github.com/fukco/media-meta-parser/mp4/box"
	"os"
	"path/filepath"
	"reflect"
)

func ExtractMeta(file *os.File, ctx *media.Context) *media.Meta {
	boxes, _ := box.GetMetaBoxes(file)
	path, _ := filepath.Abs(file.Name())
	meta := &media.Meta{
		MediaPath: path,
		Context:   ctx,
		Items:     make([]interface{}, 0, 8),
		Temp:      make(map[string]interface{}),
	}
	for i := range boxes {
		funcValue := reflect.ValueOf(box.Map[boxes[i].Type]).MethodByName("GetMeta")
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
