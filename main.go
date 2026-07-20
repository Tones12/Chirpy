package main

import (
	"net/http"
	"fmt"
)

func main() {
	serveMux := http.NewServeMux()

	server := http.Server{
		Handler:	serveMux,
		Addr:		":8080",
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("Error: %s", err)
	}
}
