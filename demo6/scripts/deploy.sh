#!/bin/bash

docker swarm init

DOCKER_NETWORK_NAME="mesh-traffic"
DOCKER_NETWORK_ID=$(docker network create --driver=overlay --attachable $DOCKER_NETWORK_NAME)
#DOCKER_BRIDGE_ID=$(docker network ls | grep docker_gwbridge | awk '{print $1}')
#DOCKER_BRIDGE_IP=$(docker network inspect $DOCKER_BRIDGE_ID | grep Gateway | grep -o -E '[0-9.]+')
ENVOY_SERVICE_1_NAME="envoy-1" && ENVOY_SERVICE_2_NAME="envoy-2"
ENVOY_IMAGE_1_NAME="envoy-1:v1" && ENVOY_IMAGE_2_NAME="envoy-2:v1"

docker image build --tag $ENVOY_IMAGE_1_NAME $(pwd)/example/x86_64-darwin/envoy/envoy-1 #--build-arg CONTROL_PLANE_HOST=$DOCKER_BRIDGE_IP
docker service create \
    --publish 9001:9001 \
    --publish 10001:18000 \
    --network $DOCKER_NETWORK_NAME \
    --name $ENVOY_SERVICE_1_NAME $ENVOY_IMAGE_1_NAME

docker image build --tag $ENVOY_IMAGE_2_NAME $(pwd)/example/x86_64-darwin/envoy/envoy-2 #--build-arg CONTROL_PLANE_HOST=$DOCKER_BRIDGE_IP
docker service create \
    --publish 9002:9002 \
    --publish 10002:18000 \
    --network $DOCKER_NETWORK_NAME \
    --name $ENVOY_SERVICE_2_NAME $ENVOY_IMAGE_2_NAME