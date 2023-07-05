#!/bin/bash

ENVOY_SERVICE_1_NAME="envoy-1"
ENVOY_SERVICE_2_NAME="envoy-2"
DOCKER_NETWORK_NAME="mesh-traffic"
docker stop $(docker ps -aq)
docker rm $(docker ps -aq)
docker service rm $ENVOY_SERVICE_1_NAME
docker service rm $ENVOY_SERVICE_2_NAME
docker network rm $DOCKER_NETWORK_NAME
docker network rm docker_gwbridge
docker swarm leave -f

docker system prune -a