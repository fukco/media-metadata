package rtmd

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/fukco/media-metadata/internal/common"
	"io"
	"math"
)

type metadataSetType uint8

const (
	lensUnitMetadataSet metadataSetType = iota + 1
	cameraUnitMetadataSet
	userDefinedAcquisitionMetadataSet
)

// RTMD Reference: EBU TECH-3349
type RTMD struct {
	Timecode                                   *Timecode
	LensUnitMetadata                           *LensUnitMetadata
	CameraUnitMetadata                         *CameraUnitMetadata
	UserDefinedAcquisitionMetadata             *UserDefinedAcquisitionMetadata
	userDefinedAcquisitionMetadataUnKnownSlice []*UserDefinedAcquisitionMetadataUnknown
}

type LensUnitMetadata struct {
	IrisFNumber                 float64
	FocusPositionFromImagePlane float64
	FocusRingPosition           uint16
	LensZoom35mmPtr             float64
	LensZoomPtr                 float64
	ZoomRingPosition            uint16
	unKnownTags                 []*tag
}

type CameraUnitMetadata struct {
	AutoExposureMode                string
	ExposureIndexOfPhotoMeter       uint16
	AutoFocusSensingAreaSetting     string
	ImagerDimensionWidth            uint16
	ImagerDimensionHeight           uint16
	CaptureFrameRate                *common.Rational
	ShutterSpeedAngle               float64
	ShutterSpeedTime                *common.Rational
	CameraMasterGainAdjustment      float64
	ISOSensitivity                  uint16
	ElectricalExtenderMagnification float64
	AutoWhiteBalanceMode            string
	CameraAttributes                string
	WhiteBalance                    uint16
	CaptureGammaEquation            string
	ColorPrimaries                  string
	CodingEquations                 string
	unKnownTags                     []*tag
}

type UserDefinedAcquisitionMetadata struct {
	CameraProcessDiscriminationCode string
	MonitoringDescriptions          string
	PreCDLTransform                 string
	LightingPreset                  string
	ImageStabilizerEnabled          bool
}

type Timecode struct {
	Hour  int
	Min   int
	Sec   int
	Frame int
}

func (t Timecode) String() string {
	if t.Frame < 10 {
		return fmt.Sprintf("%2d:%2d:%2d:%2d", t.Hour, t.Min, t.Sec, t.Frame)
	}
	return fmt.Sprintf("%2d:%2d:%2d:%d", t.Hour, t.Min, t.Sec, t.Frame)
}

type UserDefinedAcquisitionMetadataUnknown struct {
	id   []byte
	tags []*tag
}

type MetadataSetInfo struct {
	name       string
	bodyOffset int64
	bodyLength uint16
}

type rawData []byte

type tag struct {
	code code
	data rawData
}

type code uint16

func (t code) String() string {
	return fmt.Sprintf("%#04X", t)
}

func (data rawData) BigEndianUint16() uint16 {
	return binary.BigEndian.Uint16(data)
}

func (data rawData) BigEndianUint32() uint32 {
	return binary.BigEndian.Uint32(data)
}

func (data rawData) CommonDistanceFormat() float64 {
	u := binary.BigEndian.Uint16(data)
	e := int8(u>>8&0xf0) >> 4
	m := u & 0x0fff
	return float64(m) * math.Pow10(int(e))
}

func ReadRTMD(r io.ReadSeeker, sampleSize int, offset int) (*RTMD, error) {
	_, err := r.Seek(int64(offset), 0)
	if err != nil {
		return nil, err
	}
	n := 0
	buf := bytes.NewBuffer([]byte{})
	if _, err := io.CopyN(buf, r, 28); err != nil {
		return nil, err
	}
	frameHeader := buf.Next(28)
	n += 28
	if binary.BigEndian.Uint32(frameHeader[:4]) != 0x001c0100 {
		return nil, fmt.Errorf("not RTMD tag")
	}
	rtmd := &RTMD{
		Timecode: &Timecode{
			Hour:  int(frameHeader[13]),
			Min:   int(frameHeader[14]),
			Sec:   int(frameHeader[15]),
			Frame: int(binary.BigEndian.Uint16(frameHeader[16:18])),
		},
		LensUnitMetadata:                           &LensUnitMetadata{unKnownTags: make([]*tag, 0, 8)},
		CameraUnitMetadata:                         &CameraUnitMetadata{unKnownTags: make([]*tag, 0, 8)},
		UserDefinedAcquisitionMetadata:             &UserDefinedAcquisitionMetadata{},
		userDefinedAcquisitionMetadataUnKnownSlice: make([]*UserDefinedAcquisitionMetadataUnknown, 0, 16),
	}

	for {
		if sampleSize > 0 && n >= sampleSize {
			break
		}
		if _, err := io.CopyN(buf, r, 20); err != nil {
			return nil, err
		}
		header := buf.Next(20)
		n += 20
		if hex.EncodeToString(header[:4]) != "060e2b34" {
			break
		}
		var dataSetType metadataSetType
		switch hex.EncodeToString(header[:16]) {
		case LensUnitMetadataHex:
			dataSetType = lensUnitMetadataSet
		case CameraUnitMetadataHex:
			dataSetType = cameraUnitMetadataSet
		case UserDefinedAcquisitionMetadataHex:
			dataSetType = userDefinedAcquisitionMetadataSet
			rtmd.userDefinedAcquisitionMetadataUnKnownSlice = append(rtmd.userDefinedAcquisitionMetadataUnKnownSlice,
				&UserDefinedAcquisitionMetadataUnknown{tags: make([]*tag, 0, 16)})
		}
		length := binary.BigEndian.Uint16(header[18:])

		if _, err := io.CopyN(buf, r, int64(length)); err != nil {
			return nil, err
		}
		content := buf.Next(int(length))

		for i := 0; i < int(length); {
			myTag := &tag{}
			myTag.code = code(binary.BigEndian.Uint16(content[i : i+2]))
			size := int(binary.BigEndian.Uint16(content[i+2 : i+4]))

			tagData := make([]byte, size)
			copy(tagData, content[i+4:i+4+size])
			myTag.data = tagData
			i += size + 4
			switch dataSetType {
			case lensUnitMetadataSet:
				err := myTag.code.processLensUnitMetadata(rtmd, myTag.data)
				if errors.Is(err, ErrNotMatchedTag) {
					rtmd.LensUnitMetadata.unKnownTags = append(rtmd.LensUnitMetadata.unKnownTags, myTag)
				}
			case cameraUnitMetadataSet:
				err := myTag.code.processCameraUnitMetadata(rtmd, myTag.data)
				if errors.Is(err, ErrNotMatchedTag) {
					//fmt.Printf("%X %s\n", int(myTag.code), hex.EncodeToString(myTag.data))
					rtmd.CameraUnitMetadata.unKnownTags = append(rtmd.CameraUnitMetadata.unKnownTags, myTag)
				}
			case userDefinedAcquisitionMetadataSet:
				err := myTag.code.processUserDefinedAcquisitionMetadata(rtmd, myTag.data)
				if errors.Is(err, ErrNotMatchedTag) {
					//fmt.Printf("%X %s\n", int(myTag.code), hex.EncodeToString(myTag.data))
					unKnownTags := rtmd.userDefinedAcquisitionMetadataUnKnownSlice[len(rtmd.userDefinedAcquisitionMetadataUnKnownSlice)-1].tags
					unKnownTags = append(unKnownTags, myTag)
					rtmd.userDefinedAcquisitionMetadataUnKnownSlice[len(rtmd.userDefinedAcquisitionMetadataUnKnownSlice)-1].tags = unKnownTags
				}
			}
		}

	}

	return rtmd, nil
}
