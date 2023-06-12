#!/bin/sh

# cp /usr/local/Cellar/envoy/1.25.1/share/envoy/configs/envoy-demo.yaml .
docker volume create --driver local --opt type=debugfs --opt device=debugfs debugfs
docker build -t ebpf-envoy:v1 -f ./Dockerfile.tools --build-arg OS_TAG=20.04 .
docker run --rm -it \
    --privileged \
    --cap-add=ALL \
    --pid=host \
    --env BPFTRACE_STRLEN=200 \
    --name ebpf-envoy \
    -v /lib/modules:/lib/modules:ro \
    -v debugfs:/sys/kernel/debug:rw \
    ebpf-envoy:v1

# docker run -d --name envoy -p 9901:9901 -p 10000:10000 envoyproxy/envoy:v1.25-latest