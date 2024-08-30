package box

import (
	"github.com/fukco/media-metadata/internal/manufacturer"
	"io"
)

type FileStructure struct {
	BoxDetails []*BoxDetail
	*Context
}

type BoxDetail struct {
	*BoxInfo
	Boxer
	Children []*BoxDetail
}

func ReadFileStructure(r io.ReadSeeker) (*FileStructure, error) {
	end, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	if _, err = r.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	fileStructure := &FileStructure{
		Context: &Context{},
	}
	details, err := readBoxDetails(r, fileStructure, uint64(end))
	if err != nil {
		return nil, err
	}
	fileStructure.BoxDetails = details
	return fileStructure, nil
}

func readBoxDetails(r io.ReadSeeker, fileStructure *FileStructure, end uint64) ([]*BoxDetail, error) {
	details := make([]*BoxDetail, 0, 8)

	for {
		sn, _ := r.Seek(0, io.SeekCurrent)
		if uint64(sn) >= end {
			break
		}
		bi, err := ReadBoxInfo(r)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if !bi.IsSupportedBox() {
			_, err = bi.SeekToEnd(r)
			if err != nil {
				return nil, err
			}
			continue
		}

		if fileStructure.Mfr == manufacturer.Unknown {
			fileStructure.Mfr = bi.Type.GetManufacturer()
		}

		payload, err := ReadBoxPayload(r, bi, fileStructure.Context)
		if err != nil {
			return nil, err
		}
		boxDetail := &BoxDetail{
			BoxInfo: bi,
			Boxer:   payload,
		}

		if bi.IsContainerBox() {
			_, err = bi.SeekToPayload(r)
			if err != nil {
				return nil, err
			}
			children, err := readBoxDetails(r, fileStructure, bi.Offset+bi.Size)
			if err != nil {
				return nil, err
			}
			if len(children) > 0 {
				boxDetail.Children = children
			}
		}
		details = append(details, boxDetail)

	}
	return details, nil
}
