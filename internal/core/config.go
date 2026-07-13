package core

type Server struct {
	Port int
}

type Config struct {
	Server   Server
	Routes   []Route
	Backends []Backend
}
