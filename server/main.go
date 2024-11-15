package main

import (
	"fmt"
	"net/http"
)

func main() {
	server := &Server{}
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", server.helloHandler)
	http.ListenAndServe(":8080", mux)
}

type Server struct{}

func (s *Server) helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}
