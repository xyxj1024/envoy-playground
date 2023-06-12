# Service Mesh Examples with Envoy and Docker

Spring 2023 @Washington University in St. Louis

Repo description:
- Demo 1: Envoy monitoring with Prometheus and Grafana
- Demo 2: Observing containerized Envoy with `bpftrace` programs
- Demo 3: Envoy dynamic configuration "hello world" (single Envoy instance)
- Demo 4: gRPC communication with Envoy xDS-based global load balancing
- Demo 5: Dockerized gRPC communication with Envoy external authorization

## Appendix: Some notes taken along the way

### Docker documentations on labels

[Docker Engine Documentation](https://docs.docker.com/config/labels-custom-metadata/)

[Container-level and service-level labels](https://docs.docker.com/compose/compose-file/compose-file-v3/)

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

### Service discovery in a microservices architecture

In a modern, cloud-based microservices architecture, services instances have dynamically assigned network locations; moreover, the set of service instances changes dynamically because of autoscaling, failures, and upgrades.

How do clients of a service (in the case of [client-side discovery](https://microservices.io/patterns/client-side-discovery.html)) and/or routers (in the case of [server-side discovery](https://microservices.io/patterns/server-side-discovery.html)) know about the available instances of a service? Implement a [service registry](https://microservices.io/patterns/service-registry.html), which is a database of services, their instances and their locations. Service instances are registered with the service registry on startup and deregistered on shutdown. Client of the service and/or routers query the service registry to find the available instances of a service.

[Blog post by Igor Kolomiyets](https://itnext.io/enable-services-auto-discovery-in-docker-swarm-in-15-minutes-ae30f3877dc8)