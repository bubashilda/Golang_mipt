//go:build !solution

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	for _, url := range os.Args[1:] {
		resp, err := http.Get(url)
		check(err)
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		check(err)
		fmt.Printf("%s", body)
	}
}
