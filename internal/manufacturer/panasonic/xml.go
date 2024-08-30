package panasonic

type ClipMain struct {
	ClipContent        ClipContent
	CameraUnitMetadata CameraUnitMetadata `xml:"UserArea>AcquisitionMetadata>CameraUnitMetadata" json:"CameraUnitMetadata"`
}

type ClipContent struct {
	GlobalClipID string
	Duration     string
	Video        Video `xml:"EssenceList>Video"`
	ClipMetadata ClipMetadata
}

type Video struct {
	Codec         Codec
	ActiveLine    string
	ActivePixel   string
	BitDepth      string
	FrameRate     string
	StartTimecode string
}

type Codec struct {
	BitRate string `xml:",attr"`
	Codec   string `xml:",chardata"`
}

type ClipMetadata struct {
	Shoot  Shoot
	Device Device
}

type Shoot struct {
	StartDate string
}

type Device struct {
	Manufacturer string
	ModelName    string
}

//type UserArea struct {
//	AcquisitionMetadata AcquisitionMetadata
//}
//
//type AcquisitionMetadata struct {
//	CameraUnitMetadata CameraUnitMetadata
//}

type CameraUnitMetadata struct {
	ISOSensitivity string
	Gamma          Gamma
	Gamut          Gamut
}

type Gamma struct {
	CaptureGamma string
}

type Gamut struct {
	CaptureGamut string
}
