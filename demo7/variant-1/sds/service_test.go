package sds_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	"envoy-sds/sds"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	secret "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
)

func TestService_FetchSecrets(t *testing.T) {
	srv := sds.New()
	defer srv.Stop()

	s := grpc.NewServer()
	srv.Register(s)

	lis := bufconn.Listen(1024 * 1024)
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(fmt.Sprintf("Server exited with error: %v", err))
		}
	}()

	ctx := context.Background()

	dialer := func(ctx context.Context, str string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.DialContext(
		ctx,
		"bufconn",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufconn: %v", err)
	}
	defer conn.Close()

	client := secret.NewSecretDiscoveryServiceClient(conn)

	tests := []struct {
		name      string
		req       *discovery.DiscoveryRequest
		succeeded bool
	}{
		{"ok", &discovery.DiscoveryRequest{
			VersionInfo:   "versionInfo",
			Node:          &core.Node{Id: "test-id", Cluster: "test-cluster"},
			ResourceNames: []string{"one"},
			TypeUrl:       "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret",
			ResponseNonce: "response-nonce",
		}, false},
		{"ok multiple", &discovery.DiscoveryRequest{
			VersionInfo:   "versionInfo",
			Node:          &core.Node{Id: "test-id", Cluster: "test-cluster"},
			ResourceNames: []string{"one", "two"},
			TypeUrl:       "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret",
			ResponseNonce: "response-nonce",
		}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c, err := client.FetchSecrets(context.Background(), test.req)
			fmt.Println(c)
			if (err != nil) != test.succeeded {
				t.Errorf("Service.FetchSecrets() error: %v, succeeded: %v", err, test.succeeded)
				return
			}
		})
	}
}
