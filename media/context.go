package media

import "github.com/fukco/media-meta-parser/manufacturer"

type Context struct {
	Manufacturer manufacturer.Manufacturer
	//ModelName        string
	MediaType        Type
	MajorBrand       string
	CompatibleBrands []string
}
