package nrtmd

const (
	CameraUnitMetadataSet = "CameraUnitMetadataSet"
	CaptureGammaEquation  = "CaptureGammaEquation"
	CaptureColorPrimaries = "CaptureColorPrimaries"
)

type NonRealTimeMeta struct {
	Duration          Duration
	CreationDate      CreationDate
	Device            Device
	VideoFormat       VideoFormat
	SubStream         SubStream
	RecordingMode     RecordingMode
	AcquisitionRecord AcquisitionRecord
}

type Duration struct {
	Value string `xml:"value,attr"`
}

type CreationDate struct {
	Value string `xml:"value,attr"`
}

type VideoFormat struct {
	VideoFrame  VideoFrame
	VideoLayout VideoLayout
}

type VideoFrame struct {
	VideoCodec string `xml:"videoCodec,attr"`
	CaptureFps string `xml:"captureFps,attr"`
	FormatFps  string `xml:"formatFps,attr"`
}

type VideoLayout struct {
	Pixel             string `xml:"pixel,attr"`
	NumOfVerticalLine string `xml:"numOfVerticalLine,attr"`
	AspectRatio       string `xml:"aspectRatio,attr"`
}

type SubStream struct {
	Codec string `xml:"codec,attr"`
}

type RecordingMode struct {
	Type string `xml:"type,attr"`
}

type Device struct {
	Manufacturer string `xml:"manufacturer,attr"`
	ModelName    string `xml:"modelName,attr"`
	SerialNo     string `xml:"serialNo,attr"`
}

type AcquisitionRecord struct {
	Groups []Group `xml:"Group"`
}

type Group struct {
	Name  string      `xml:"name,attr"`
	Items []GroupItem `xml:"Item"`
}

type GroupItem struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}
