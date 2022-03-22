package atom

import (
	"fmt"
	"github.com/fukco/media-meta-parser/media"
	"io"
	"os"
	"reflect"
)

// Type is mpeg box type
type Type [4]byte

var Map = make(map[Type]interface{}, 64)

func appendAtomMap(atomType Type, i interface{}) {
	Map[atomType] = i
}

func strToType(code string) Type {
	if len(code) != 4 {
		panic(fmt.Errorf("invalid box type id length: [%s]", code))
	}
	return Type{code[0], code[1], code[2], code[3]}
}

type IAtom interface {
	GetMeta(r io.ReadSeeker, ai *Info, ctx *media.Context, meta *media.Meta) error
}

type Atom struct {
	Size     uint64
	Type     [4]byte
	UserType [16]byte
}

type FullAtom struct {
	Atom
	Version uint8
	Flags   [3]byte
}

type ContainerAtom struct {
}

func (b *ContainerAtom) getMeta(io.ReadSeeker, *Info, *media.Context) error {
	return nil
}

func isSupportedAtom(bi *Info) bool {
	if _, ok := Map[bi.Type]; ok {
		return true
	}
	return false
}

func IsContainerAtom(bi *Info) bool {
	if v, ok := Map[bi.Type]; ok {
		t := reflect.TypeOf(v)
		for i := 0; i < t.NumField(); i++ {
			if reflect.TypeOf(ContainerAtom{}) == t.Field(i).Type {
				return true
			}
		}
	}
	return false
}

func GetMetaAtoms(file *os.File) ([]*Info, error) {
	atomInfos := make([]*Info, 0, 8)
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	for {
		// read 8 bytes
		if ai, err := ReadAtomInfo(file); err == nil {
			if isSupportedAtom(ai) {
				if IsContainerAtom(ai) {
					_, err = ai.SeekToPayload(file)
					if err != nil {
						return nil, err
					}
					if IsFullAtom(ai) { //fullAtom has 1 byte version and 3 bytes flags as atom body
						_, err := file.Seek(4, io.SeekCurrent)
						if err != nil {
							return nil, err
						}
					}
					continue
				} else {
					atomInfos = append(atomInfos, ai)
				}
			}
			_, err := ai.SeekToEnd(file)
			if err != nil {
				return nil, err
			}
		} else {
			if err == io.EOF {
				break
			}
			return nil, err
		}

	}
	return atomInfos, nil
}
