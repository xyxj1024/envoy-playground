package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

func redServer(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(rw, "red")
}

func main() {
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("listen error:", err)
	}

	go http.Serve(l, http.HandlerFunc(redServer))

	select {}
}
