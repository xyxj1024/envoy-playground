package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

func blackServer(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(rw, "black")
}
func whiteServer(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(rw, "white")
}

func main() {
	l1, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatal("listen error:", err)
	}
	l2, err := net.Listen("tcp", ":8083")
	if err != nil {
		log.Fatal("listen error:", err)
	}
	go http.Serve(l1, http.HandlerFunc(blackServer))
	go http.Serve(l2, http.HandlerFunc(whiteServer))

	select {}
}
