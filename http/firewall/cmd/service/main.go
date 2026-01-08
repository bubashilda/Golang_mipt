package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	port := flag.String("port", "", "port to listen")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := io.Copy(w, r.Body)
		if err != nil {
			log.Printf("Error copying body: %v", err)
		}
		defer r.Body.Close()

		if cont_len := r.Header.Get("Content-Length"); cont_len != "" {
			w.Header().Set("Content-Length", cont_len)
		}
		if cont_type := r.Header.Get("Content-Type"); cont_type != "" {
			w.Header().Set("Content-Type", cont_type)
		}
	})

	fmt.Printf("Starting test service on port %s\n", *port)
	fmt.Printf("Service ready at http://localhost:%s\n", *port)

	_ = http.ListenAndServe(":"+*port, nil)
}
