## Demo Setup

The demo code is mostly copied from [https://github.com/kanurag94/envoy-sds](https://github.com/kanurag94/envoy-sds).

## Run Code

```bash
# Generate certificates
cd cert
bash cert.sh && cd ..

# Run tests
go test envoy-sds/sds
```

At the current stage, I can observe the following error messages:

```console
--- FAIL: TestService_FetchSecrets_MTLS (20.01s)
    --- FAIL: TestService_FetchSecrets_MTLS/ok (20.00s)
        service_mtls_test.go:97: Service.FetchSecrets() error: rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing: dial tcp [2606:2800:220:1:248:1893:25c8:1946]:50051: i/o timeout", succeeded: false
    --- FAIL: TestService_FetchSecrets_MTLS/ok_multiple (0.00s)
        service_mtls_test.go:97: Service.FetchSecrets() error: rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing: dial tcp [2606:2800:220:1:248:1893:25c8:1946]:50051: i/o timeout", succeeded: false
```