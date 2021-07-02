package media

type Meta struct {
	MediaPath string
	Context   *Context
	Items     []interface{}
	Temp      map[string]interface{} `json:"-"`
}
