package sds

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	secret "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
)

// Service is used as a SDS service handler
type Service struct {
	stopChannel chan struct{}
}

func New() *Service {
	return &Service{
		stopChannel: make(chan struct{}),
	}
}

func (s *Service) Stop() error {
	close(s.stopChannel)
	return nil
}

func (s *Service) Register(grpcServer *grpc.Server) {
	secret.RegisterSecretDiscoveryServiceServer(grpcServer, s)
}

func (s *Service) DeltaSecrets(dss secret.SecretDiscoveryService_DeltaSecretsServer) (err error) {
	return errors.New("method DeltaSecrets not implemented")
}

func (s *Service) StreamSecrets(sss secret.SecretDiscoveryService_StreamSecretsServer) (err error) {
	errorChannel := make(chan error)
	requestChannel := make(chan *discovery.DiscoveryRequest)

	go func() {
		for {
			r, err := sss.Recv()
			if err != nil {
				errorChannel <- err
				return
			}
			requestChannel <- r
		}
	}()

	var (
		nonce       string
		versionInfo string
		req         *discovery.DiscoveryRequest
	)

	for {
		select {
		case r := <-requestChannel:
			if r.ErrorDetail != nil {
				fmt.Printf("failed discovery request, error: %s\n", err.Error())
				continue
			}

			if req != nil {
				switch {
				case nonce != r.ResponseNonce:
					fmt.Println("invalid responseNonce")
					continue
				case r.VersionInfo == "": // first DiscoveryRequest
					versionInfo = s.versionInfo()
				case r.VersionInfo == versionInfo: // consecutive request ACK
					fmt.Println("version_info received")
					continue
				default:
					versionInfo = s.versionInfo()
				}
			} else {
				versionInfo = s.versionInfo()
			}

			req = r
			for _, name := range req.ResourceNames {
				fmt.Printf("Request for resource: %s received\n", name)
			}

		case err := <-errorChannel:
			fmt.Printf("error occurred on channel: %s\n", err.Error())
			return err
		case <-s.stopChannel:
			return nil
		}

		res, err := getDiscoveryResponse(req, versionInfo)
		if err != nil {
			fmt.Printf("error while creating response: %s\n", err.Error())
			return err
		}

		if err := sss.Send(res); err != nil {
			fmt.Printf("error sending stream response: %s\n", err.Error())
			return err
		}
		fmt.Println("Response sent to the client")

		nonce = res.Nonce
	}
}

func (s *Service) FetchSecrets(ctx context.Context, req *discovery.DiscoveryRequest) (*discovery.DiscoveryResponse, error) {
	return getDiscoveryResponse(req, s.versionInfo())
}

func getDiscoveryResponse(req *discovery.DiscoveryRequest, versionInfo string) (*discovery.DiscoveryResponse, error) {
	nonce, err := randomHex(64)
	if err != nil {
		return nil, fmt.Errorf("error generating nonce: %s", err.Error())
	}

	var b []byte
	var resources []*anypb.Any
	for _, name := range req.ResourceNames {
		b, err = getGenericSecret(name)
		if err != nil {
			return nil, err
		}
		resources = append(resources, &anypb.Any{
			TypeUrl: req.TypeUrl,
			Value:   b,
		})
	}

	return &discovery.DiscoveryResponse{
		VersionInfo: versionInfo,
		Resources:   resources,
		TypeUrl:     req.TypeUrl,
		Nonce:       nonce,
		Canary:      false,
		ControlPlane: &core.ControlPlane{
			Identifier: "control_plane",
		},
	}, nil
}

// randomHex returns a random hexadecimal string
func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func getGenericSecret(name string) ([]byte, error) {
	secret := tls.Secret{
		Name: name,
		Type: &tls.Secret_GenericSecret{
			GenericSecret: &tls.GenericSecret{
				Secret: &core.DataSource{
					Specifier: &core.DataSource_InlineBytes{
						InlineBytes: make([]byte, 32),
					},
				},
			},
		},
	}
	v, err := proto.Marshal(&secret)
	if err != nil {
		return v, fmt.Errorf("error marshaling secret: %s", err.Error())
	}
	return v, err
}

// versionInfo returns current time as version information
func (s *Service) versionInfo() string {
	return time.Now().UTC().Format(time.RFC3339)
}
