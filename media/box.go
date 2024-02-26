package media

import (
	"io"
	"os"
	"reflect"
)

type IBox interface {
	GetMeta(m *Media, bi *BoxInfo, meta *Meta) error
}

var Mp4Map = make(map[BoxType]interface{}, 64)
var QuicktimeMap = make(map[BoxType]interface{}, 64)
var CommonMap = make(map[BoxType]interface{}, 64)

func AppendMp4Map(boxType BoxType, i interface{}) {
	Mp4Map[boxType] = i
}
func AppendQuicktimeMap(boxType BoxType, i interface{}) {
	QuicktimeMap[boxType] = i
}
func AppendCommonMap(boxType BoxType, i interface{}) {
	CommonMap[boxType] = i
}

// Box is ISO/IEC 14496-12 Box
type Box struct {
	Size     uint64
	Type     [4]byte
	UserType [16]byte
}

// FullBox is ISO/IEC 14496-12 FullBox
type FullBox struct {
	Box
	Version uint8
	Flags   [3]byte
}

type ContainerBox struct {
}

func (b *ContainerBox) getMeta(m *Media, bi *BoxInfo, meta *Meta) error {
	return nil
}

func isSupportedBox(bi *BoxInfo, boxMap map[BoxType]interface{}) bool {
	if _, ok := boxMap[bi.Type]; ok {
		return true
	}
	return false
}

func IsContainerBox(bi *BoxInfo, boxMap map[BoxType]interface{}) bool {
	if v, ok := boxMap[bi.Type]; ok {
		t := reflect.TypeOf(v)
		for i := 0; i < t.NumField(); i++ {
			if reflect.TypeOf(ContainerBox{}) == t.Field(i).Type {
				return true
			}
		}
	}
	return false
}

func GetMetaBoxes(file *os.File, boxMap map[BoxType]interface{}) ([]*BoxInfo, error) {
	boxInfos := make([]*BoxInfo, 0, 8)
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	for {
		// read 8 bytes
		if bi, err := ReadBoxInfo(file); err == nil {
			if isSupportedBox(bi, boxMap) {
				if IsContainerBox(bi, boxMap) {
					_, err = bi.SeekToPayload(file)
					if err != nil {
						return nil, err
					}
					if isFullBox(bi, boxMap) { //fullBox has 1 byte version and 3 bytes flags as box body
						_, err := file.Seek(4, io.SeekCurrent)
						if err != nil {
							return nil, err
						}
					}
					continue
				} else {
					boxInfos = append(boxInfos, bi)
				}
			}
			_, err := bi.SeekToEnd(file)
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
	return boxInfos, nil
}
