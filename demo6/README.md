## Demo Setup

> Christian Posta: "The point is, you can build a workflow and process that statically configures the parts you need while use dynamic xDS services to discover the pieces you need at runtime. One of the reasons why you see different control-plane implementation is not everyone has a fully dynamic and fungible environment where all of the pieces should be dynamic. Adopt the level of dynamism that's most appropriate for your system given the existing constraints and available workflows."

### Development Environment

```bash
$ set | grep "MACHTYPE"
MACHTYPE=x86_64-apple-darwin22

$ docker version --format '{{.Server.Version}}'
24.0.2

$ go version
go version go1.20.6 darwin/amd64

$ envoy --version | grep -o '[0-9].[0-9]\+.[0-9]' | tail -n1
1.26.2
```

### How It Works

The control plane instance runs on the manager node of a Docker swarm cluster.

Some [writeup](https://xyxj1024.github.io/blog/a-control-plane-for-containerized-envoy-proxies) for this demo.

## Run Code

Deploy, and then update selected Envoy service(s):

```bash
# Create overlay network and attached swarm services
bash $(pwd)/deploy/scripts/deploy.sh

# Build and run control plane
go build

go run envoy-swarm-control --debug \
    --xds-port 18000 \
    --ingress-network mesh-traffic

# Update envoy-1
docker service update \
    --label-add envoy.status.node-id=local_node_1 \
    --label-add envoy.listener.port=10000 \
    --label-add envoy.endpoint.port=8080 \
    --label-add envoy.route.path=/ \
    --label-add envoy.route.upstream-host=app-1 \
    envoy-1

# Update envoy-2
docker service update \
    --label-add envoy.status.node-id=local_node_2 \
    --label-add envoy.listener.port=10000 \
    --label-add envoy.endpoint.port=80 \
    --label-add envoy.route.path=/ \
    --label-add envoy.route.upstream-host=www.wustl.edu \
    envoy-2
```

Let's take a look at how our example Envoy instances function:

```bash
# Check Envoy logs
docker service logs -f envoy-1

# Access envoy-1
curl -i http://localhost:8001/

# Access envoy-2
curl -i http://localhost:8002/
```

Results:

![screenshot](images/screenshot.png)

To clean up:

```bash
bash $(pwd)/deploy/scripts/cleanup.sh
```