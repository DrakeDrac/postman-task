package main

import (
	"fmt"
	"io"
	"net/http"
)

func webServerTest(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got req")
	io.WriteString(w, "yoyo")
}

func main() {
	http.HandleFunc("/", webServerTest)

	http.ListenAndServe(":8080", nil)
}