package xavc

import (
	"errors"
	"fmt"
	"github.com/fukco/media-metadata/internal/manufacturer/sony/rtmd"
	"io"
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
	// ImageStabilizer
	ImageStabilizerSlice []*FrameData
}

type FrameData struct {
	Frame int
	Data  string
}

func appendFrameData(slice []*FrameData, input *FrameData) []*FrameData {
	if len(slice) == 0 {
		slice = append(slice, input)
	}
	if input.Data != slice[len(slice)-1].Data {
		slice = append(slice, input)
	}
	return slice
}

func RtmdCollectionAppend(collection *RtmdCollection, index int, rtmd *rtmd.RTMD) {
	whiteBalance := "Unknown"
	if rtmd.CameraUnitMetadata.AutoWhiteBalanceMode == "Auto" && rtmd.UserDefinedAcquisitionMetadata.LightingPreset == "Reserved" {
		whiteBalance = "Auto"
	} else if rtmd.CameraUnitMetadata.AutoWhiteBalanceMode == "Preset" && rtmd.UserDefinedAcquisitionMetadata.LightingPreset == "SunLight" {
		whiteBalance = "SunLight"
	} else if rtmd.CameraUnitMetadata.AutoWhiteBalanceMode == "Preset" && rtmd.UserDefinedAcquisitionMetadata.LightingPreset == "Other" {
		whiteBalance = "Other"
	} else if rtmd.CameraUnitMetadata.AutoWhiteBalanceMode == "Preset" && rtmd.UserDefinedAcquisitionMetadata.LightingPreset == "Cloudy" {
		whiteBalance = "Cloudy"
	} else if rtmd.CameraUnitMetadata.AutoWhiteBalanceMode == "Preset" && rtmd.UserDefinedAcquisitionMetadata.LightingPreset == "Incandescent" {
		whiteBalance = "Incandescent"
	} else if rtmd.CameraUnitMetadata.AutoWhiteBalanceMode == "Preset" && rtmd.UserDefinedAcquisitionMetadata.LightingPreset == "Fluorescent" {
		whiteBalance = "Fluorescent"
	} else if rtmd.CameraUnitMetadata.AutoWhiteBalanceMode == "Preset" && rtmd.UserDefinedAcquisitionMetadata.LightingPreset == "Custom" {
		whiteBalance = "Custom"
	}
	collection.WhiteBalanceSlice = appendFrameData(collection.WhiteBalanceSlice, &FrameData{
		Frame: index,
		Data:  whiteBalance,
	})
	collection.ExposureModeSlice = appendFrameData(collection.ExposureModeSlice, &FrameData{
		Frame: index,
		Data:  rtmd.CameraUnitMetadata.AutoExposureMode,
	})
	collection.AutoFocusSensingAreaSlice = appendFrameData(collection.AutoFocusSensingAreaSlice, &FrameData{
		Frame: index,
		Data:  rtmd.CameraUnitMetadata.AutoFocusSensingAreaSetting,
	})
	collection.ShutterSpeedSlice = appendFrameData(collection.ShutterSpeedSlice, &FrameData{
		Frame: index,
		Data:  rtmd.CameraUnitMetadata.ShutterSpeedTime.String(),
	})
	collection.ApertureSlice = appendFrameData(collection.ApertureSlice, &FrameData{
		Frame: index,
		Data:  fmt.Sprintf("F%.1f", rtmd.LensUnitMetadata.IrisFNumber),
	})
	collection.ISOSlice = appendFrameData(collection.ISOSlice, &FrameData{
		Frame: index,
		Data:  fmt.Sprintf("ISO%d", rtmd.CameraUnitMetadata.ISOSensitivity),
	})
	collection.FocalLengthSlice = appendFrameData(collection.FocalLengthSlice, &FrameData{
		Frame: index,
		Data:  fmt.Sprintf("%.0fmm", rtmd.LensUnitMetadata.LensZoomPtr*1000),
	})
	collection.FocalLength35mmSlice = appendFrameData(collection.FocalLengthSlice, &FrameData{
		Frame: index,
		Data:  fmt.Sprintf("%.0fmm", rtmd.LensUnitMetadata.LensZoom35mmPtr*1000),
	})
	var focusPosition string
	if rtmd.LensUnitMetadata.FocusPositionFromImagePlane > 0 {
		focusPosition = fmt.Sprintf("%.2fm", rtmd.LensUnitMetadata.FocusPositionFromImagePlane)
	}
	collection.FocusPositionSlice = appendFrameData(collection.FocusPositionSlice, &FrameData{
		Frame: index,
		Data:  focusPosition,
	})
	collection.CaptureGammaEquationSlice = appendFrameData(collection.CaptureGammaEquationSlice, &FrameData{
		Frame: index,
		Data:  fmt.Sprintf("%s/%s", rtmd.CameraUnitMetadata.ColorPrimaries, rtmd.CameraUnitMetadata.CaptureGammaEquation),
	})
	gainAdjustment := "0dB"
	if rtmd.CameraUnitMetadata.CameraMasterGainAdjustment != 0 {
		gainAdjustment = fmt.Sprintf("%.2fdB", rtmd.CameraUnitMetadata.CameraMasterGainAdjustment)
	}
	collection.CameraMasterGainAdjustmentSlice = appendFrameData(collection.CameraMasterGainAdjustmentSlice, &FrameData{
		Frame: index,
		Data:  gainAdjustment,
	})
	imageStabilizer := "Disabled"
	if rtmd.UserDefinedAcquisitionMetadata.ImageStabilizerEnabled {
		imageStabilizer = "Enabled"
	}
	collection.ImageStabilizerSlice = appendFrameData(collection.ImageStabilizerSlice, &FrameData{
		Frame: index,
		Data:  imageStabilizer,
	})
}

func ReadRtmdSlice(r io.ReadSeeker, start int, count int) (*RtmdCollection, error) {
	info, err := getMetaSampleInfo(r)
	if err != nil {
		return nil, err
	}
	if start >= int(info.SampleCount) || count <= 0 {
		return nil, errors.New("invalid input")
	}
	if start+count > int(info.SampleCount) {
		count = int(info.SampleCount) - start
	}

	rtmdCollection := &RtmdCollection{}
	for i := start; i < start+count; i++ {
		offset := int(info.ChunkOffsets[i/int(info.SamplesPerChunk)]) + info.Size*(i%int(info.SamplesPerChunk))
		if rtmd, err := rtmd.ReadRTMD(r, info.Size, offset); err != nil {
			return nil, err
		} else {
			RtmdCollectionAppend(rtmdCollection, start+i, rtmd)
		}
	}
	return rtmdCollection, nil
}
