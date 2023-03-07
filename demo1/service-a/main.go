package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Service struct {
	Name, Port   string
	ResponseTime int
}

/*
type Service struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}
*/

func (service Service) handler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Hello from service A!\n")
	time.Sleep(time.Duration(service.ResponseTime) * time.Millisecond)

	// Call service-b
	fmt.Fprintf(w, "Calling Service B: ")
	// We're using the http.DefaultClient
	req, err := http.NewRequest("GET", "http://service-a-envoy:8788/", nil)
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	// Add headers: https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers.html
	req.Header.Add("x-request-id", r.Header.Get("x-request-id"))           // uniquely identify a request as well as perform stable access logging and tracing
	req.Header.Add("x-b3-traceid", r.Header.Get("x-b3-traceid"))           // used by Zipkin tracer
	req.Header.Add("x-b3-spanid", r.Header.Get("x-b3-spanid"))             // used by Zipkin tracer
	req.Header.Add("x-b3-parentspanid", r.Header.Get("x-b3-parentspanid")) // used by Zipkin tracer
	req.Header.Add("x-b3-sampled", r.Header.Get("x-b3-sampled"))           // used by Zipkin tracer
	req.Header.Add("x-b3-flags", r.Header.Get("x-b3-flags"))               // used by Zipkin tracer
	req.Header.Add("x-ot-span-context", r.Header.Get("x-ot-span-context")) // establish proper parent-child relationships between tracing spans when used with the LightStep tracer

	// Create a client (no control parameters specified)
	// See: https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
	/*
		c := &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
						Timeout:   30 * time.Second,
						KeepAlive: 30 * time.Second,
				}).Dial,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			}
		}
	*/
	client := &http.Client{}
	// Do method sends an HTTP request and returns an HTTP response
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	// Close the connection
	defer resp.Body.Close()
	var body1 []byte
	if resp.StatusCode == 200 {
		body1, err = ioutil.ReadAll(resp.Body)
	}
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	fmt.Fprintf(w, "%s", string(body1))

	// Similarly we call service-c
	fmt.Fprintf(w, "Calling Service C: ")
	req, err = http.NewRequest("GET", "http://service-a-envoy:8791/", nil)
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	req.Header.Add("x-request-id", r.Header.Get("x-request-id"))
	req.Header.Add("x-b3-traceid", r.Header.Get("x-b3-traceid"))
	req.Header.Add("x-b3-spanid", r.Header.Get("x-b3-spanid"))
	req.Header.Add("x-b3-parentspanid", r.Header.Get("x-b3-parentspanid"))
	req.Header.Add("x-b3-sampled", r.Header.Get("x-b3-sampled"))
	req.Header.Add("x-b3-flags", r.Header.Get("x-b3-flags"))
	req.Header.Add("x-ot-span-context", r.Header.Get("x-ot-span-context"))

	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	defer resp.Body.Close()
	var body2 []byte
	if resp.StatusCode == 200 {
		body2, err = ioutil.ReadAll(resp.Body)
	}
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	fmt.Fprintf(w, "%s", string(body2))

	/*
		w.Header().Set("Content-Type", "text/html;charset=utf-8")
		t, err := template.ParseFiles("/Users/xuanyuanxingjian/Documents/meng_program/cse-master-project/Envoy/demo3/template.html")
		if err != nil {
			log.Fatal("Unable to load template:", err)
		}
		service := Service{
			Name:    "Service A",
			Message: "Hello World!",
		}
		err = t.Execute(w, service)
		if err != nil {
			log.Fatal("Unable to execute template:", err)
		}
	*/
}

func main() {
	var serviceName = flag.String("s", "", "Name of this Service")
	var port = flag.String("p", "", "HTTP port to listen to")
	var responseTime = flag.Int("r", 0, "Time in ms to wait before response")
	flag.Parse()

	service := Service{*serviceName, *port, *responseTime}
	http.HandleFunc("/", service.handler)
	log.Fatal(http.ListenAndServe(":8081", nil))

	fmt.Printf("%v listening on port: %v, Press <Enter to exit>\n", service.Name, service.Port)
	fmt.Scanln()
}
