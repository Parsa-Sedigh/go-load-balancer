package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var (
	port = flag.Int("port", 8081, "port to start the demo service on")
)

type DemoServer struct {
}

func (d *DemoServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("All good from server %d.", *port)))
}

func main() {
	flag.Parse()

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), &DemoServer{}); err != nil {
		log.Fatal(err)
	}
}
