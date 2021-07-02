package quicktime

import (
	"fmt"
	"github.com/fukco/media-meta-parser/media"
	"os"
	"path/filepath"
	"reflect"
)

// AtomType is mpeg box type
type AtomType [4]byte

var atomMap = make(map[AtomType]interface{}, 64)

func AppendAtomMap(atomType AtomType, i interface{}) {
	atomMap[atomType] = i
}

func StrToAtomType(code string) AtomType {
	if len(code) != 4 {
		panic(fmt.Errorf("invalid box type id length: [%s]", code))
	}
	return AtomType{code[0], code[1], code[2], code[3]}
}

func ExtractMeta(file *os.File, ctx *media.Context) *media.Meta {
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)
	atoms, _ := GetMetaAtoms(file)
	path, _ := filepath.Abs(file.Name())
	meta := &media.Meta{
		MediaPath: path,
		Context:   ctx,
		Items:     make([]interface{}, 0, 8),
		Temp:      make(map[string]interface{}),
	}
	for i := range atoms {
		functionValue := reflect.ValueOf(atomMap[atoms[i].Type]).MethodByName("GetMeta")
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
