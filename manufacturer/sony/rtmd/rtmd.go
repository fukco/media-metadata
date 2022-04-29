package rtmd

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/fukco/media-meta-parser/common"
	"io"
)

type metadataSetType string

const (
	lensUnitMetadataSet               metadataSetType = "LensUnitMetadata"
	cameraUnitMetadataSet             metadataSetType = "CameraUnitMetadata"
	userDefinedAcquisitionMetadataSet metadataSetType = "UserDefinedAcquisitionMetadata"
)

// RTMD Reference: EBU TECH-3349
type RTMD struct {
	Timecode                            *Timecode
	LensUnitMetadata                    *LensUnitMetadata
	CameraUnitMetadata                  *CameraUnitMetadata
	UserDefinedAcquisitionMetadataSlice []*UserDefinedAcquisitionMetadata
}

type LensUnitMetadata struct {
	IrisFNumber                 float64
	FocusPositionFromImagePlane float64
	FocusRingPosition           uint16
	LensZoom35mmPtr             *float64
	LensZoomPtr                 *float64
	ZoomRingPosition            uint16
	unKnownTags                 []*tag
}

type CameraUnitMetadata struct {
	AutoExposureMode                string
	ExposureIndexOfPhotoMeter       uint16
	AutoFocusSensingAreaSetting     string
	ImagerDimensionWidth            uint16
	ImagerDimensionHeight           uint16
	CaptureFrameRate                *common.Fraction
	ShutterSpeedAngle               float64
	ShutterSpeedTime                *common.Fraction
	CameraMasterGainAdjustment      float64
	ISOSensitivity                  uint16
	ElectricalExtenderMagnification float64
	AutoWhiteBalanceMode            string
	CaptureGammaEquation            string
	ColorPrimaries                  string
	CodingEquations                 string
	LightingPreset                  string
	unKnownTags                     []*tag
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

type UserDefinedAcquisitionMetadata struct {
	ID          []byte
	unKnownTags []*tag
}

type MetadataSetInfo struct {
	name       string
	bodyOffset int64
	bodyLength uint16
}

type tag struct {
	name tagName
	data []byte
}

func ReadRTMD(r io.ReadSeeker) (*RTMD, error) {
	rtmd := &RTMD{
		LensUnitMetadata:                    &LensUnitMetadata{unKnownTags: make([]*tag, 0, 16)},
		CameraUnitMetadata:                  &CameraUnitMetadata{unKnownTags: make([]*tag, 0, 16)},
		UserDefinedAcquisitionMetadataSlice: make([]*UserDefinedAcquisitionMetadata, 0, 16),
	}
	infos, err := readRTMDLayout(r, rtmd)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer([]byte{})
	for i := range infos {
		switch infos[i].name {
		case string(userDefinedAcquisitionMetadataSet):
			rtmd.UserDefinedAcquisitionMetadataSlice = append(rtmd.UserDefinedAcquisitionMetadataSlice,
				&UserDefinedAcquisitionMetadata{unKnownTags: make([]*tag, 0, 16)})
		}
		if _, err := r.Seek(infos[i].bodyOffset, io.SeekStart); err != nil {
			return nil, err
		}
		if _, err := io.CopyN(buf, r, int64(infos[i].bodyLength)); err != nil {
			return nil, err
		}
		content := buf.Next(int(infos[i].bodyLength))
		var n uint16 = 0
		for {
			if n >= infos[i].bodyLength {
				break
			}
			myTag := &tag{}
			myTag.name = tagName{content[n], content[n+1]}
			size := binary.BigEndian.Uint16(content[n+2 : n+4])

			tagData := make([]byte, size)
			copy(tagData, content[n+4:n+4+size])
			myTag.data = tagData
			n += size + 4
			if f, ok := rtmdMap[myTag.name]; ok {
				if err := f(myTag, rtmd); err != nil {
					return nil, err
				}
			} else {
				switch infos[i].name {
				case string(lensUnitMetadataSet):
					rtmd.LensUnitMetadata.unKnownTags = append(rtmd.LensUnitMetadata.unKnownTags, myTag)
				case string(cameraUnitMetadataSet):
					rtmd.CameraUnitMetadata.unKnownTags = append(rtmd.CameraUnitMetadata.unKnownTags, myTag)
				case string(userDefinedAcquisitionMetadataSet):
					unKnownTags := &rtmd.UserDefinedAcquisitionMetadataSlice[len(rtmd.UserDefinedAcquisitionMetadataSlice)-1].unKnownTags
					*unKnownTags = append(*unKnownTags, myTag)
				}
			}
		}
	}
	return rtmd, nil
}

func readRTMDLayout(r io.ReadSeeker, rtmd *RTMD) ([]*MetadataSetInfo, error) {
	buf := bytes.NewBuffer([]byte{})
	if _, err := io.CopyN(buf, r, 28); err != nil {
		return nil, err
	}
	frameHeader := buf.Next(28)
	rtmd.Timecode = &Timecode{
		Hour:  int(frameHeader[13]),
		Min:   int(frameHeader[14]),
		Sec:   int(frameHeader[15]),
		Frame: int(binary.BigEndian.Uint16(frameHeader[16:18])),
	}
	infos := make([]*MetadataSetInfo, 0, 4)
	for {
		if _, err := io.CopyN(buf, r, 20); err != nil {
			return nil, err
		}
		info := &MetadataSetInfo{}
		header := buf.Next(20)
		if bytes.Compare(header[:4], []byte{0x06, 0x0E, 0x2B, 0x34}) == 0 {
			info.bodyOffset, _ = r.Seek(0, io.SeekCurrent)
			info.bodyLength = binary.BigEndian.Uint16(header[18:])
			switch hex.EncodeToString(header[:16]) {
			case string(LensUnitMetadataHex):
				info.name = string(lensUnitMetadataSet)
			case string(CameraUnitMetadataHex):
				info.name = string(cameraUnitMetadataSet)
			case string(UserDefinedAcquisitionMetadataHex):
				info.name = string(userDefinedAcquisitionMetadataSet)
			}
			infos = append(infos, info)
		} else {
			break
		}
		if _, err := r.Seek(int64(info.bodyLength), io.SeekCurrent); err != nil {
			return nil, err
		}
	}
	return infos, nil
}
