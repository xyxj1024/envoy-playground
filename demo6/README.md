## Demo Setup

### Development Environment

```bash
$ set | grep "MACHTYPE"
MACHTYPE=x86_64-apple-darwin21

$ go version
go version go1.20.5 darwin/amd64

$ envoy --version | grep -o '[0-9].[0-9]\+.[0-9]' | tail -n1
1.26.2
```

### How It Works

![System design details](./charts/envoy-swarm-control.svg)

The `ingress` network is a special overlay network that facilitates load balancing among a service's node. When any swarm node receives a request on a published port, it hands that request off to the `IPVS` module, which keeps track of all the IP addresses participating in that service, selects one of them, and routes the request to it, over the `ingress` network.

The `docker_gwbridge` is a bridge network that connects the overlay networks (including the `ingress` network) to an individual Docker daemon's physical network. By default, each container a service is running is connected to its local Docker daemon host's `docker_gwbridge` network.

## Run Code

```bash
go run envoy-swarm-control --debug --cert-dir $(pwd)/example/x86_64-darwin/cert

docker service update \
    --label-add envoy.endpoint.port=80 \
    --label-add envoy.route.domain=example.com \
    --label-add envoy.route.upstream-host=www.google.com \
    envoy-1

docker service update \
    --label-add envoy.endpoint.port=80 \
    --label-add envoy.route.domain=example.com \
    --label-add envoy.route.upstream-host=www.wustl.edu \
    envoy-2
```