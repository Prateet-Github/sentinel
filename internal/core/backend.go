package core

type Backend struct {
	Name            string `yaml:"name"`
	URL             string `yaml:"url"`
	HealthCheckPath string `yaml:"health_check_path"`
}
