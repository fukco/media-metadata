package xavc

import (
	"fmt"
	"github.com/fukco/media-metadata/internal/box"
	"github.com/fukco/media-metadata/internal/manufacturer/sony/rtmd"
	"io"
)

type RtmdDisp struct {
	// 自动白平衡模式
	AutoWhiteBalanceMode string
	// 白平衡
	WhiteBalanceKelvin string
	// 白平衡光照预设
	LightingPreset string
	// 曝光模式
	ExposureMode string
	// 自动对焦范围
	AutoFocusSensingArea string
	// 快门角度
	ShutterAngle string
	// 快门速度
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
	rtmdDisp.WhiteBalanceKelvin = fmt.Sprintf("%d", rtmd.CameraUnitMetadata.WhiteBalance)
	rtmdDisp.LightingPreset = rtmd.UserDefinedAcquisitionMetadata.LightingPreset
	rtmdDisp.ExposureMode = rtmd.CameraUnitMetadata.AutoExposureMode
	rtmdDisp.AutoFocusSensingArea = rtmd.CameraUnitMetadata.AutoFocusSensingAreaSetting
	rtmdDisp.ShutterSpeed = rtmd.CameraUnitMetadata.ShutterSpeedTime.String()
	rtmdDisp.Aperture = fmt.Sprintf("F%.1f", rtmd.LensUnitMetadata.IrisFNumber)
	rtmdDisp.ISO = fmt.Sprintf("ISO%d", rtmd.CameraUnitMetadata.ISOSensitivity)
	if rtmd.CameraUnitMetadata.ShutterSpeedAngle > 0 {
		rtmdDisp.ShutterAngle = fmt.Sprintf("%.1f°", rtmd.CameraUnitMetadata.ShutterSpeedAngle)
	}
	if rtmd.LensUnitMetadata.LensZoomPtr != 0 {
		rtmdDisp.FocalLength = fmt.Sprintf("%.0fmm", rtmd.LensUnitMetadata.LensZoomPtr*1000)
	}
	if rtmd.LensUnitMetadata.LensZoom35mmPtr != 0 {
		rtmdDisp.FocalLength35mm = fmt.Sprintf("%.0fmm", rtmd.LensUnitMetadata.LensZoom35mmPtr*1000)
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

func searchMetaTrackMedia(boxDetails []*box.BoxDetail) *box.BoxDetail {
	for _, detail := range boxDetails {
		if isMetaTrackMedia(detail) {
			return detail
		}
		if len(detail.Children) > 0 {
			result := searchMetaTrackMedia(detail.Children)
			if result != nil {
				return result
			}
			continue
		}
	}
	return nil
}

func isMetaTrackMedia(boxDetail *box.BoxDetail) bool {
	if boxDetail.Type == box.MediaBox {
		for _, detail := range boxDetail.Children {
			if detail.Type == box.HandlerReferenceBox {
				hdlr := detail.Boxer.(*box.Hdlr)
				if string(hdlr.HandlerType[:]) == "meta" {
					return true
				}
			}
		}
	}
	return false
}

type sampleInfo struct {
	Size              uint32
	SampleCount       uint32
	ChunkOffsets      []uint32
	SamplesCountSlice []uint32
}

func (s *sampleInfo) getSampleOffset(input int) uint32 {
	if uint32(input) >= s.SampleCount {
		return 0
	}
	var comparand uint32 = 0
	for i := 0; i < len(s.SamplesCountSlice); i++ {
		if uint32(input) >= comparand && uint32(input) < s.SamplesCountSlice[i] {
			return s.ChunkOffsets[i] + s.Size*(uint32(input)-comparand)
		}
		comparand = s.SamplesCountSlice[i]
	}
	return 0
}

func getSampleInfoFromMetaTrackMediaBox(boxDetail *box.BoxDetail) (*sampleInfo, error) {
	if boxDetail.Type != box.MediaBox {
		return nil, fmt.Errorf("not mdia box")
	}
	info := &sampleInfo{}
	var stsc *box.Stsc
	chunkCount := 0
	for _, c1 := range boxDetail.Children {
		if c1.Type == box.MediaInformationBox {
			for _, c2 := range c1.Children {
				if c2.Type == box.SampleTableBox {
					for _, child := range c2.Children {
						result, ok := child.Boxer.(*box.Stsc)
						if ok {
							stsc = result
						}
						stsz, ok := child.Boxer.(*box.Stsz)
						if ok {
							info.Size = stsz.Size
							info.SampleCount = stsz.Count
						}
						stco, ok := child.Boxer.(*box.Stco)
						if ok {
							info.ChunkOffsets = stco.Offsets
							chunkCount = int(stco.Count)
						}
					}
				}
			}
		}
	}
	if stsc == nil {
		return nil, fmt.Errorf("not found stsc")
	}
	info.SamplesCountSlice = make([]uint32, chunkCount)
	var x uint32 = 0
	for i, j := 0, 0; i < chunkCount; i++ {
		samplesPerChunk := stsc.Entries[j].SamplesPerChunk
		if j+1 < int(stsc.EntryCount) && i+2 >= int(stsc.Entries[j+1].FirstChunk) {
			j++
		}
		x += samplesPerChunk
		info.SamplesCountSlice[i] = x
	}
	return info, nil
}

func getMetaSampleInfo(r io.ReadSeeker) (*sampleInfo, error) {
	fileStructure, err := box.ReadFileStructure(r)
	if err != nil {
		return nil, err
	}
	return getSampleInfoFromMetaTrackMediaBox(searchMetaTrackMedia(fileStructure.BoxDetails))
}
