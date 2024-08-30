package main

/*
#include <stdio.h>
#include <stdbool.h>
#include <stdlib.h>
struct DRMetadata
{
	bool IsSupportMedia;
	char *DateRecorded;
	char *CameraType;
	char *CameraManufacturer;
	char *CameraSerial;
	char *CameraId;
	char *CameraNotes;
	char *CameraFormat;
	char *MediaType;
	char *TimeLapseInterval;
	char *CameraFps;
	char *ShutterType;
	char *ShutterAngle;
	char *Shutter;
	char *ISO;
	char *WhitePoint;
	char *WhiteBalanceTint;
	char *CameraFirmware;
	char *LUTUsed;
	char *LensType;
	char *LensNumber;
	char *LensNotes;
	char *CameraApertureType;
	char *CameraAperture;
	char *FocalPoint;
	char *Distance;
	char *Filter;
	char *NDFilter;
	char *CompressionRatio;
	char *CodecBitrate;
	char *SensorAreaCaptured;
	char *PARNotes;
	char *AspectRatioNotes;
	char *GammaNotes;
	char *ColorSpaceNotes;
};

struct DRSonyNrtmd
{
	bool IsSupportMedia;
	char *Manufacturer;
	char *FileFormatAndRecFrameRate;
	char *ModelName;
	char *FormatFPS;
	char *CaptureFPS;
	char *VideoBitrate;
	char *Profile;
	char *RecordingMode;
	bool IsProxyOn;
	long long int CreationTimestamp;
	int TimecodeSecs;
	int TimecodeFrame;
};

typedef struct tagDRFrameData
{
	int Frame;
	char *Data;
} DRFrameData;

typedef struct tagDRFrameDataArray
{
	DRFrameData array[1000];
	int len;
} DRFrameDataArray;

struct DRSonyRtmdDisp
{
	DRFrameDataArray WhiteBalanceModeArray;
	DRFrameDataArray ExposureModeArray;
	DRFrameDataArray AutoFocusSensingAreaArray;
	DRFrameDataArray ShutterSpeedArray;
	DRFrameDataArray ApertureArray;
	DRFrameDataArray ISOArray;
	DRFrameDataArray FocalLengthArray;
	DRFrameDataArray FocalLength35mmArray;
	DRFrameDataArray FocusPositionArray;
	DRFrameDataArray CaptureGammaEquationArray;
	DRFrameDataArray CameraMasterGainAdjustmentArray;
	long long int Offset;
};
*/
import "C"

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fukco/media-metadata/internal"
	"github.com/fukco/media-metadata/internal/manufacturer"
	"github.com/fukco/media-metadata/internal/meta"
	"github.com/fukco/media-metadata/internal/output/resolve"
	"github.com/fukco/media-metadata/internal/output/resolve/xavc"
	"os"
)

func consoleOutput(meta *meta.Metadata) {
	if s, err := json.Marshal(meta); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(s))
	}
}

func drProcessMediaFile(absPath string) *resolve.DRMetadata {
	f, err := internal.GetMediaFile(absPath)
	defer f.Close()
	if err != nil {
		return nil
	}
	m, err := meta.Read(f)
	if err != nil || m == nil {
		return nil
	}
	return resolve.GetDRMetadataFromMeta(m)
}

//export DRProcessMediaFile
func DRProcessMediaFile(absPath *C.char) C.struct_DRMetadata {
	drMetadata := drProcessMediaFile(C.GoString(absPath))
	var result C.struct_DRMetadata
	if drMetadata == nil {
		result.IsSupportMedia = C._Bool(false)
	} else {
		result.IsSupportMedia = C._Bool(true)
		result.DateRecorded = C.CString(drMetadata.DateRecorded)
		result.CameraType = C.CString(drMetadata.CameraType)
		result.CameraManufacturer = C.CString(drMetadata.CameraManufacturer)
		result.CameraSerial = C.CString(drMetadata.CameraSerial)
		result.CameraId = C.CString(drMetadata.CameraId)
		result.CameraNotes = C.CString(drMetadata.CameraNotes)
		result.CameraFormat = C.CString(drMetadata.CameraFormat)
		result.MediaType = C.CString(drMetadata.MediaType)
		result.TimeLapseInterval = C.CString(drMetadata.TimeLapseInterval)
		result.CameraFps = C.CString(drMetadata.CameraFps)
		result.ShutterType = C.CString(drMetadata.ShutterType)
		result.ShutterAngle = C.CString(drMetadata.ShutterAngle)
		result.Shutter = C.CString(drMetadata.Shutter)
		result.ISO = C.CString(drMetadata.ISO)
		result.WhitePoint = C.CString(drMetadata.WhitePoint)
		result.WhiteBalanceTint = C.CString(drMetadata.WhiteBalanceTint)
		result.CameraFirmware = C.CString(drMetadata.CameraFirmware)
		result.LUTUsed = C.CString(drMetadata.LUTUsed)
		result.LensType = C.CString(drMetadata.LensType)
		result.LensNumber = C.CString(drMetadata.LensNumber)
		result.LensNotes = C.CString(drMetadata.LensNotes)
		result.CameraApertureType = C.CString(drMetadata.CameraApertureType)
		result.CameraAperture = C.CString(drMetadata.CameraAperture)
		result.FocalPoint = C.CString(drMetadata.FocalPoint)
		result.Distance = C.CString(drMetadata.Distance)
		result.Filter = C.CString(drMetadata.Filter)
		result.NDFilter = C.CString(drMetadata.NDFilter)
		result.CompressionRatio = C.CString(drMetadata.CompressionRatio)
		result.CodecBitrate = C.CString(drMetadata.CodecBitrate)
		result.SensorAreaCaptured = C.CString(drMetadata.SensorAreaCaptured)
		result.PARNotes = C.CString(drMetadata.PARNotes)
		result.AspectRatioNotes = C.CString(drMetadata.AspectRatioNotes)
		result.GammaNotes = C.CString(drMetadata.GammaNotes)
		result.ColorSpaceNotes = C.CString(drMetadata.ColorSpaceNotes)
	}
	return result
}

func drSonyNrtmdDisp(absPath string) *xavc.NrtmdDisp {
	f, err := internal.GetMediaFile(absPath)
	defer f.Close()

	if err != nil {
		fmt.Println(err)
		return nil
	}
	m, err := meta.Read(f)
	if err != nil || m == nil || m.Manufacturer != manufacturer.SONY {
		return nil
	}
	return xavc.NrtmdDispFromMeta(m)
}

//export DRSonyNrtmdDisp
func DRSonyNrtmdDisp(absPath *C.char) C.struct_DRSonyNrtmd {
	SonyNrtmdDisp := drSonyNrtmdDisp(C.GoString(absPath))
	var result C.struct_DRSonyNrtmd
	if SonyNrtmdDisp == nil {
		result.IsSupportMedia = C._Bool(false)
	} else {
		result.IsSupportMedia = C._Bool(true)
		result.Manufacturer = C.CString(SonyNrtmdDisp.Manufacturer)
		result.FileFormatAndRecFrameRate = C.CString(SonyNrtmdDisp.FileFormatAndRecFrameRate)
		result.ModelName = C.CString(SonyNrtmdDisp.ModelName)
		result.FormatFPS = C.CString(SonyNrtmdDisp.FormatFPS)
		result.CaptureFPS = C.CString(SonyNrtmdDisp.CaptureFPS)
		result.VideoBitrate = C.CString(SonyNrtmdDisp.VideoBitrate)
		result.Profile = C.CString(SonyNrtmdDisp.Profile)
		result.RecordingMode = C.CString(SonyNrtmdDisp.RecordingMode)
		result.IsProxyOn = C._Bool(SonyNrtmdDisp.IsProxyOn)
		result.CreationTimestamp = C.longlong(SonyNrtmdDisp.CreationTimestamp)
		result.TimecodeSecs = C.int(SonyNrtmdDisp.TimecodeSecs)
		result.TimecodeFrame = C.int(SonyNrtmdDisp.TimecodeFrame)
	}
	return result
}

func drSonyRtmdDisp(absPath string, start int, count int) *xavc.RtmdCollection {
	f, err := os.Open(absPath)
	defer func() {
		if err := f.Close(); err != nil {
			return
		}
	}()
	if err != nil {
		return nil
	} else {
		if rtmdCollection, err := xavc.ReadRtmdSlice(f, start, count); err != nil {
			return nil
		} else {
			return rtmdCollection
		}
	}
}

//export DrSonyRtmdDisp
func DrSonyRtmdDisp(absPath *C.char, start C.int, count C.int) C.struct_DRSonyRtmdDisp {
	rtmdCollection := drSonyRtmdDisp(C.GoString(absPath), int(start), int(count))
	var result C.struct_DRSonyRtmdDisp
	for i := 0; i < len(rtmdCollection.WhiteBalanceSlice); i++ {
		result.WhiteBalanceModeArray.array[i].Frame = C.int(rtmdCollection.WhiteBalanceSlice[i].Frame)
		result.WhiteBalanceModeArray.array[i].Data = C.CString(rtmdCollection.WhiteBalanceSlice[i].Data)
	}
	result.WhiteBalanceModeArray.len = C.int(len(rtmdCollection.WhiteBalanceSlice))
	for i := 0; i < len(rtmdCollection.ExposureModeSlice); i++ {
		result.ExposureModeArray.array[i].Frame = C.int(rtmdCollection.ExposureModeSlice[i].Frame)
		result.ExposureModeArray.array[i].Data = C.CString(rtmdCollection.ExposureModeSlice[i].Data)
	}
	result.ExposureModeArray.len = C.int(len(rtmdCollection.ExposureModeSlice))
	for i := 0; i < len(rtmdCollection.AutoFocusSensingAreaSlice); i++ {
		result.AutoFocusSensingAreaArray.array[i].Frame = C.int(rtmdCollection.AutoFocusSensingAreaSlice[i].Frame)
		result.AutoFocusSensingAreaArray.array[i].Data = C.CString(rtmdCollection.AutoFocusSensingAreaSlice[i].Data)
	}
	result.AutoFocusSensingAreaArray.len = C.int(len(rtmdCollection.AutoFocusSensingAreaSlice))
	for i := 0; i < len(rtmdCollection.ShutterSpeedSlice); i++ {
		result.ShutterSpeedArray.array[i].Frame = C.int(rtmdCollection.ShutterSpeedSlice[i].Frame)
		result.ShutterSpeedArray.array[i].Data = C.CString(rtmdCollection.ShutterSpeedSlice[i].Data)
	}
	result.ShutterSpeedArray.len = C.int(len(rtmdCollection.ShutterSpeedSlice))
	for i := 0; i < len(rtmdCollection.ApertureSlice); i++ {
		result.ApertureArray.array[i].Frame = C.int(rtmdCollection.ApertureSlice[i].Frame)
		result.ApertureArray.array[i].Data = C.CString(rtmdCollection.ApertureSlice[i].Data)
	}
	result.ApertureArray.len = C.int(len(rtmdCollection.ApertureSlice))
	for i := 0; i < len(rtmdCollection.ISOSlice); i++ {
		result.ISOArray.array[i].Frame = C.int(rtmdCollection.ISOSlice[i].Frame)
		result.ISOArray.array[i].Data = C.CString(rtmdCollection.ISOSlice[i].Data)
	}
	result.ISOArray.len = C.int(len(rtmdCollection.ISOSlice))
	for i := 0; i < len(rtmdCollection.FocalLengthSlice); i++ {
		result.FocalLengthArray.array[i].Frame = C.int(rtmdCollection.FocalLengthSlice[i].Frame)
		result.FocalLengthArray.array[i].Data = C.CString(rtmdCollection.FocalLengthSlice[i].Data)
	}
	result.FocalLengthArray.len = C.int(len(rtmdCollection.FocalLengthSlice))
	for i := 0; i < len(rtmdCollection.FocalLength35mmSlice); i++ {
		result.FocalLength35mmArray.array[i].Frame = C.int(rtmdCollection.FocalLength35mmSlice[i].Frame)
		result.FocalLength35mmArray.array[i].Data = C.CString(rtmdCollection.FocalLength35mmSlice[i].Data)
	}
	result.FocalLength35mmArray.len = C.int(len(rtmdCollection.FocalLength35mmSlice))
	for i := 0; i < len(rtmdCollection.FocusPositionSlice); i++ {
		result.FocusPositionArray.array[i].Frame = C.int(rtmdCollection.FocusPositionSlice[i].Frame)
		result.FocusPositionArray.array[i].Data = C.CString(rtmdCollection.FocusPositionSlice[i].Data)
	}
	result.FocusPositionArray.len = C.int(len(rtmdCollection.FocusPositionSlice))
	for i := 0; i < len(rtmdCollection.CaptureGammaEquationSlice); i++ {
		result.CaptureGammaEquationArray.array[i].Frame = C.int(rtmdCollection.CaptureGammaEquationSlice[i].Frame)
		result.CaptureGammaEquationArray.array[i].Data = C.CString(rtmdCollection.CaptureGammaEquationSlice[i].Data)
	}
	result.CaptureGammaEquationArray.len = C.int(len(rtmdCollection.CaptureGammaEquationSlice))
	for i := 0; i < len(rtmdCollection.CameraMasterGainAdjustmentSlice); i++ {
		result.CameraMasterGainAdjustmentArray.array[i].Frame = C.int(rtmdCollection.CameraMasterGainAdjustmentSlice[i].Frame)
		result.CameraMasterGainAdjustmentArray.array[i].Data = C.CString(rtmdCollection.CameraMasterGainAdjustmentSlice[i].Data)
	}
	result.CameraMasterGainAdjustmentArray.len = C.int(len(rtmdCollection.CameraMasterGainAdjustmentSlice))
	return result
}

// -file /path/to/file
func main() {
	filePath := flag.String("file", "", "media file full path")
	flag.Parse()

	if *filePath == "" {
		fmt.Println("Please input file path!")
		os.Exit(1)
	}
	f, err := internal.GetMediaFile(*filePath)
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	m, err := meta.Read(f)
	if err != nil {
		fmt.Println(err)
		return
	}
	consoleOutput(m)

	fmt.Println("Processing Successfully!")
}
