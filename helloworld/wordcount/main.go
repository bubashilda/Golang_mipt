//go:build !solution

package main

import (
	"bufio"
	"fmt"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	counter := make(map[string]int)
	for _, file := range os.Args[1:] {
		f, err := os.Open(file)
		check(err)
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			counter[scanner.Text()]++
		}
		f.Close()
	}
	for key, value := range counter {
		if value > 1 {
			fmt.Printf("%d\t%s\n", value, key)
		}
	}
}
