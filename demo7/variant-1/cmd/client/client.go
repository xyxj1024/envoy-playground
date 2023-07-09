package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	secret "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
)

const (
	protocol = "unix"
	socket   = "/tmp/uds_path"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dialer := func(ctx context.Context, addr string) (net.Conn, error) {
		return net.Dial(protocol, addr)
	}

	conn, err := grpc.Dial(
		socket,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialer),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Returns a new client API for SDS service
	client := secret.NewSecretDiscoveryServiceClient(conn)

	req := []*discovery.DiscoveryRequest{
		{
			VersionInfo:   "versionInfo",
			Node:          &core.Node{Id: "test-id", Cluster: "test-cluster"},
			ResourceNames: []string{"one"},                                                        // list of resources to subscribe to
			TypeUrl:       "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret", // type of the resource that is being requested
			ResponseNonce: "response-nonce",
		},
	}

	stream, err := client.StreamSecrets(ctx)
	if err != nil {
		log.Fatalf("%v.StreamSecrets(): %v", client, err)
	}

	for _, r := range req {
		if err := stream.Send(r); err != nil {
			log.Fatalf("%v.Send(%v): %v", stream, r, err)
		}
	}

	res, err := stream.Recv()
	if err != nil {
		log.Fatalf("%v.Recv(): %v", stream, err)
	}
	log.Printf("%v", res)
}
