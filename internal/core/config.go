package core

type Server struct {
	Port int `yaml:"port"`
}

type Config struct {
	Server   Server    `yaml:"server"`
	Routes   []Route   `yaml:"routes"`
	Backends []Backend `yaml:"backends"`
}
