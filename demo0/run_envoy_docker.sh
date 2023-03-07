#!/bin/sh
docker pull bitnami/envoy:latest
docker run --rm -p 0.0.0.0:10000:10000 -p 0.0.0.0:9001:9001 -v "$PWD/envoy-config.yaml":/opt/bitnami/envoy/conf/envoy.yaml --name envoy -t bitnami/envoy