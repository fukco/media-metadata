package box

import (
	"bytes"
	"encoding/binary"
	"io"
)

type BoxHeader struct {
	Size         uint32
	Type         BoxType
	ExtendedSize uint64
	UserType     UserType
}

type FullBoxHeader struct {
	Version uint8
	Flags   [3]byte
}

func ReadBoxHeader(r io.ReadSeeker) (*BoxHeader, *FullBoxHeader, error) {
	header := &BoxHeader{}
	fullBoxHeader := &FullBoxHeader{}

	// read 8 bytes
	buf := bytes.NewBuffer([]byte{})
	if _, err := io.CopyN(buf, r, 8); err != nil {
		return nil, nil, err
	}

	// pick size and type
	data := buf.Next(8)
	header.Size = binary.BigEndian.Uint32(data[:4])
	header.Type = BoxType(binary.BigEndian.Uint32(data[4:8]))

	if header.Size == 1 {
		// read more 8 bytes
		if _, err := io.CopyN(buf, r, 8); err != nil {
			return nil, nil, err
		}
		header.ExtendedSize = binary.BigEndian.Uint64(buf.Next(8))
	}

	def := header.Type.getBoxDef()
	if header.Type == UUIDExtensionBox {
		if _, err := io.CopyN(buf, r, 16); err != nil {
			return nil, nil, err
		}
		copy(header.UserType[:], buf.Next(16))
		def = header.UserType.getBoxDef()
	}

	if def != nil {
		if def.boxOrFullBox != IsBox {
			// read more 4 bytes
			if _, err := io.CopyN(buf, r, 4); err != nil {
				return nil, nil, err
			}
			data = buf.Next(4)

			if def.boxOrFullBox == IsFullBox {
				fullBoxHeader.Version = data[0]
				fullBoxHeader.Flags = [...]byte{data[1], data[2], data[3]}
				return header, fullBoxHeader, nil
			} else if def.boxOrFullBox == Unknown {
				if data[0]|data[1]|data[2]|data[3] == 0x00 {
					fullBoxHeader.Version = 0
					fullBoxHeader.Flags = [3]byte{0, 0, 0}
					return header, fullBoxHeader, nil
				} else {
					if _, err := r.Seek(-4, io.SeekCurrent); err != nil {
						return nil, nil, err
					}
				}

			}
		}
	}

	return header, nil, nil
}
