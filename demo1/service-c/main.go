package main

import (
	"fmt"
	"log"
	"net/http"
)

/*
type Service struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}
*/

func handler(w http.ResponseWriter, r *http.Request) {
	/*
		w.Header().Set("Content-Type", "text/html;charset=utf-8")
		t, err := template.ParseFiles("/Users/xuanyuanxingjian/Documents/meng_program/cse-master-project/Envoy/demo3/template.html")
		if err != nil {
			log.Fatal("Unable to load template:", err)
		}
		service := Service{
			Name:    "Service C",
			Message: "Hello World!",
		}
		err = t.Execute(w, service)
		if err != nil {
			log.Fatal("Unable to execute template:", err)
		}
	*/
	fmt.Fprintf(w, "Hello from service C!\n")
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8083", nil))
}
