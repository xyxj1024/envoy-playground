#!/usr/bin/env bash

go run ../cmd/certgen/certgen.go

openssl rsa -in ./ca.key -check -noout
openssl rsa -in ./server.key -check -noout
openssl verify -CAfile ./ca.crt ./server.crt
openssl verify -CAfile ./ca.crt ./client.crt

openssl x509 -in ./server.crt -text
openssl x509 -in ./client.crt -text

openssl x509 -pubkey -in ./ca.crt -noout | openssl md5
openssl pkey -pubout -in ./ca.key | openssl md5

openssl x509 -pubkey -in ./server.crt -noout | openssl md5
openssl pkey -pubout -in ./server.key | openssl md5