package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"envoy-sds/sds"

	"google.golang.org/grpc"
)

const (
	protocol = "unix"
	socket   = "/tmp/uds_path" /* Unix Domain Socket path */
)

func main() {
	lis, err := net.Listen(protocol, socket)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove(socket)
		os.Exit(1)
	}()

	grpcServer := grpc.NewServer()
	sds := sds.New()
	sds.Register(grpcServer)

	log.Printf("server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
