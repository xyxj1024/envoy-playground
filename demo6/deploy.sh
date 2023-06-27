#!/bin/bash

docker swarm init

DOCKER_NETWORK_NAME="mesh-traffic"
DOCKER_NETWORK_ID=$(docker network create --driver=overlay --attachable $DOCKER_NETWORK_NAME)
#DOCKER_BRIDGE_ID=$(docker network ls | grep docker_gwbridge | awk '{print $1}')
#DOCKER_BRIDGE_IP=$(docker network inspect $DOCKER_BRIDGE_ID | grep Gateway | grep -o -E '[0-9.]+')
ENVOY_SERVICE_NAME="envoy"

docker image build --tag envoy-mesh:v1 \
    #--build-arg CONTROL_PLANE_HOST=$DOCKER_BRIDGE_IP \
    $(pwd)/example/x86_64-darwin/envoy
docker service create \
    --publish mode=host,target=80,published=80 \
    --publish mode=host,target=443,published=443 \
    --publish mode=host,target=18000,published=18000 \
    --label envoy.endpoint.port=8080 \
    --label envoy.endpoint.timeout=30m \
    --label envoy.route.domain=google.com \
    --label envoy.route.path=/ \
    --network $DOCKER_NETWORK_NAME \
    --name $ENVOY_SERVICE_NAME envoy-mesh:v1