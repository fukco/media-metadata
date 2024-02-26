package xavc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/fukco/media-meta-parser/manufacturer/sony/rtmd"
	"github.com/fukco/media-meta-parser/media"
	"io"
)

type RtmdDisp struct {
	// 自动白平衡模式
	AutoWhiteBalanceMode string
	// 白平衡光照预设
	LightingPreset string
	// 曝光模式
	ExposureMode string
	// 自动对焦范围
	AutoFocusSensingArea string
	// 快门
	ShutterSpeed string
	// 光圈
	Aperture string
	// ISO
	ISO string
	// 焦距
	FocalLength string
	// 35mm等效焦距
	FocalLength35mm string
	// 对焦距离
	FocusPosition string
	// 捕获Gamma
	CaptureGammaEquation string
	// 增益
	CameraMasterGainAdjustment string
}

func rtmd2rtmdDisp(rtmd *rtmd.RTMD) *RtmdDisp {
	rtmdDisp := &RtmdDisp{}
	rtmdDisp.AutoWhiteBalanceMode = rtmd.CameraUnitMetadata.AutoWhiteBalanceMode
	rtmdDisp.LightingPreset = rtmd.CameraUnitMetadata.LightingPreset
	rtmdDisp.ExposureMode = rtmd.CameraUnitMetadata.AutoExposureMode
	rtmdDisp.AutoFocusSensingArea = rtmd.CameraUnitMetadata.AutoFocusSensingAreaSetting
	rtmdDisp.ShutterSpeed = rtmd.CameraUnitMetadata.ShutterSpeedTime.String()
	rtmdDisp.Aperture = fmt.Sprintf("F%.1f", rtmd.LensUnitMetadata.IrisFNumber)
	rtmdDisp.ISO = fmt.Sprintf("ISO%d", rtmd.CameraUnitMetadata.ISOSensitivity)
	if rtmd.LensUnitMetadata.LensZoomPtr != nil {
		rtmdDisp.FocalLength = fmt.Sprintf("%.0fmm", *rtmd.LensUnitMetadata.LensZoomPtr*1000)
	}
	if rtmd.LensUnitMetadata.LensZoom35mmPtr != nil {
		rtmdDisp.FocalLength35mm = fmt.Sprintf("%.0fmm", *rtmd.LensUnitMetadata.LensZoom35mmPtr*1000)
	}
	if rtmd.LensUnitMetadata.FocusPositionFromImagePlane > 0 {
		rtmdDisp.FocusPosition = fmt.Sprintf("%.2fm", rtmd.LensUnitMetadata.FocusPositionFromImagePlane)
	}
	rtmdDisp.CaptureGammaEquation = fmt.Sprintf("%s/%s", rtmd.CameraUnitMetadata.ColorPrimaries, rtmd.CameraUnitMetadata.CaptureGammaEquation)
	if rtmd.CameraUnitMetadata.CameraMasterGainAdjustment == 0 {
		rtmdDisp.CameraMasterGainAdjustment = "0dB"
	} else {
		rtmdDisp.CameraMasterGainAdjustment = fmt.Sprintf("%.2fdB", rtmd.CameraUnitMetadata.CameraMasterGainAdjustment)
	}
	return rtmdDisp
}

func ReadMdat(r io.ReadSeeker) (*media.BoxInfo, error) {
	buf := bytes.NewBuffer([]byte{})
	offset := uint64(0)
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	for {
		if _, err := io.CopyN(buf, r, 8); err != nil {
			return nil, err
		}
		size := uint64(binary.BigEndian.Uint32(buf.Next(4)))
		boxType := string(buf.Next(4))
		headerSize := 8
		if size == 1 {
			if _, err := io.CopyN(buf, r, 8); err != nil {
				return nil, err
			}
			size = binary.BigEndian.Uint64(buf.Next(8))
			headerSize = 16
		}
		if boxType == "mdat" {
			boxInfo := &media.BoxInfo{
				Offset:     offset,
				Size:       size,
				HeaderSize: uint64(headerSize),
				Type:       media.StrToType("mdat"),
			}
			return boxInfo, nil
		} else {
			if _, err := r.Seek(int64(size)-int64(headerSize), io.SeekCurrent); err != nil {
				if err == io.EOF {
					return nil, err
				}
			}
		}
		offset += size
	}
}

func getHeaderAndBlockSize(r io.ReadSeeker, mdatBox *media.BoxInfo) (int, []byte, int64, error) {
	_, err := r.Seek(int64(mdatBox.Offset+mdatBox.HeaderSize), io.SeekStart)
	if err != nil {
		return 0, nil, 0, err
	}
	buf := bytes.NewBuffer([]byte{})
	if _, err := io.CopyN(buf, r, 1024); err != nil {
		return 0, nil, 0, err
	}
	n := bytes.Index(buf.Next(1024), []byte{0x00, 0x1c, 0x01, 0x00})
	if n > 0 {
		_, err := r.Seek(int64(n-1024), io.SeekCurrent)
		_ = err
	} else {
		return 0, nil, 0, errors.New("not found 001c0100")
	}
	if _, err := io.CopyN(buf, r, int64(1024*12)); err != nil {
		return 0, nil, 0, err
	}
	header := buf.Bytes()[:8]
	block := buf.Bytes()[28:]
	index := bytes.Index(block, header)
	if index > 0 {
		return n, header, int64(index + 28), nil
	} else {
		return 0, header, 0, errors.New("not a valid value")
	}
}

func ReadRtmdInBlock(r io.ReadSeeker, offset uint64) *RtmdDisp {
	if _, err := r.Seek(int64(offset), io.SeekStart); err != nil {
		return nil
	}
	if readRTMD, err := rtmd.ReadRTMD(r); err != nil {
		return nil
	} else {
		rtmdDisp := rtmd2rtmdDisp(readRTMD)
		return rtmdDisp
	}
}
