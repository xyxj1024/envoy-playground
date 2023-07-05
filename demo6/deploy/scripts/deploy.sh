#!/bin/bash

docker swarm init

DEMO_BASEDIR=~/Desktop/demo6
DOCKER_NETWORK_NAME="mesh-traffic"
DOCKER_NETWORK_ID=$(docker network create --driver=overlay --attachable $DOCKER_NETWORK_NAME)
ENVOY_SERVICE_1_NAME="envoy-1" && ENVOY_SERVICE_2_NAME="envoy-2"
ENVOY_IMAGE_1_NAME="envoy-1:v1" && ENVOY_IMAGE_2_NAME="envoy-2:v1"

docker image build --tag $ENVOY_IMAGE_1_NAME $DEMO_BASEDIR/deploy/envoy/envoy-1 #--build-arg CONTROL_PLANE_HOST=$DOCKER_BRIDGE_IP
docker service create \
    --publish 9001:9001 \
    --publish 10001:18000 \
    --network $DOCKER_NETWORK_NAME \
    --name $ENVOY_SERVICE_1_NAME $ENVOY_IMAGE_1_NAME

docker image build --tag $ENVOY_IMAGE_2_NAME $DEMO_BASEDIR/deploy/envoy/envoy-2 #--build-arg CONTROL_PLANE_HOST=$DOCKER_BRIDGE_IP
docker service create \
    --publish 9002:9002 \
    --publish 10002:18000 \
    --network $DOCKER_NETWORK_NAME \
    --name $ENVOY_SERVICE_2_NAME $ENVOY_IMAGE_2_NAME