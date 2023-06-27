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

![System design details](envoy-swarm-control.drawio.svg)

The `ingress` network is a special overlay network that facilitates load balancing among a service's node. When any swarm node receives a request on a published port, it hands that request off to the `IPVS` module, which keeps track of all the IP addresses participating in that service, selects one of them, and routes the request to it, over the `ingress` network.

The `docker_gwbridge` is a bridge network that connects the overlay networks (including the `ingress` network) to an individual Docker daemon's physical network. By default, each container a service is running is connected to its local Docker daemon host's `docker_gwbridge` network.

## Run Code

```bash
$ docker swarm init

$ DOCKER_NETWORK_NAME="mesh-traffic"
$ docker network create --driver=overlay --attachable $DOCKER_NETWORK_NAME
$ DOCKER_BRIDGE_ID=$(docker network ls | grep docker_gwbridge | awk '{print $1}')
$ DOCKER_BRIDGE_IP=$(docker network inspect $DOCKER_BRIDGE_ID | grep Gateway | grep -o -E '[0-9.]+')
$ ENVOY_SERVICE_NAME="envoy"

# Build and deploy the Dockerized Envoy service:
$ docker image build --tag envoy-mesh:v1 \
    --build-arg CONTROL_PLANE_HOST=$DOCKER_BRIDGE_IP \
    $(pwd)/example/x86_64-darwin/envoy
$ docker service create \
    --publish mode=host,target=80,published=80 \
    --publish mode=host,target=443,published=443 \
    --publish mode=host,target=18000,published=18000 \
    --network $DOCKER_NETWORK_NAME \
    --name $ENVOY_SERVICE_NAME envoy-mesh:v1
```

```bash
go run envoy-swarm-control --debug --cert-dir $(pwd)/example/x86_64-darwin/cert

docker service logs -f $ENVOY_SERVICE_NAME
```