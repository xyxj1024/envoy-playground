#!/bin/bash

docker swarm init

DEMO_BASEDIR=~/Documents/projects/Repos/github-envoy-playground/demo6

DOCKER_NETWORK_NAME="mesh-traffic"
DOCKER_NETWORK_ID=$(docker network create --driver=overlay --attachable $DOCKER_NETWORK_NAME)

ENVOY_SERVICE_1_NAME="envoy-1" && ENVOY_SERVICE_2_NAME="envoy-2"
ENVOY_IMAGE_1_NAME="envoy-1:v1" && ENVOY_IMAGE_2_NAME="envoy-2:v1"

# For each Envoy service, we might want three sets of port publishing rules:
# 1. administration server port
# 2. static listener port
# 3. xDS port for gRPC streaming
docker image build --tag $ENVOY_IMAGE_1_NAME $DEMO_BASEDIR/deploy/envoy/envoy-1
docker service create \
    --publish 9001:9001 \
    --publish 8001:10000 \
    --publish 10001:18000 \
    --network $DOCKER_NETWORK_NAME \
    --name $ENVOY_SERVICE_1_NAME $ENVOY_IMAGE_1_NAME

docker image build --tag $ENVOY_IMAGE_2_NAME $DEMO_BASEDIR/deploy/envoy/envoy-2
docker service create \
    --publish 9002:9002 \
    --publish 8002:10000 \
    --publish 10002:18000 \
    --network $DOCKER_NETWORK_NAME \
    --name $ENVOY_SERVICE_2_NAME $ENVOY_IMAGE_2_NAME

docker image build --tag app-1:v1 $DEMO_BASEDIR/deploy/app
docker service create \
    --publish 8080:8080 \
    --network $DOCKER_NETWORK_NAME \
    --name app-1 app-1:v1