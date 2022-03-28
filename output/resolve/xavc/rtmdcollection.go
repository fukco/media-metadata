package xavc

import (
	"bytes"
	"fmt"
	"github.com/fukco/media-meta-parser/manufacturer/sony/rtmd"
	"io"
	"os"
)

type RtmdCollection struct {
	// 白平衡
	WhiteBalanceSlice []*FrameData
	// 曝光模式
	ExposureModeSlice []*FrameData
	// 自动对焦范围
	AutoFocusSensingAreaSlice []*FrameData
	// 快门
	ShutterSpeedSlice []*FrameData
	// 光圈
	ApertureSlice []*FrameData
	// ISO
	ISOSlice []*FrameData
	// 焦距
	FocalLengthSlice []*FrameData
	// 35mm等效焦距
	FocalLength35mmSlice []*FrameData
	// 对焦距离
	FocusPositionSlice []*FrameData
	// Gamma
	CaptureGammaEquationSlice []*FrameData
	// Gain
	CameraMasterGainAdjustmentSlice []*FrameData
	Offset                          int64
}

type FrameData struct {
	Frame int
	Data  string
}

func RtmdCollectionAppend(collection *RtmdCollection, index int, rtmd *rtmd.RTMD) {
	whiteBalance := "Unknown"
	if rtmd.CameraUnitMetadata.AutoWhiteBalanceMode == "Auto" && rtmd.CameraUnitMetadata.LightingPreset == "Reserved" {
		whiteBalance = "Auto"
	} else if rtmd.CameraUnitMetadata.AutoWhiteBalanceMode == "Preset" && rtmd.CameraUnitMetadata.LightingPreset == "SunLight" {
		whiteBalance = "SunLight"
	} else if rtmd.CameraUnitMetadata.AutoWhiteBalanceMode == "Preset" && rtmd.CameraUnitMetadata.LightingPreset == "Other" {
		whiteBalance = "Other"
	} else if rtmd.CameraUnitMetadata.AutoWhiteBalanceMode == "Preset" && rtmd.CameraUnitMetadata.LightingPreset == "Cloudy" {
		whiteBalance = "Cloudy"
	} else if rtmd.CameraUnitMetadata.AutoWhiteBalanceMode == "Preset" && rtmd.CameraUnitMetadata.LightingPreset == "Incandescent" {
		whiteBalance = "Incandescent"
	} else if rtmd.CameraUnitMetadata.AutoWhiteBalanceMode == "Preset" && rtmd.CameraUnitMetadata.LightingPreset == "Fluorescent" {
		whiteBalance = "Fluorescent"
	} else if rtmd.CameraUnitMetadata.AutoWhiteBalanceMode == "Preset" && rtmd.CameraUnitMetadata.LightingPreset == "Custom" {
		whiteBalance = "Custom"
	}
	if len(collection.WhiteBalanceSlice) < 1 {
		collection.WhiteBalanceSlice = append(collection.WhiteBalanceSlice, &FrameData{
			Frame: index,
			Data:  whiteBalance,
		})
	} else {
		last := collection.WhiteBalanceSlice[len(collection.WhiteBalanceSlice)-1]
		if last.Data != whiteBalance {
			collection.WhiteBalanceSlice = append(collection.WhiteBalanceSlice, &FrameData{
				Frame: index,
				Data:  whiteBalance,
			})
		}
	}
	if len(collection.ExposureModeSlice) < 1 {
		collection.ExposureModeSlice = append(collection.ExposureModeSlice, &FrameData{
			Frame: index,
			Data:  rtmd.CameraUnitMetadata.AutoExposureMode,
		})
	} else {
		last := collection.ExposureModeSlice[len(collection.ExposureModeSlice)-1]
		if last.Data != rtmd.CameraUnitMetadata.AutoExposureMode {
			collection.ExposureModeSlice = append(collection.ExposureModeSlice, &FrameData{
				Frame: index,
				Data:  rtmd.CameraUnitMetadata.AutoExposureMode,
			})
		}
	}
	if len(collection.AutoFocusSensingAreaSlice) < 1 {
		collection.AutoFocusSensingAreaSlice = append(collection.AutoFocusSensingAreaSlice, &FrameData{
			Frame: index,
			Data:  rtmd.CameraUnitMetadata.AutoFocusSensingAreaSetting,
		})
	} else {
		last := collection.AutoFocusSensingAreaSlice[len(collection.AutoFocusSensingAreaSlice)-1]
		if last.Data != rtmd.CameraUnitMetadata.AutoFocusSensingAreaSetting {
			collection.AutoFocusSensingAreaSlice = append(collection.AutoFocusSensingAreaSlice, &FrameData{
				Frame: index,
				Data:  rtmd.CameraUnitMetadata.AutoFocusSensingAreaSetting,
			})
		}
	}
	if len(collection.ShutterSpeedSlice) < 1 {
		collection.ShutterSpeedSlice = append(collection.ShutterSpeedSlice, &FrameData{
			Frame: index,
			Data:  rtmd.CameraUnitMetadata.ShutterSpeedTime.String(),
		})
	} else {
		last := collection.ShutterSpeedSlice[len(collection.ShutterSpeedSlice)-1]
		if last.Data != rtmd.CameraUnitMetadata.ShutterSpeedTime.String() {
			collection.ShutterSpeedSlice = append(collection.ShutterSpeedSlice, &FrameData{
				Frame: index,
				Data:  rtmd.CameraUnitMetadata.ShutterSpeedTime.String(),
			})
		}
	}
	if len(collection.ApertureSlice) < 1 {
		collection.ApertureSlice = append(collection.ApertureSlice, &FrameData{
			Frame: index,
			Data:  fmt.Sprintf("F%.1f", rtmd.LensUnitMetadata.IrisFNumber),
		})
	} else {
		last := collection.ApertureSlice[len(collection.ApertureSlice)-1]
		if last.Data != fmt.Sprintf("F%.1f", rtmd.LensUnitMetadata.IrisFNumber) {
			collection.ApertureSlice = append(collection.ApertureSlice, &FrameData{
				Frame: index,
				Data:  fmt.Sprintf("F%.1f", rtmd.LensUnitMetadata.IrisFNumber),
			})
		}
	}
	if len(collection.ISOSlice) < 1 {
		collection.ISOSlice = append(collection.ISOSlice, &FrameData{
			Frame: index,
			Data:  fmt.Sprintf("ISO%d", rtmd.CameraUnitMetadata.ISOSensitivity),
		})
	} else {
		last := collection.ISOSlice[len(collection.ISOSlice)-1]
		if last.Data != fmt.Sprintf("ISO%d", rtmd.CameraUnitMetadata.ISOSensitivity) {
			collection.ISOSlice = append(collection.ISOSlice, &FrameData{
				Frame: index,
				Data:  fmt.Sprintf("ISO%d", rtmd.CameraUnitMetadata.ISOSensitivity),
			})
		}
	}
	if rtmd.LensUnitMetadata.LensZoomPtr != nil {
		if len(collection.FocalLengthSlice) < 1 {
			collection.FocalLengthSlice = append(collection.FocalLengthSlice, &FrameData{
				Frame: index,
				Data:  fmt.Sprintf("%.0fmm", *rtmd.LensUnitMetadata.LensZoomPtr*1000),
			})
		} else {
			last := collection.FocalLengthSlice[len(collection.FocalLengthSlice)-1]
			if last.Data != fmt.Sprintf("%.0fmm", *rtmd.LensUnitMetadata.LensZoomPtr*1000) {
				collection.FocalLengthSlice = append(collection.FocalLengthSlice, &FrameData{
					Frame: index,
					Data:  fmt.Sprintf("%.0fmm", *rtmd.LensUnitMetadata.LensZoomPtr*1000),
				})
			}
		}
	}
	if rtmd.LensUnitMetadata.LensZoom35mmPtr != nil {
		if len(collection.FocalLength35mmSlice) < 1 {
			collection.FocalLength35mmSlice = append(collection.FocalLength35mmSlice, &FrameData{
				Frame: index,
				Data:  fmt.Sprintf("%.0fmm", *rtmd.LensUnitMetadata.LensZoom35mmPtr*1000),
			})
		} else {
			last := collection.FocalLength35mmSlice[len(collection.FocalLength35mmSlice)-1]
			if last.Data != fmt.Sprintf("%.0fmm", *rtmd.LensUnitMetadata.LensZoom35mmPtr*1000) {
				collection.FocalLength35mmSlice = append(collection.FocalLength35mmSlice, &FrameData{
					Frame: index,
					Data:  fmt.Sprintf("%.0fmm", *rtmd.LensUnitMetadata.LensZoom35mmPtr*1000),
				})
			}
		}
	}
	var focusPosition string
	if rtmd.LensUnitMetadata.FocusRingPosition == 0xffff {
		focusPosition = "+∞"
	} else {
		focusPosition = fmt.Sprintf("%.2fm", rtmd.LensUnitMetadata.FocusPositionFromImagePlane)
	}
	if len(collection.FocusPositionSlice) < 1 {
		collection.FocusPositionSlice = append(collection.FocusPositionSlice, &FrameData{
			Frame: index,
			Data:  focusPosition,
		})
	} else {
		last := collection.FocusPositionSlice[len(collection.FocusPositionSlice)-1]
		if last.Data != focusPosition {
			collection.FocusPositionSlice = append(collection.FocusPositionSlice, &FrameData{
				Frame: index,
				Data:  focusPosition,
			})
		}
	}
	if len(collection.CaptureGammaEquationSlice) < 1 {
		collection.CaptureGammaEquationSlice = append(collection.CaptureGammaEquationSlice, &FrameData{
			Frame: index,
			Data:  fmt.Sprintf("%s/%s", rtmd.CameraUnitMetadata.ColorPrimaries, rtmd.CameraUnitMetadata.CaptureGammaEquation),
		})
	} else {
		last := collection.CaptureGammaEquationSlice[len(collection.CaptureGammaEquationSlice)-1]
		if last.Data != fmt.Sprintf("%s/%s", rtmd.CameraUnitMetadata.ColorPrimaries, rtmd.CameraUnitMetadata.CaptureGammaEquation) {
			collection.CaptureGammaEquationSlice = append(collection.CaptureGammaEquationSlice, &FrameData{
				Frame: index,
				Data:  fmt.Sprintf("%s/%s", rtmd.CameraUnitMetadata.ColorPrimaries, rtmd.CameraUnitMetadata.CaptureGammaEquation),
			})
		}
	}
	gainAdjustment := "0dB"
	if rtmd.CameraUnitMetadata.CameraMasterGainAdjustment != 0 {
		gainAdjustment = fmt.Sprintf("%.2fdB", rtmd.CameraUnitMetadata.CameraMasterGainAdjustment)
	}
	if len(collection.CameraMasterGainAdjustmentSlice) < 1 {
		collection.CameraMasterGainAdjustmentSlice = append(collection.CameraMasterGainAdjustmentSlice, &FrameData{
			Frame: index,
			Data:  gainAdjustment,
		})
	} else {
		last := collection.CameraMasterGainAdjustmentSlice[len(collection.CameraMasterGainAdjustmentSlice)-1]
		if last.Data != gainAdjustment {
			collection.CameraMasterGainAdjustmentSlice = append(collection.CameraMasterGainAdjustmentSlice, &FrameData{
				Frame: index,
				Data:  gainAdjustment,
			})
		}
	}
}

func ReadRtmdSlice(f *os.File, start int, count int, offset int64) (*RtmdCollection, error) {
	mdat, _ := ReadMdat(f)
	firstFrameOffset, header, blockSize, _ := getHeaderAndBlockSize(f, mdat)
	rtmdCollection := &RtmdCollection{}
	if offset < 1 {
		//jump to the offset
		offset = int64(mdat.Offset+mdat.HeaderSize) + int64(firstFrameOffset)
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			return nil, err
		}
		buf := bytes.NewBuffer([]byte{})
		for i := 0; i < start; i++ {
			for {
				if _, err := io.CopyN(buf, f, blockSize); err != nil {
					return nil, err
				} else {
					if index := bytes.Index(buf.Next(int(blockSize)), header); index < 0 {
						continue
					} else {
						offset += blockSize
						if index > 0 {
							offset += int64(index)
							if _, err := f.Seek(int64(index), io.SeekCurrent); err != nil {
								return nil, err
							}
						}
						break
					}
				}
			}
		}
	} else {
		if uint64(offset) >= mdat.Offset+mdat.Size {
			return nil, nil
		}
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			return nil, err
		}
	}

	buf := bytes.NewBuffer([]byte{})
	for i := 0; i < count; i++ {
		for {
			if offset >= int64(mdat.Offset+mdat.Size) {
				break
			}
			if _, err := io.CopyN(buf, f, blockSize); err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			} else {
				offset += blockSize
			}
			block := buf.Next(int(blockSize))
			if index := bytes.Index(block, header); index < 0 {
				continue
			} else {
				if _, err := f.Seek(int64(index)-blockSize, io.SeekCurrent); err != nil {
					return nil, err
				} else {
					if readRTMD, err := rtmd.ReadRTMD(f); err != nil {
						return nil, err
					} else {
						RtmdCollectionAppend(rtmdCollection, start+i, readRTMD)
						offset += int64(index)
						if _, err := f.Seek(offset, io.SeekStart); err != nil {
							return nil, err
						}
						break
					}
				}
			}
		}
	}
	rtmdCollection.Offset = offset
	return rtmdCollection, nil
}
