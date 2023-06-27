#!/bin/bash

ENVOY_SERVICE_NAME="envoy"
DOCKER_NETWORK_NAME="mesh-traffic"
docker stop $(docker ps -aq)
docker rm $(docker ps -aq)
docker service rm $ENVOY_SERVICE_NAME
docker network rm $DOCKER_NETWORK_NAME
docker network rm docker_gwbridge
docker swarm leave -f

docker system prune -a