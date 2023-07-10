## Demo Setup

The demo code is mostly copied from [https://github.com/kanurag94/envoy-sds](https://github.com/kanurag94/envoy-sds) and [Maksim Paskal's lightweight Envoy control plane](https://github.com/maksim-paskal/envoy-control-plane/tree/main/pkg/certs).

The client simply sends a `DiscoveryRequest` to the SDS server and receives a `DiscoveryResponse`.

## Run Code

```bash
# Initialize Go module
go mod init envoy-sds && go mod tidy

# Generate certificates
go run cmd/certgen/certgen.go

# Run tests
go test envoy-sds/cert && go test envoy-sds/sds

# Terminal 1
go run cmd/server.go

# Terminal 2
go run cmd/client/client.go
```

Example output by the client:

```console
2023/07/09 23:27:26 DiscoveryResponse:
version_info:"2023-07-10T04:27:26Z" resources:{type_url:"type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret" value:"\n\x03one*$\n\"\x12 \x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"} type_url:"type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret" nonce:"3588b5bcf7a5424f17bc0cc47154cb374d511ca47486305c33f4c7548eb929e7d33364cc44a0a2c0919178aa2fee5f3fafa0ef5f0744d0ae7a11da3278c3cc64" control_plane:{identifier:"control_plane"}
```