package quicktime

import (
	"github.com/fukco/media-meta-parser/media"
	"io"
	"os"
	"reflect"
)

type IAtom interface {
	GetMeta(r io.ReadSeeker, ai *AtomInfo, ctx *media.Context, meta *media.Meta) error
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

func (b *ContainerAtom) getMeta(io.ReadSeeker, *AtomInfo, *media.Context) error {
	return nil
}

func isSupportedAtom(bi *AtomInfo) bool {
	if _, ok := atomMap[bi.Type]; ok {
		return true
	}
	return false
}

func IsContainerAtom(bi *AtomInfo) bool {
	if v, ok := atomMap[bi.Type]; ok {
		t := reflect.TypeOf(v)
		for i := 0; i < t.NumField(); i++ {
			if reflect.TypeOf(ContainerAtom{}) == t.Field(i).Type {
				return true
			}
		}
	}
	return false
}

func GetMetaAtoms(file *os.File) ([]*AtomInfo, error) {
	atomInfos := make([]*AtomInfo, 0, 8)
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
					if isFullAtom(ai) { //fullAtom has 1 byte version and 3 bytes flags as atom body
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
