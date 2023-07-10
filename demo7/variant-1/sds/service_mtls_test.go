package sds_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"os"
	"testing"

	"envoy-sds/sds"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	secret "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
)

const (
	basedir     = "/Users/xuanyuanxingjian/Documents/projects/Repos/github-envoy-playground/demo7/variant-1"
	virtualhost = "localhost" // to add custom domains, modify the /etc/hosts file
	localhost   = "127.0.0.1"
	port        = ":50051"
	protocol    = "tcp"
)

func TestService_FetchSecrets_MTLS(t *testing.T) {
	srv := sds.New()
	defer srv.Stop()

	tlsCredentials, err := loadTLSCredentials()
	if err != nil {
		log.Fatal("cannot load TLS credentials: ", err)
	}

	s := grpc.NewServer(grpc.Creds(tlsCredentials))
	srv.Register(s)

	lis, err := net.Listen(protocol, localhost+port)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(fmt.Sprintf("Server exited with error: %v", err))
		}
	}()

	ctx := context.Background()

	tlsClientCredentials, err := loadClientTLSCredentials()
	if err != nil {
		log.Fatal("cannot load client TLS credentials: ", err)
	}

	conn, err := grpc.DialContext(
		ctx,
		virtualhost+port,
		grpc.WithTransportCredentials(tlsClientCredentials),
	)
	if err != nil {
		t.Fatalf("Failed to dial virtualhost: %v", err)
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

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	pemServerCA, err := os.ReadFile(basedir + "/cert/ca.crt")
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	serverCert, err := tls.LoadX509KeyPair(
		basedir+"/cert/server.crt",
		basedir+"/cert/server.key",
	)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(config), nil
}

func loadClientTLSCredentials() (credentials.TransportCredentials, error) {
	pemServerCA, err := os.ReadFile(basedir + "/cert/ca.crt")
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	clientCert, err := tls.LoadX509KeyPair(
		basedir+"/cert/client.crt",
		basedir+"/cert/client.key",
	)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config), nil
}
