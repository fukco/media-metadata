package quicktime

import (
	"fmt"
	"github.com/fukco/media-meta-parser/media"
	"github.com/fukco/media-meta-parser/quicktime/atom"
	"os"
	"path/filepath"
	"reflect"
)

func ExtractMeta(file *os.File, ctx *media.Context) *media.Meta {
	atoms, _ := atom.GetMetaAtoms(file)
	path, _ := filepath.Abs(file.Name())
	meta := &media.Meta{
		MediaPath: path,
		Context:   ctx,
		Items:     make([]interface{}, 0, 8),
		Temp:      make(map[string]interface{}),
	}
	for i := range atoms {
		functionValue := reflect.ValueOf(atom.Map[atoms[i].Type]).MethodByName("GetMeta")
		if functionValue.IsValid() {
			result := functionValue.Call([]reflect.Value{reflect.ValueOf(file), reflect.ValueOf(atoms[i]), reflect.ValueOf(ctx), reflect.ValueOf(meta)})
			err := result[0].Interface()
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return meta
}
