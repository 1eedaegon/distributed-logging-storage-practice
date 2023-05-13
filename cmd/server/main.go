package main

import (
	"log"

	"github.com/1eedaegon/distributed-logging-storage-practice/internal/server"
)

func main() {
	srv := server.NewHTTPServer(":8080")
	log.Fatal(srv.ListenAndServe())
}
