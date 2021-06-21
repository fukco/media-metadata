package sony

const (
	CameraUnitMetadataSet = "CameraUnitMetadataSet"
	CaptureGammaEquation  = "CaptureGammaEquation"
	CaptureColorPrimaries = "CaptureColorPrimaries"
)

type NonRealTimeMeta struct {
	Duration          Duration
	Device            Device
	VideoFormat       VideoFormat
	AcquisitionRecord AcquisitionRecord
}

type Duration struct {
	Value string `xml:"value,attr"`
}

type VideoFormat struct {
	VideoFrame  VideoFrame
	VideoLayout VideoLayout
}

type VideoFrame struct {
	CaptureFps string `xml:"captureFps,attr"`
}

type VideoLayout struct {
	Pixel             string `xml:"pixel,attr"`
	NumOfVerticalLine string `xml:"numOfVerticalLine,attr"`
	AspectRatio       string `xml:"aspectRatio,attr"`
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
