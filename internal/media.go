package internal

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func getMediaFormat(file *os.File) Format {
	if strings.EqualFold(filepath.Ext(file.Name()), string(Mp4Extension)) ||
		strings.EqualFold(filepath.Ext(file.Name()), string(NRAWExtension)) {
		return Mp4
	}
	if strings.EqualFold(filepath.Ext(file.Name()), string(MovExtension)) {
		return Quicktime
	}
	return Unknown
}

func isSupportExtension(file *os.File) bool {
	if strings.EqualFold(filepath.Ext(file.Name()), string(Mp4Extension)) ||
		strings.EqualFold(filepath.Ext(file.Name()), string(MovExtension)) ||
		strings.EqualFold(filepath.Ext(file.Name()), string(NRAWExtension)) {
		return true
	}
	return false
}

// IsSupportMediaFile check file is support media file
func IsSupportMediaFile(file *os.File) bool {
	if !isSupportExtension(file) {
		return false
	}
	buf := make([]byte, 4)
	if _, err := file.ReadAt(buf, 4); err != nil {
		return false
	}
	if string(buf) == "ftyp" {
		return true
	}

	return false
}

func GetMediaFile(s string) (*os.File, error) {
	if fileInfo, err := os.Stat(s); err != nil {
		return nil, err
	} else {
		if fileInfo.IsDir() {
			return nil, errors.New("input file path is illegal")
		}
	}
	f, err := os.Open(s)
	if err != nil {
		return nil, err
	}
	if !IsSupportMediaFile(f) {
		return nil, errors.New("not support media file")
	}
	return f, nil
}
