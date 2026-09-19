package main

import (
	"log"
	"net"

	"github.com/Prateet-Github/sentinel/internal/controlplane"
	controlv1 "github.com/Prateet-Github/sentinel/proto"
	"google.golang.org/grpc"
)

func main() {
	listener, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()

	controlServer := &controlplane.Server{}

	controlv1.RegisterSentinelControlServer(server, controlServer)

	log.Println("Sentinel Control Plane listening on :9090")

	if err := server.Serve(listener); err != nil {
		log.Fatal(err)
	}
}
