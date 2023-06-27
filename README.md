# Service Mesh Examples with Envoy and Docker

Spring 2023 @Washington University in St. Louis

> Kelsey Hightower: "I might build a little prototype, I might build something at the hello-world level, and then I'd like to share it. So I'll just put all of my notes like on GitHub, right? Like the whole concept of the README. I just learn about service mesh, here's a little prototype that I built to crystallize what I've learned, and it's going to throw it here on GitHub so others can like check my notes or maybe you can learn from it, too." ([The ReadME Podcast, Episode 30](https://github.com/readme/podcast/kelsey-hightower))

Repo description:
- Demo 1: Envoy monitoring with Prometheus and Grafana
- Demo 2: Observing containerized Envoy with `bpftrace` programs
- Demo 3: Envoy dynamic configuration "hello world" (single Envoy instance)
- Demo 4: gRPC communication with Envoy xDS-based global load balancing
- Demo 5: Dockerized gRPC communication with Envoy external authorization

## Useful Links

- [Envoy data plaine API's `envoy` package](https://pkg.go.dev/github.com/envoyproxy/go-control-plane/envoy)
- [Docker engine documentation](https://docs.docker.com/config/labels-custom-metadata/)
- [Container-level and service-level labels](https://docs.docker.com/compose/compose-file/compose-file-v3/)
- [Open source projects built on Envoy proxy](https://www.envoyproxy.io/community.html)
- [Jordan Webb](https://jordemort.dev/), "[The container orchestrator landscape](https://lwn.net/Articles/905164/)," August 23, 2022.
- Viktor Adam, "[Podlike](https://blog.viktoradam.net/2018/05/14/podlike/)," May 14, 2018.
- Hechao Li, "[Linux Bridge - Part 1](https://hechao.li/2017/12/13/linux-bridge-part1/)," December 13, 2017.
- Karl Matthias, "[Sidecar: Service Discovery for all Docker Environments](https://relistan.com/sidecar-service-discovery-for-all-docker-environments)," August 04, 2016.
- [Lyft: Using Envoy as an Explicit `CONNECT` and Transparent Proxy to disrupt malicious traffic, 11/02/2022](https://eng.lyft.com/internet-egress-filtering-of-services-at-lyft-72e99e29a4d9)
- [Cloudflare: How to build your own public key infrastructure, 06/24/2015](https://blog.cloudflare.com/how-to-build-your-own-public-key-infrastructure/)
- [Cloudflare: Moving k8s communication to gRPC, 03/20/2021](https://blog.cloudflare.com/moving-k8s-communication-to-grpc/)

## Appendix: Some notes taken along the way

### Securing application-to-application communication

A certificate lets a website or service prove its identity. Practically speaking, a certificate is a file with some identity information about the owner, a public key, and a signature from a certificate authority (CA). Each certificate also contains a public key. Each public key has an associated private key, which is kept securely under the certificate owner's control. The private key can be used to create digital signatures that can be verified by the associated public key.

A certificate typically contains:
- Information about the organization that the certificate is issued to
- A public key
- Information about the organization that issued the certificate
- The rights granted by the issuer
- The validity period for the certificate
- Which hostnames the certificate is valid for
- The allowed uses (client authentication, server authentication)
- A digital signature by the issuer certificate's private key

The fact that the certificate is itself digitally signed by a third party CA means that if the verifier trusts the third party, they have assurances that the certificate is legitimate. The CA can give a certificate certain rights, such as a period of time in which the identity of the certificate should be trusted. Sometimes certificates are signed by what's called an intermediate CA, which is itself signed by a different CA. In this case, a certificate verifier can follow the chain until they find a certificate that they trust &mdash; the root.

This chain of trust model can be very useful for the CA. It allows the root certificate's private key to be kept offline and only used for signing intermediate certificates. Intermediate CA certificates can be shorter lived and be used to sign endpoint certificates on demand. Shorter-lived online intermediates are easier to manage and revoke if compromised.

### Containers and port forwarding

```bash
# Create an NGINX container,
$ docker run -d --name nginx-1 nginx
# Whose IP address should only exists inside the Linux VM started by Docker Desktop,
$ CONT_IP=$(
    docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' nginx-1
)
# Thus not routable from the host system.
$ ping $CONT_IP
ping $CONT_IP
PING 172.17.0.2 (172.17.0.2): 56 data bytes
Request timeout for icmp_seq 0
Request timeout for icmp_seq 1
Request timeout for icmp_seq 2
Request timeout for icmp_seq 3
Request timeout for icmp_seq 4
^C
--- 172.17.0.2 ping statistics ---
6 packets transmitted, 0 packets received, 100.0% packet loss
# The following request will not be responded:
$ curl $CONT_IP:80

# Create a regular port forwarding from the host's 0.0.0.0:8080 to $CONT_IP:80
$ docker run -d --publish 8080:80 --name nginx-2 nginx
$ sudo lsof -i -P | grep LISTEN | grep :8080
com.docke 29407 xuanyuanxingjian   47u  IPv6 0x47db63421ff81dc3      0t0    TCP *:8080 (LISTEN)
$ ps 29407
  PID   TT  STAT      TIME COMMAND
29407   ??  S      0:13.53 /Applications/Docker.app/Contents/MacOS/com.docker.backend -watchdog -native-api
$ curl localhost:8080
# On Linux, try:
$ sudo iptables -t nat -L
```

### Expose single port on host for multiple containers with the `SO_REUSEPORT` socket option

This is an experiment borrowed from [this blog post](https://iximiuz.com/en/posts/multiple-containers-same-port-reverse-proxy/) by Ivan Velichko.

Starting from Linux 3.9, one can bind an arbitrary number of sockets to exactly the same interface-port pair as long as all of them use the [`SO_REUSEPORT`](https://lwn.net/Articles/542629/) socket option. Check out the [`http_server.go`](https://github.com/xyxj1024/envoy-playground/blob/main/docker-networking-with-go/multiple-containers-same-port/http_server.go) program file. The `docker build -t http_server .` command roughly took 80s to finish executing on my Mac machine.

```bash
# Prepare the sandbox
$ docker run -d --rm \
> --name app_sandbox \
> --publish 80:8080 \
> alpine sleep infinity

# Run first application container
$ docker run -d --rm \
> --network container:app_sandbox \
> --env INSTANCE=foo \
> --env HOST=0.0.0.0 \
> --env PORT=8080 \
> http_server

# Run second application container
$ docker run -d --rm \
> --network container:app_sandbox \
> --env INSTANCE=bar \
> --env HOST=0.0.0.0 \
> --env PORT=8080 \
> http_server

# Send requests to application containers
$ for i in {1..300}; do curl -s $(ipconfig getifaddr en0) 2>&1; done | sort | uniq -c
 160 Hello from bar
 140 Hello from foo

# List containers
$ docker ps
CONTAINER ID   IMAGE         COMMAND                  CREATED          STATUS          PORTS                  NAMES
4b458ce106e3   http_server   "go run http_server.…"   13 minutes ago   Up 13 minutes                          relaxed_murdock
284c2a328f74   http_server   "go run http_server.…"   13 minutes ago   Up 13 minutes                          hungry_nightingale
985420e2c4e6   alpine        "sleep infinity"         14 minutes ago   Up 14 minutes   0.0.0.0:80->8080/tcp   app_sandbox

# Check listening TCP sockets (none on port 8080)
$ sudo lsof -i -P | grep LISTEN | grep :80
launchd       1             root   32u  IPv6 0x47db63421b014fc3      0t0    TCP localhost:8021 (LISTEN)
launchd       1             root   33u  IPv4 0x47db634bb327f723      0t0    TCP localhost:8021 (LISTEN)
launchd       1             root   35u  IPv6 0x47db63421b014fc3      0t0    TCP localhost:8021 (LISTEN)
launchd       1             root   36u  IPv4 0x47db634bb327f723      0t0    TCP localhost:8021 (LISTEN)
com.docke 29407 xuanyuanxingjian   82u  IPv6 0x47db63421ff847c3      0t0    TCP *:80 (LISTEN)
```

### gRPC API architecture

HTTP REST APIs generally use JSON for their request and response format. [Protocol Buffer](https://protobuf.dev/) is the native request/response format of gRPC because it has a standard schema agreed upon by the client and server during registration. Once a schema is defined, the `protoc` command can be used to generate code for [many languages](https://grpc.io/docs/languages/). Protocol Buffer data is structured as messages, with each message containing information stored in the form of fields. The fields are strongly typed, providing type safety unlike JSON or XML.

Often overlooked from a developer's perspective, HTTP client libraries are clunky and require code that defines paths, handles parameters, and deals with responses in bytes. gRPC abstracts all of this away and makes network calls feel like any other function calls defined for a `struct`. gRPC can easily stream data between client and server and is commonly used in microservice architecture.

gRPC lets you decide between four types of service methods:
- **Unary**: client sends a single request to the server and gets a single response back, just like a normal function call.
- **Server Streaming**: server returns a stream of messages in response to a client's request.
- **Client Streaming**: client sends a stream of messages to the server and the server replies in a single message, usually once the client has finished streaming.
- **Bi-directional Streaming**: the client and server can both send streams of messages to each other asynchronously.

### Service discovery in a microservices architecture

[Blog post by Igor Kolomiyets](https://itnext.io/enable-services-auto-discovery-in-docker-swarm-in-15-minutes-ae30f3877dc8): Enable auto-discovery of Docker swarm services using [Registrator](https://github.com/gliderlabs/registrator), [Consul](https://github.com/hashicorp/consul), and [Rotor](https://github.com/turbinelabs/rotor) (already shut down).

In a modern, cloud-based microservices architecture, services instances have dynamically assigned network locations; moreover, the set of service instances changes dynamically because of autoscaling, failures, and upgrades.

How do clients of a service (in the case of [client-side discovery](https://microservices.io/patterns/client-side-discovery.html)) and/or routers (in the case of [server-side discovery](https://microservices.io/patterns/server-side-discovery.html)) know about the available instances of a service? Implement a [service registry](https://microservices.io/patterns/service-registry.html) or service discovery registry, which is a database of services, their instances and their locations. Service instances are registered with the service registry on startup and deregistered on shutdown. Client of the service and/or routers query the service registry to find the available instances of a service.

#### [Traefik](https://traefik.io/traefik/): The cloud native application proxy

Labels of a Docker container can be accessed through the following structure:

```go
type Container struct {
    ID         string `json:"Id"`
    Names      []string
    Image      string
    ImageID    string
    Command    string
    Created    int64
    Ports      []Port
    SizeRw     int64 `json:",omitempty"`
    SizeRootFs int64 `json:",omitempty"`
    Labels     map[string]string
    State      string
    Status     string
    HostConfig struct {
        NetworkMode string `json:",omitempty"`
    }
    NetworkSettings *SummaryNetworkSettings
    Mounts          []MountPoint
}
```

which is defined in [this Go module](https://pkg.go.dev/github.com/docker/docker).

Configuration discovery in [Traefik](https://github.com/traefik/traefik) is achieved through *Providers*. The providers are infrastructure components, whether orchestrators, container engines, cloud providers, or key-value stores. The idea is that Traefik queries the provider APIs in order to find relevant information about routing, and when Traefik detects a change, it dynamically updates the routes.

When using Docker as a [provider](https://doc.traefik.io/traefik/providers/overview/), Traefik uses container labels to retrieve its [routing](https://doc.traefik.io/traefik/routing/providers/docker) configuration. By default, Traefik watches for container-level labels on a standalone Docker Engine. When using Docker compose, labels are specified by the directive `labels` from the "[services](https://docs.docker.com/compose/compose-file/compose-file-v3/#service-configuration-reference)" objects. While in Swarm Mode, Traefik uses labels found on services, not on individual containers. Therefore, if you use a compose file with Swarm Mode, labels should be defined in the [`deploy`](https://docs.docker.com/compose/compose-file/compose-file-v3/#labels-1) part of your service.

#### [Registrator](https://gliderlabs.github.io/registrator/latest/): Service registry bridge for Docker

A service is anything listening on a port:

```go
type Service strut {
    ID    string
    Name  string
    IP    string
    Port  int
    Tags  []string
    Attrs map[string]string
}
```

The fields `ID`, `Name`, `Tags`, and `Attrs` can be overridden by user-defined container metadata stored as environment variables or labels.

#### [Sidecar](https://github.com/newrelic/sidecar): A dynamic service discovery platform

Sidecar works at the level of services and has the means of mapping containers to service endpoints. It has a lifecycle for services and it exchanges that information regularly with peers.

Sidecar uses a SWIM-based gossip protocol (derived from that used in HashiCorp's Serf) to communicate with peers and exchange service information on an ongoing basis.

Each host keeps its own copy of the shared state used to configure a local proxy, which listens locally and binds well known ports for each service.

Services become known not by their hostname, but by their `ServicePort`. This is a common pattern in modern distributed systems.