# https://hub.docker.com/_/golang/tags?page=1&name=alpine
FROM golang:1.20.6-alpine3.18

EXPOSE 8080

WORKDIR /app
COPY app_red.go ./

CMD ["go", "run", "app_red.go"]