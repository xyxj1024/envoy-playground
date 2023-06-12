## 1. Preparation

[Install the protocol buffer compiler](https://grpc.io/docs/protoc-installation/):

```bash
brew install protobuf
```

[Generate Go code](https://protobuf.dev/reference/go/go-generated/#package) for `echo.proto`:

```bash
# The -u flag instructs get to update modules providing dependencies of packages
# named on the command line to use newer minor or patch releases when available.
go get -u google.golang.org/protobuf/cmd/protoc-gen-go
go install google.golang.org/protobuf/cmd/protoc-gen-go
go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc

protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       protos/echo.proto
```

Go module initialization:

```bash
go mod init envoy-demo4 && go mod tidy
```

## 2. Run Without xDS Management Server

First, modify the line `127.0.0.1 localhost` into `127.0.0.1 be.cluster.local` inside `/etc/hosts`. (One might need to run `go env -w GO111MODULE=on` based on their Go settings.)

```bash
# Terminal 1
go run server/server.go --grpcport :50051 --servername server1

# Terminal 2
go run client/client.go dns:///be.cluster.local:50051
```

## 3. xDS-Based Global Load Balancing

Start three gRPC servers:

```bash
go run server/server.go --grpcport :50051 --servername server1
go run server/server.go --grpcport :50052 --servername server2
go run server/server.go --grpcport :50053 --servername server3
```

Start xDS management server:

```bash
go run xds/main.go --upstream_port=50051 --upstream_port=50052 --upstream_port=50053
```

Start gRPC client:

```bash
export GRPC_XDS_BOOTSTRAP=xds/xds_bootstrap.json
go run client/client.go --host xds:///be-srv
```

Then, we would find consistency check for snapshot failed:

```console
mismatched "type.googleapis.com/envoy.config.route.v3.RouteConfiguration" reference and resource lengths: len(map[]) != 1
```

Comment out the consistency check, and the code should work.