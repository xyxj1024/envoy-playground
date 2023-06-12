## 1. Preparation

Go module initialization:

```bash
$ go mod init envoy-demo5 && go mod tidy
```

[Install the protocol buffer compiler](https://grpc.io/docs/protoc-installation/):

```bash
$ brew install protobuf
$ protoc --version
libprotoc 3.21.12
```

[Generate Go code](https://protobuf.dev/reference/go/go-generated/#package) for `hello.proto`:

```bash
$ go get -u google.golang.org/protobuf/cmd/protoc-gen-go
$ go install google.golang.org/protobuf/cmd/protoc-gen-go
$ go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
$ protoc --go_out=. --go_opt=paths=source_relative \
>        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
>        protos/hello.proto
```

## 2. Run Without Containers

```bash
# Might need to set GO111MODULE=on for each shell process

# Remember to kill the server afterwards
$ go run ./server/server.go start

$ go run ./client/client.go
```

## 3. Containerized gRPC Server and [External Authorization](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/security/ext_authz_filter#arch-overview-ext-authz)

First, make sure to modify `client.go` so that it dials the containerized gRPC server at `0.0.0.0:1337` instead of `localhost:5050`. Then, build executables:

```bash
$ cd server
$ env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server server.go
$ cd ../auth
$ env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o auth auth.go
```

Finally, check it out:

```bash
# Terminal 1
$ docker-compose up

# Terminal 2
$ go run client/client.go
```

Envoy debugging info:

```console
envoy-grpc-example-extauthz-1  | 2023/06/02 02:51:15 >>> Authorization called check()
envoy-grpc-example-extauthz-1  | 2023/06/02 02:51:15 Inbound Headers: 
envoy-grpc-example-extauthz-1  | 2023/06/02 02:51:15 {
envoy-grpc-example-extauthz-1  |   ":authority": "localhost:1337",
envoy-grpc-example-extauthz-1  |   ":method": "POST",
envoy-grpc-example-extauthz-1  |   ":path": "/hello.HelloService/Hello",
envoy-grpc-example-extauthz-1  |   ":scheme": "http",
envoy-grpc-example-extauthz-1  |   "authorization": "Bearer foo",
envoy-grpc-example-extauthz-1  |   "bar": "baz",
envoy-grpc-example-extauthz-1  |   "content-type": "application/grpc",
envoy-grpc-example-extauthz-1  |   "te": "trailers",
envoy-grpc-example-extauthz-1  |   "user-agent": "grpc-go/1.55.0",
envoy-grpc-example-extauthz-1  |   "x-forwarded-proto": "http",
envoy-grpc-example-extauthz-1  |   "x-request-id": "e97c859f-2f2d-4d59-acb5-701c3fee50b4"
envoy-grpc-example-extauthz-1  | }
envoy-grpc-example-extauthz-1  | 2023/06/02 02:51:15 Context Extensions: 
envoy-grpc-example-extauthz-1  | 2023/06/02 02:51:15 {
envoy-grpc-example-extauthz-1  |   "x-forwarded-host": "original-host-as-context"
envoy-grpc-example-extauthz-1  | }
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.298][15][debug][router] [source/common/router/router.cc:1434] [C0][S15920461920280352603] upstream headers complete: end_stream=false
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.298][15][debug][http] [source/common/http/async_client_impl.cc:123] async http request response headers (end_stream=false):
envoy-grpc-example-envoy-1     | ':status', '200'
envoy-grpc-example-envoy-1     | 'content-type', 'application/grpc'
envoy-grpc-example-envoy-1     | 'x-envoy-upstream-service-time', '19'
envoy-grpc-example-envoy-1     | 
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.298][15][debug][client] [source/common/http/codec_client.cc:128] [C2] response complete
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.298][15][debug][pool] [source/common/conn_pool/conn_pool_base.cc:215] [C2] destroying stream: 0 remaining
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.298][15][debug][http] [source/common/http/async_client_impl.cc:150] async http request response trailers:
envoy-grpc-example-envoy-1     | 'grpc-status', '0'
envoy-grpc-example-envoy-1     | 'grpc-message', ''
envoy-grpc-example-envoy-1     | 
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.298][15][debug][router] [source/common/router/router.cc:478] [C1][S10104674368241947611] cluster 'backend' match for URL '/hello.HelloService/Hello'
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.298][15][debug][router] [source/common/router/router.cc:690] [C1][S10104674368241947611] router decoding headers:
envoy-grpc-example-envoy-1     | ':method', 'POST'
envoy-grpc-example-envoy-1     | ':scheme', 'http'
envoy-grpc-example-envoy-1     | ':path', '/hello.HelloService/Hello'
envoy-grpc-example-envoy-1     | ':authority', 'server.domain.com'
envoy-grpc-example-envoy-1     | 'content-type', 'application/grpc'
envoy-grpc-example-envoy-1     | 'user-agent', 'grpc-go/1.55.0'
envoy-grpc-example-envoy-1     | 'te', 'trailers'
envoy-grpc-example-envoy-1     | 'authorization', 'Bearer foo'
envoy-grpc-example-envoy-1     | 'bar', 'baz'
envoy-grpc-example-envoy-1     | 'x-forwarded-proto', 'http'
envoy-grpc-example-envoy-1     | 'x-request-id', 'e97c859f-2f2d-4d59-acb5-701c3fee50b4'
envoy-grpc-example-envoy-1     | 'x-custom-header-from-authz', 'permitted'
envoy-grpc-example-envoy-1     | 'x-envoy-expected-rq-timeout-ms', '15000'
envoy-grpc-example-envoy-1     | 'x-custom-to-backend', 'value-for-backend-from-envoy'
envoy-grpc-example-envoy-1     | 
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.298][15][debug][pool] [source/common/http/conn_pool_base.cc:78] queueing stream due to no available connections (ready=0 busy=0 connecting=0)
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.298][15][debug][pool] [source/common/conn_pool/conn_pool_base.cc:291] trying to create new connection
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.298][15][debug][pool] [source/common/conn_pool/conn_pool_base.cc:145] creating a new connection (connecting=0)
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.298][15][debug][http2] [source/common/http/http2/codec_impl.cc:1605] [C3] updating connection-level initial window size to 268435456
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.298][15][debug][connection] [./source/common/network/connection_impl.h:98] [C3] current connecting state: true
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.298][15][debug][client] [source/common/http/codec_client.cc:57] [C3] connecting
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.298][15][debug][connection] [source/common/network/connection_impl.cc:941] [C3] connecting to 192.168.65.2:8123
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.299][15][debug][connection] [source/common/network/connection_impl.cc:960] [C3] connection in progress
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.299][15][debug][http2] [source/common/http/http2/codec_impl.cc:1350] [C2] stream 1 closed: 0
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.299][15][debug][http2] [source/common/http/http2/codec_impl.cc:1414] [C2] Recouping 0 bytes of flow control window for stream 1.
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.306][15][debug][connection] [source/common/network/connection_impl.cc:688] [C3] connected
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.307][15][debug][client] [source/common/http/codec_client.cc:88] [C3] connected
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.307][15][debug][pool] [source/common/conn_pool/conn_pool_base.cc:328] [C3] attaching to next stream
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.307][15][debug][pool] [source/common/conn_pool/conn_pool_base.cc:182] [C3] creating stream
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.307][15][debug][router] [source/common/router/upstream_request.cc:550] [C1][S10104674368241947611] pool ready
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.307][15][debug][client] [source/common/http/codec_client.cc:141] [C3] encode complete
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.314][15][debug][router] [source/common/router/router.cc:1434] [C1][S10104674368241947611] upstream headers complete: end_stream=false
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.314][15][debug][http] [source/common/http/conn_manager_impl.cc:1700] [C1][S10104674368241947611] encoding headers via codec (end_stream=false):
envoy-grpc-example-envoy-1     | ':status', '200'
envoy-grpc-example-envoy-1     | 'content-type', 'application/grpc'
envoy-grpc-example-envoy-1     | 'x-envoy-upstream-service-time', '15'
envoy-grpc-example-envoy-1     | 'date', 'Fri, 02 Jun 2023 02:51:15 GMT'
envoy-grpc-example-envoy-1     | 'server', 'envoy'
envoy-grpc-example-envoy-1     | 
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.314][15][debug][client] [source/common/http/codec_client.cc:128] [C3] response complete
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.314][15][debug][pool] [source/common/conn_pool/conn_pool_base.cc:215] [C3] destroying stream: 0 remaining
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.314][15][debug][http] [source/common/http/conn_manager_impl.cc:1730] [C1][S10104674368241947611] encoding trailers via codec:
envoy-grpc-example-envoy-1     | 'grpc-status', '0'
envoy-grpc-example-envoy-1     | 'grpc-message', ''
envoy-grpc-example-envoy-1     | 
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.314][15][debug][http] [source/common/http/conn_manager_impl.cc:1805] [C1][S10104674368241947611] Codec completed encoding stream.
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.314][15][debug][http2] [source/common/http/http2/codec_impl.cc:1350] [C1] stream 1 closed: 0
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.314][15][debug][http2] [source/common/http/http2/codec_impl.cc:1414] [C1] Recouping 0 bytes of flow control window for stream 1.
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.314][15][debug][http2] [source/common/http/http2/codec_impl.cc:1350] [C3] stream 1 closed: 0
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.314][15][debug][http2] [source/common/http/http2/codec_impl.cc:1414] [C3] Recouping 0 bytes of flow control window for stream 1.
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.320][15][debug][connection] [source/common/network/connection_impl.cc:656] [C1] remote close
envoy-grpc-example-envoy-1     | [2023-06-02 02:51:15.320][15][debug][connection] [source/common/network/connection_impl.cc:250] [C1] closing socket: 0
```