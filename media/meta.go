package media

type Meta struct {
	Items []interface{}
	Temp  map[string]interface{} `json:"-"`
}
