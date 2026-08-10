package core

type Route struct {
	Method  string `yaml:"method"`
	Path    string `yaml:"path"`
	Backend string `yaml:"backend"`
}
