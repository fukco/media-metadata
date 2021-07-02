package panasonic

type ClipMain struct {
	ClipContent ClipContent
	UserArea    UserArea
}

type ClipContent struct {
	Duration     string
	EssenceList  EssenceList
	ClipMetadata ClipMetadata
}

type EssenceList struct {
	Video Video
}

type Video struct {
	ActiveLine    string
	ActivePixel   string
	BitDepth      string
	FrameRate     string
	StartTimecode string
}

type ClipMetadata struct {
	Device Device
}

type Device struct {
	Manufacturer string
	ModelName    string
}

type UserArea struct {
	AcquisitionMetadata AcquisitionMetadata
}

type AcquisitionMetadata struct {
	CameraUnitMetadata CameraUnitMetadata
}

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
