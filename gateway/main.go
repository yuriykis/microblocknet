package main

import (
	"context"
	"fmt"
	"log"

	apiclient "github.com/yuriykis/microblocknet/node/service/api_client"
)

func main() {
	// listenAddr := flag.String("listen-addr", ":6000", "The address to listen on for incoming HTTP requests")
	// flag.Parse()

	// nodesAddrs := []string{"node1:3000", "node2:3001", "node3:3002"}
	apiClient := apiclient.NewHTTPClient("http://localhost:4001")
	res, err := apiClient.GetBlockByHeight(context.Background(), 0)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Println(res.Block)
}
