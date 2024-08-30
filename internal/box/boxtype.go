package box

import (
	"encoding/binary"
	"fmt"
	"github.com/fukco/media-metadata/internal/manufacturer"
	"strings"
)

type BoxType uint32

const (
	FileTypeBox          BoxType = 0x66747970 //"ftyp"
	MovieBox             BoxType = 0x6D6F6F76 //"moov"
	MovieHeaderBox       BoxType = 0x6D766864 //"mvhd"
	MetaBox              BoxType = 0x6D657461 //"meta"
	MetadataItemKeysAtom BoxType = 0x6B657973 //"keys"
	MetadataItemListAtom BoxType = 0x696C7374 //"ilst"
	XMLBox               BoxType = 0x786D6C20 //"xml "
	MediaDataBox         BoxType = 0x6D646174 //"mdat"
	UserDataBox          BoxType = 0x75647461 //"udta"
	UUIDExtensionBox     BoxType = 0x75756964 //"uuid"
	PanasonicPANABox     BoxType = 0x50414E41 //"PANA"
	NikonNCDTBox         BoxType = 0x4E434454 //"NCDT"
	NikonNCTGBox         BoxType = 0x4E435447 //"NCTG"
	FujiMVTGBox          BoxType = 0x4D565447 //"MVTG"
	CanonCNTH            BoxType = 0x434E5448 //"CNTH"
	CanonCNDA            BoxType = 0x434E4441 //"CNDA"
	VideoProfile         BoxType = 0x56505246 //"VPRF"
	TrackBox             BoxType = 0x7472616B //"trak"
	MediaBox             BoxType = 0x6D646961 //"mdia"
	HandlerReferenceBox  BoxType = 0x68646C72 //"hdlr"
	MediaInformationBox  BoxType = 0x6D696E66 //"minf"
	SampleTableBox       BoxType = 0x7374626C //"stbl"
	SampleSizeBox        BoxType = 0x7374737A //"stsz"
	ChunkOffsetBox       BoxType = 0x7374636F //"stco"
)

type UserType [16]byte

func isASCIIPrintableCharacter(c byte) bool {
	return c >= 0x20 && c <= 0x7e
}

func isPrintable(c byte) bool {
	return isASCIIPrintableCharacter(c) || c == 0xa9
}

func StrToBoxType(code string) BoxType {
	if len(code) != 4 {
		panic(fmt.Errorf("invalid box type id length: [%s]", code))
	}
	return BoxType(binary.BigEndian.Uint32([]byte{code[0], code[1], code[2], code[3]}))
}

func (boxType BoxType) String() string {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(boxType))
	if isPrintable(b[0]) && isPrintable(b[1]) && isPrintable(b[2]) && isPrintable(b[3]) {
		s := string(b)
		s = strings.ReplaceAll(s, string([]byte{0xa9}), "(c)")
		return s
	}
	return fmt.Sprintf("%#08X", boxType)
}

func (boxType BoxType) getBoxDef() *BoxDef {
	if def, ok := BoxMap[boxType]; ok {
		return &def
	}
	return nil
}

func (boxType BoxType) GetManufacturer() manufacturer.Manufacturer {
	switch boxType {
	case PanasonicPANABox:
		return manufacturer.PANASONIC
	case FujiMVTGBox:
		return manufacturer.FUJIFILM
	case NikonNCDTBox:
		return manufacturer.NIKON
	default:
		return manufacturer.Unknown
	}
}

func (userType UserType) getBoxDef() *BoxDef {
	if def, ok := UUIDBoxMap[userType]; ok {
		return &def
	}
	return nil
}
