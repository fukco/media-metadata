package xavc

import (
	"os"
	"testing"
)

func TestReadMdat(t *testing.T) {
	f, err := os.Open("D:\\Videos\\From Network\\影视飓风\\mediastorm-free-a7s3-001.MP4")
	defer func() {
		if err := f.Close(); err != nil {
			return
		}
	}()
	if err != nil {
		return
	} else {
		mdat, err := ReadMdat(f)
		print(mdat, err)
	}
}

func TestReadRtmdInBlock(t *testing.T) {
	f, err := os.Open("D:\\Videos\\From Network\\影视飓风\\mediastorm-free-a7s3-001.MP4")
	defer func() {
		if err := f.Close(); err != nil {
			return
		}
	}()
	if err != nil {
		return
	} else {
		mdat, err := ReadMdat(f)
		block := ReadRtmdInBlock(f, mdat.Offset+mdat.HeaderSize)
		print(block, err)
	}
}

func TestGetBlockSize(t *testing.T) {
	//f, err := os.Open("D:\\Videos\\From Network\\影视飓风\\mediastorm-free-a7s3-001.MP4")
	//f, err := os.Open("D:\\Downloads\\C0089-copy1.MP4")
	f, err := os.Open("D:\\Videos\\Test\\XAVC S\\20220217_C0547.MP4")
	defer func() {
		if err := f.Close(); err != nil {
			return
		}
	}()
	if err != nil {
		return
	} else {
		mdat, _ := ReadMdat(f)
		header, blockSize, _ := getHeaderAndBlockSize(f, mdat)
		print(header, blockSize)
	}
}

func TestReadRtmdSlice(t *testing.T) {
	f, err := os.Open("D:\\Videos\\A7SIII\\2022TX Stabilizer\\20220228_C0598.MP4")
	defer func() {
		if err := f.Close(); err != nil {
			return
		}
	}()
	if err != nil {
		return
	} else {
		rtmdCollection, _ := ReadRtmdSlice(f, 0, 1000, 0)
		_ = rtmdCollection
	}
}
