package router

// type Params map[string]string. each map will cause heap allocation that gave 2 allocs/op

type Param struct {
	Key   string
	Value string
}

// Params is a slice instead of a map to prevent heap allocations
type Params []Param

func (p Params) Get(key string) string {
	for _, param := range p {
		if param.Key == key {
			return param.Value
		}
	}
	return ""
}
