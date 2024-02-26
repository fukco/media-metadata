package exif

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/fukco/media-meta-parser/manufacturer"
	"github.com/fukco/media-meta-parser/media"
	"strings"
)

func readByteOrder(data []byte) (binary.ByteOrder, error) {
	if bytes.Compare(data, []byte{0x4D, 0x4D, 0x00, 0x2A}) == 0 {
		return binary.BigEndian, nil
	} else if bytes.Compare(data, []byte{0x49, 0x49, 0x2A, 0x00}) == 0 {
		return binary.LittleEndian, nil
	}
	return nil, errors.New("illegal byte order")
}

func readIFD(data []byte, offset uint32, directoryType DirectoryType, exif *Base, meta *media.Meta, mfr manufacturer.Manufacturer) error {
	order := exif.TiffHeader.Order
	numOfDE := order.Uint16(data[offset : offset+2])
	entries := make([]*TiffEntry, 0, numOfDE)
	for i := 0; uint16(i) < numOfDE; i++ {
		start := offset + 2 + uint32(i*12)
		tagId := order.Uint16(data[start : start+2])
		dataType := DataType(order.Uint16(data[start+2 : start+4]))
		count := order.Uint32(data[start+4 : start+8])
		valueOrOffsetBytes := data[start+8 : start+12]
		entry := &TiffEntry{
			TagId: tagId,
			Type:  dataType,
			Count: count,
			Order: order,
		}
		if TypeSize[dataType]*count > 4 {
			valueOffset := order.Uint32(valueOrOffsetBytes)
			entry.ValOrOffset = valueOffset
			entry.Val = data[valueOffset : valueOffset+TypeSize[dataType]*count]
		} else {
			entry.Val = valueOrOffsetBytes
		}
		if tagId == 0x8769 {
			if err := readIFD(data, order.Uint32(valueOrOffsetBytes), ExifIFD, exif, meta, mfr); err != nil {
				return err
			}
		} else if tagId == 0x927c {
			if string(data[order.Uint32(valueOrOffsetBytes):order.Uint32(valueOrOffsetBytes)+9]) == string(Panasonic) {
				if err := readMakerNotes(data, order.Uint32(valueOrOffsetBytes)+12, exif, Panasonic); err != nil {
					return err
				}
			} else if string(data[order.Uint32(valueOrOffsetBytes):order.Uint32(valueOrOffsetBytes)+8]) == string(FUJIFILM) {
				if err := readMakerNotes(data[order.Uint32(valueOrOffsetBytes):], 12, exif, FUJIFILM); err != nil {
					return err
				}
			} else if mfr == manufacturer.CANON {
				if err := readMakerNotes(data, order.Uint32(valueOrOffsetBytes), exif, Canon); err != nil {
					return err
				}
			}
		} else {
			if err := entry.convertVals(); err != nil {
				return err
			}
			entries = append(entries, entry)
		}
	}
	nextIFDOffset := order.Uint32(data[offset+2+uint32(numOfDE*12) : offset+2+uint32(numOfDE*12)+4])
	directory := &TiffDirectory{
		Group:               []string{string(directoryType)},
		NumOfDE:             numOfDE,
		Entries:             entries,
		NextDirectoryOffset: nextIFDOffset,
	}
	exif.Directories = append(exif.Directories, directory)

	if nextIFDOffset != 0 {
		if err := readIFD(data, nextIFDOffset, IFD1, exif, meta, mfr); err != nil {
			return err
		}
	}
	return nil
}

func readMakerNotes(data []byte, offset uint32, exif *Base, maker Maker) error {
	order := exif.TiffHeader.Order
	numOfDE := order.Uint16(data[offset : offset+2])
	entries := make([]*TiffEntry, 0, numOfDE)
	for i := 0; uint16(i) < numOfDE; i++ {
		start := offset + 2 + uint32(i*12)
		tagId := order.Uint16(data[start : start+2])
		dataType := DataType(order.Uint16(data[start+2 : start+4]))
		count := order.Uint32(data[start+4 : start+8])
		valueOrOffsetBytes := data[start+8 : start+12]
		entry := &TiffEntry{
			TagId: tagId,
			Type:  dataType,
			Count: count,
			Order: order,
		}
		if TypeSize[dataType]*count > 4 {
			valueOffset := order.Uint32(valueOrOffsetBytes)
			entry.ValOrOffset = valueOffset
			entry.Val = data[valueOffset : valueOffset+TypeSize[dataType]*count]
		} else {
			entry.Val = valueOrOffsetBytes
		}
		if err := entry.convertVals(); err != nil {
			return err
		}
		entries = append(entries, entry)

	}

	directory := &TiffDirectory{
		Group:   []string{fmt.Sprintf("%s: %s", MakerIFD, maker)},
		NumOfDE: numOfDE,
		Entries: entries,
	}
	exif.MakerIFD = &MakerDirectory{Directories: []*TiffDirectory{directory}, Maker: maker}
	return nil
}

func exifEntryConvert2Tags(entry *TiffEntry, group Group) ([]*ExifTag, error) {
	tagDefinition := exifTagDefinitionMap[entry.TagId]
	var value interface{}
	format := entry.Format
	var valueStr string
	switch format {
	case IntVal:
		if entry.Count == 1 {
			value = entry.IntVals[0]
		} else {
			value = entry.IntVals
		}
	case FloatVal:
		if entry.Count == 1 {
			value = entry.FloatVals[0]
		} else {
			value = entry.FloatVals
		}
	case RatVal:
		if entry.Count == 1 {
			value = entry.RatVals[0]
		} else {
			value = entry.RatVals
		}
	case StringVal:
		value = strings.TrimSpace(strings.Join(entry.StrVals, " "))
	case UndefVal:
		value = entry.Val
	case OtherVal:
		value = entry.Val
	}
	exifTag := &ExifTag{
		ID:            entry.TagId,
		OriginalValue: entry.Val,
		Group:         group,
	}
	if tagDefinition == nil {
		exifTag.Undefined = true
		exifTag.Value = fmt.Sprintf("%v", value)
	} else {
		fn := tagDefinition.Fn
		if fn == nil {
			valueStr = fmt.Sprintf("%v", value)
		} else {
			valueStr = fn(value)
		}
		exifTag.Name = tagDefinition.Name
		exifTag.Value = valueStr
	}
	return []*ExifTag{exifTag}, nil
}

func makerNotesConvert2Tags(entry *TiffEntry, group Group, maker Maker) ([]*ExifTag, error) {
	var tagDefinition *TagDefinition
	switch maker {
	case Panasonic:
		tagDefinition = panasonicTagDefinitionMap[entry.TagId]
	case FUJIFILM:
		tagDefinition = fujiExifTagDefinitionMap[entry.TagId]
	case Canon:
		tagDefinition = canonExifTagDefinitionMap[entry.TagId]
	}
	var value interface{}
	format := entry.Format
	var valueStr string
	switch format {
	case IntVal:
		if entry.Count == 1 {
			value = entry.IntVals[0]
		} else {
			value = entry.IntVals
		}
	case FloatVal:
		if entry.Count == 1 {
			value = entry.FloatVals[0]
		} else {
			value = entry.FloatVals
		}
	case RatVal:
		if entry.Count == 1 {
			value = entry.RatVals[0]
		} else {
			value = entry.RatVals
		}
	case StringVal:
		value = strings.TrimSpace(strings.Join(entry.StrVals, " "))
	case UndefVal:
		value = entry.Val
	case OtherVal:
		value = entry.Val
	}
	if tagDefinition == nil {
		exifTag := &ExifTag{
			ID:            entry.TagId,
			OriginalValue: entry.Val,
			Group:         group,
			Undefined:     true,
			Value:         fmt.Sprintf("%v", value),
		}
		return []*ExifTag{exifTag}, nil
	} else {
		if tagDefinition.SubTagDefinition == nil {
			fn := tagDefinition.Fn
			if fn == nil {
				valueStr = fmt.Sprintf("%v", value)
			} else {
				valueStr = fn(value)
			}
			exifTag := &ExifTag{
				ID:            entry.TagId,
				Name:          tagDefinition.Name,
				OriginalValue: entry.Val,
				Value:         valueStr,
				Group:         group,
			}
			return []*ExifTag{exifTag}, nil
		} else {
			if tagDefinition.SubTagDefinition.tagDefinitionType == byIndex {
				data, ok := value.([]int64)
				tags := make([]*ExifTag, 0, len(data))
				if ok {
					group = append(group, tagDefinition.Name)
					subTagDefinitionMap := tagDefinition.SubTagDefinition.subTagDefinitionMap
					for key := range subTagDefinitionMap {
						subTagDefinition := subTagDefinitionMap[key]
						if subTagDefinition.Fn != nil {
							if key.(int) < len(data) {
								valueStr = subTagDefinition.Fn(data[key.(int)])
							} else {
								return nil, errors.New("invalid data or tag definition")
							}
						} else {
							valueStr = fmt.Sprintf("%v", data[key.(int)])
						}
						exifTag := &ExifTag{
							ID:            uint16(key.(int)),
							Name:          subTagDefinition.Name,
							OriginalValue: data[key.(int)],
							Value:         valueStr,
							Group:         group,
						}
						tags = append(tags, exifTag)
					}
				}
				return tags, nil
			}
		}
	}
	return nil, nil
}

func toExifMeta(base *Base) (*ExifMeta, error) {
	exifTags := make([]*ExifTag, 0, 8)
	if base.Directories != nil {
		for i := range base.Directories {
			directory := base.Directories[i]
			for j := range directory.Entries {
				entry := directory.Entries[j]
				if tags, err := exifEntryConvert2Tags(entry, directory.Group); err != nil {
					return nil, err
				} else {
					exifTags = append(exifTags, tags...)
				}
			}
		}
	}
	if base.MakerIFD != nil {
		for i := range base.MakerIFD.Directories {
			directory := base.MakerIFD.Directories[i]
			for j := range directory.Entries {
				entry := directory.Entries[j]
				if tags, err := makerNotesConvert2Tags(entry, directory.Group, base.MakerIFD.Maker); err != nil {
					return nil, err
				} else {
					exifTags = append(exifTags, tags...)
				}
			}
		}
	}
	exifMeta := &ExifMeta{Tags: make(map[string][]*ExifTag, 8)}
	for i := range exifTags {
		exifTag := exifTags[i]
		groupName := strings.Join(exifTag.Group, "/")
		groupTags, ok := exifMeta.Tags[groupName]
		if !ok {
			groupTags = make([]*ExifTag, 0, 64)
		}
		groupTags = append(groupTags, exifTag)
		exifMeta.Tags[groupName] = groupTags
	}
	return exifMeta, nil
}

func readExif(data []byte, ignoreHeader bool, meta *media.Meta, mfr manufacturer.Manufacturer) (*Base, error) {
	var order binary.ByteOrder = binary.LittleEndian
	var offset uint32
	if !ignoreHeader {
		readOrder, err := readByteOrder(data[:4])
		if err != nil {
			return nil, err
		}
		order = readOrder
		offset = order.Uint32(data[4:8])
	}
	header := &TiffHeader{order, offset}
	exif := &Base{TiffHeader: header}
	if err := readIFD(data, offset, IFD0, exif, meta, mfr); err != nil {
		return nil, err
	}
	return exif, nil
}

func Process(data []byte, ignoreHeader bool, meta *media.Meta, mfr manufacturer.Manufacturer) error {
	exif, _ := readExif(data, ignoreHeader, meta, mfr)
	exifMeta, err := toExifMeta(exif)
	if err != nil {
		return err
	}
	meta.Items = append(meta.Items, exifMeta)
	return nil
}

func ProcessJPEG(data []byte, meta *media.Meta, mfr manufacturer.Manufacturer) error {
	exifIdentifierCodeIndex := bytes.Index(data, []byte("Exif"))
	size := binary.BigEndian.Uint16(data[exifIdentifierCodeIndex-2 : exifIdentifierCodeIndex])
	return Process(data[exifIdentifierCodeIndex+6:exifIdentifierCodeIndex-2+int(size)], false, meta, mfr)
}
