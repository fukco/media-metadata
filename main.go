package main

/*
#include <stdio.h>
struct DRMetadata
{
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
	char *Shutter;
	char *ISO;
	char *WhitePoint;
	char *WhiteBalanceTint;
	char *CameraFirmware;
	char *LensType;
	char *LensNumber;
	char *LensNotes;
	char *CameraApertureType;
	char *CameraAperture;
	char *FocalPoint;
	char *Distance;
	char *Filter;
	char *NdFilter;
	char *CompressionRatio;
	char *CodecBitrate;
	char *AspectRatioNotes;
	char *GammaNotes;
	char *ColorSpaceNotes;
};
*/
import "C"

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/fukco/media-meta-parser/media"
	"github.com/fukco/media-meta-parser/output"
	"golang.org/x/text/encoding/unicode"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

// getMediaFiles is a func return media files
func getMediaFiles(p string, recursive bool) ([]*os.File, error) {
	fi, err := os.Stat(p)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, errors.New("input is not a folder path")
	}
	mediaFiles := make([]*os.File, 0, 64)
	if err := filepath.WalkDir(p, func(path string, entry os.DirEntry, err error) error {
		if !recursive && p != path {
			if entry.IsDir() {
				return filepath.SkipDir
			}
		}
		mediaFile, _ := os.Open(path)
		if IsSupportExtension(mediaFile) {
			mediaFiles = append(mediaFiles, mediaFile)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return mediaFiles, nil
}

func getMediaFile(p string) (*os.File, error) {
	if fileInfo, err := os.Stat(p); err != nil {
		return nil, err
	} else {
		if fileInfo.IsDir() {
			return nil, errors.New("input file path is illegal")
		}
	}
	mediaFile, err := os.Open(p)
	if err != nil || !IsSupportExtension(mediaFile) {
		return nil, err
	}
	return mediaFile, nil
}

func getMediaMeta(mediaFiles []*os.File) []*media.Meta {
	if len(mediaFiles) <= 0 {
		return nil
	}
	slice := make([]*media.Meta, 0, len(mediaFiles))
	for i := range mediaFiles {
		mediaFile := mediaFiles[i]
		var meta *media.Meta
		is, ctx, err := IsSupportMediaFile(mediaFile)
		if err != nil {
			fmt.Println(err)
			continue
		} else if !is {
			continue
		}
		meta = ExtractMeta(mediaFile, ctx)
		if meta != nil {
			slice = append(slice, meta)
		}
	}
	return slice
}

func consoleOutput(metaSlice []*media.Meta) {
	for i := range metaSlice {
		meta := metaSlice[i]
		if s, err := json.Marshal(meta); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(string(s))
		}
	}
}

func resolveCSVOutput(metaSlice []*media.Meta, outputPath string) {
	if len(metaSlice) <= 0 {
		return
	}
	csvFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Println(err)
	}
	enc := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewEncoder()
	writer := csv.NewWriter(enc.Writer(csvFile))
	reflectedValue := reflect.ValueOf(&output.DRMetadata{})
	header := make([]string, 0, reflectedValue.Type().Elem().NumField())
	for i := 0; i < reflectedValue.Type().Elem().NumField(); i++ {
		name := reflectedValue.Type().Elem().Field(i).Tag.Get("csv")
		header = append(header, name)
	}
	if err := writer.Write(header); err != nil {
		fmt.Println(err)
	}
	defer writer.Flush()

	for i := range metaSlice {
		meta := metaSlice[i]
		dir, file := filepath.Split(meta.MediaPath)
		if os.IsPathSeparator(dir[len(dir)-1]) {
			dir = dir[:len(dir)-1]
		}
		drMetadata := &output.DRMetadata{
			FileName:      file,
			ClipDirectory: dir,
		}
		if err := output.DRMetadataFromMeta(meta, drMetadata); err != nil {
			fmt.Println(err)
		}
		entry := make([]string, 0, reflectedValue.Type().Elem().NumField())
		for i := 0; i < reflectedValue.Type().Elem().NumField(); i++ {
			v := reflect.ValueOf(drMetadata).Elem().Field(i).Interface()
			entry = append(entry, fmt.Sprint(v))

		}
		if err := writer.Write(entry); err != nil {
			fmt.Println(err)
		}
	}
	fmt.Printf("Metadata write to %s succeed!\n", outputPath)
}

func drProcessMediaFile(absPath string) *output.DRMetadata {
	drMetadata := &output.DRMetadata{}
	if f, err := os.Open(absPath); err != nil {
		return nil
	} else {
		var meta *media.Meta
		is, ctx, err := IsSupportMediaFile(f)
		if err != nil {
			fmt.Println(err)
			return nil
		} else if !is {
			return nil
		}
		meta = ExtractMeta(f, ctx)
		if meta == nil {
			return nil
		}
		if err := output.DRMetadataFromMeta(meta, drMetadata); err != nil {
			fmt.Println(err)
			return nil
		}
		return drMetadata
	}
}

//export DRProcessMediaFile
func DRProcessMediaFile(absPath *C.char) C.struct_DRMetadata {
	drMetadata := drProcessMediaFile(C.GoString(absPath))
	var result C.struct_DRMetadata
	if drMetadata != nil {
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
		result.Shutter = C.CString(drMetadata.Shutter)
		result.ISO = C.CString(drMetadata.ISO)
		result.WhitePoint = C.CString(drMetadata.WhitePoint)
		result.WhiteBalanceTint = C.CString(drMetadata.WhiteBalanceTint)
		result.CameraFirmware = C.CString(drMetadata.CameraFirmware)
		result.LensType = C.CString(drMetadata.LensType)
		result.LensNumber = C.CString(drMetadata.LensNumber)
		result.LensNotes = C.CString(drMetadata.LensNotes)
		result.CameraApertureType = C.CString(drMetadata.CameraApertureType)
		result.CameraAperture = C.CString(drMetadata.CameraAperture)
		result.FocalPoint = C.CString(drMetadata.FocalPoint)
		result.Distance = C.CString(drMetadata.Distance)
		result.Filter = C.CString(drMetadata.Filter)
		result.NdFilter = C.CString(drMetadata.NdFilter)
		result.CompressionRatio = C.CString(drMetadata.CompressionRatio)
		result.CodecBitrate = C.CString(drMetadata.CodecBitrate)
		result.AspectRatioNotes = C.CString(drMetadata.AspectRatioNotes)
		result.GammaNotes = C.CString(drMetadata.GammaNotes)
		result.ColorSpaceNotes = C.CString(drMetadata.ColorSpaceNotes)
	}
	return result
}

// -file /path/to/file -output_mode resolve -o /path/to/folder
// -folder /path/to/folder (-r)
func main() {
	mediaFile := flag.String("file", "", "media file full path")
	mediaFileFolder := flag.String("folder", "", "media file folder")
	recursive := flag.Bool("recursive", false, "read media file folder recursively or not(effect when use -folder mode)")
	outputPath := flag.String("o", "", "output folder")
	outputMode := flag.String("output_mode", "console", "output mode: 'console','resolve'")
	flag.Parse()

	if *mediaFile == "" && *mediaFileFolder == "" {
		fmt.Println("Please input file/folder path!")
		os.Exit(1)
	}
	var metaSlice []*media.Meta
	if *mediaFile != "" {
		file, err := getMediaFile(*mediaFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if file != nil {
			metaSlice = getMediaMeta([]*os.File{file})
		}
	} else {
		files, err := getMediaFiles(*mediaFileFolder, *recursive)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		metaSlice = getMediaMeta(files)
	}

	if *outputMode == "console" {
		consoleOutput(metaSlice)
	} else {
		if *outputPath == "" {
			if *mediaFile != "" {
				split, _ := filepath.Split(*mediaFile)
				*outputPath = split
			} else {
				*outputPath = *mediaFileFolder
			}
		} else {
			if info, err := os.Stat(*outputPath); err != nil {
				if os.IsNotExist(err) {
					fmt.Println("Your input path is not exist!")
				} else {
					fmt.Println(err)
				}
				os.Exit(1)
			} else if !info.IsDir() {
				fmt.Println("Your output folder is illegal!")
				os.Exit(1)
			}
		}
		if *outputMode == "resolve" {
			outputFile := filepath.Join(*outputPath, fmt.Sprintf("resolve-meta_%s.csv", time.Now().Format("20060102150405")))
			resolveCSVOutput(metaSlice, outputFile)
		}
	}
	fmt.Println("Processing Successfully!")
}
