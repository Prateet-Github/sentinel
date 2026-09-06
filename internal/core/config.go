package core

type Server struct {
	Port int `yaml:"port"`
}

type RateLimit struct {
	Enabled    bool    `yaml:"enabled"`
	Capacity   float64 `yaml:"capacity"`
	RefillRate float64 `yaml:"refill_rate"`
}

type Config struct {
	Server    Server    `yaml:"server"`
	RateLimit RateLimit `yaml:"rate_limit"`
	Routes    []Route   `yaml:"routes"`
	Backends  []Backend `yaml:"backends"`
}
