package sony

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
)

type metadataSetType string

const (
	lensUnitMetadataSet               metadataSetType = "LensUnitMetadata"
	cameraUnitMetadataSet             metadataSetType = "CameraUnitMetadata"
	userDefinedAcquisitionMetadataSet metadataSetType = "UserDefinedAcquisitionMetadata"
	undefined                         metadataSetType = "undefined"
)

// RTMD Reference: EBU TECH-3349
type RTMD struct {
	LensUnitMetadata                    *LensUnitMetadata
	CameraUnitMetadata                  *CameraUnitMetadata
	UserDefinedAcquisitionMetadataSlice []*UserDefinedAcquisitionMetadata
}

type LensUnitMetadata struct {
	IrisFNumber                 float64
	FocusPositionFromImagePlane float64
	FocusRingPosition           uint16
	LensZoom35mm                float64
	LensZoom                    float64
	ZoomRingPosition            uint16
	unKnownTags                 []*tag
}

type CameraUnitMetadata struct {
	AutoExposureMode                string
	ExposureIndexOfPhotoMeter       uint16
	AutoFocusSensingAreaSetting     string
	ImagerDimensionWidth            uint16
	ImagerDimensionHeight           uint16
	CaptureFrameRate                *Fraction
	ShutterSpeedAngle               float64
	ShutterSpeedTime                *Fraction
	CameraMasterGainAdjustment      float64
	ISOSensitivity                  uint16
	ElectricalExtenderMagnification float64
	AutoWhiteBalanceMode            string
	CaptureGammaEquation            string
	ColorPrimaries                  string
	CodingEquations                 string
	unKnownTags                     []*tag
}

type Fraction struct {
	// The numerator in the fraction, e.g. 2 in 2/3.
	Numerator int32
	// The value by which the numerator is divided, e.g. 3 in 2/3.
	Denominator int32
}

func (receiver Fraction) String() string {
	return fmt.Sprintf("%d/%d", receiver.Numerator, receiver.Denominator)
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
	infos, err := readRTMDLayout(r)
	if err != nil {
		return nil, err
	}

	structure := bytes.NewBuffer(make([]byte, 0, 1024))
	for i := range infos {
		switch infos[i].name {
		case string(undefined):
			continue
		case string(userDefinedAcquisitionMetadataSet):
			rtmd.UserDefinedAcquisitionMetadataSlice = append(rtmd.UserDefinedAcquisitionMetadataSlice,
				&UserDefinedAcquisitionMetadata{unKnownTags: make([]*tag, 0, 16)})
		}
		if _, err := r.Seek(infos[i].bodyOffset, io.SeekStart); err != nil {
			return nil, err
		}
		structure.Reset()
		if _, err := io.CopyN(structure, r, int64(infos[i].bodyLength)); err != nil {
			return nil, err
		}
		var n uint16 = 0
		for {
			if n >= infos[i].bodyLength {
				break
			}
			tag := &tag{}
			tag.name = tagName{structure.Bytes()[n], structure.Bytes()[n+1]}
			size := binary.BigEndian.Uint16(structure.Bytes()[n+2 : n+4])
			tag.data = structure.Bytes()[n+4 : n+4+size]
			n += size + 4
			if f, ok := rtmdMap[tag.name]; ok {
				if err := f(tag, rtmd); err != nil {
					return nil, err
				}
			} else {
				switch infos[i].name {
				case string(lensUnitMetadataSet):
					rtmd.LensUnitMetadata.unKnownTags = append(rtmd.LensUnitMetadata.unKnownTags, tag)
				case string(cameraUnitMetadataSet):
					rtmd.CameraUnitMetadata.unKnownTags = append(rtmd.CameraUnitMetadata.unKnownTags, tag)
				case string(userDefinedAcquisitionMetadataSet):
					unKnownTags := &rtmd.UserDefinedAcquisitionMetadataSlice[len(rtmd.UserDefinedAcquisitionMetadataSlice)-1].unKnownTags
					*unKnownTags = append(*unKnownTags, tag)
				}
			}
		}
	}
	return rtmd, nil
}

func readRTMDLayout(r io.ReadSeeker) ([]*MetadataSetInfo, error) {
	header := make([]byte, 2)
	if _, err := r.Read(header); err != nil {
		return nil, err
	}
	if _, err := r.Seek(int64(binary.BigEndian.Uint16(header))-2, io.SeekCurrent); err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(make([]byte, 0, 2))
	infos := make([]*MetadataSetInfo, 0, 4)
	for {
		buf.Reset()
		if _, err := io.CopyN(buf, r, 2); err != nil {
			return nil, err
		}
		info := &MetadataSetInfo{}
		infos = append(infos, info)
		name := buf.Bytes()
		if name[0] == 0x06 && name[1] == 0x0E {
			more := bytes.NewBuffer(make([]byte, 0, 18))
			if _, err := io.CopyN(more, r, 18); err != nil {
				return nil, err
			}
			all := append(buf.Bytes(), more.Bytes()...)
			info.bodyOffset, _ = r.Seek(0, io.SeekCurrent)
			info.bodyLength = binary.BigEndian.Uint16(all[18:])
			switch hex.EncodeToString(all[:16]) {
			case string(LensUnitMetadataHex):
				info.name = string(lensUnitMetadataSet)
			case string(CameraUnitMetadataHex):
				info.name = string(cameraUnitMetadataSet)
			case string(UserDefinedAcquisitionMetadataHex):
				info.name = string(userDefinedAcquisitionMetadataSet)
			}
		} else {
			info.name = string(undefined)
			sizeBytes := make([]byte, 2)
			if _, err := r.Read(sizeBytes); err != nil {
				return nil, err
			}
			info.bodyLength = binary.BigEndian.Uint16(sizeBytes)
			info.bodyOffset, _ = r.Seek(0, io.SeekCurrent)
			if name[0] == 0xFF && name[1] == 0xFF {
				break
			}
		}
		if _, err := r.Seek(int64(info.bodyLength), io.SeekCurrent); err != nil {
			return nil, err
		}
	}
	return infos, nil
}
