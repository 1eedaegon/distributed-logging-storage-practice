package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	api "github.com/1eedaegon/distributed-logging-storage-practice/api/v1"
	"google.golang.org/grpc"
)

// Service discovery client
func main() {
	addr := flag.String("addr", ":8440", "service address")
	flag.Parse()
	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	client := api.NewLogClient(conn)
	ctx := context.Background()
	res, err := client.GetServers(ctx, &api.GetServersRequest{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("servers:")
	for _, server := range res.Servers {
		fmt.Printf("\t- %v\n", server)
	}
}
