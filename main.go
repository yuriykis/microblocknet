package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/yuriykis/microblocknet/node"
)

const defaultListenAddr = ":3000"

func main() {

	if os.Getenv("DEBUG") != "" {
		debug()
		return
	}

	var (
		listenAddr        = os.Getenv("LISTEN_ADDR")
		bootstrapNodesVar = os.Getenv("BOOTSTRAP_NODES")
		bootstrapNodes    []string
	)
	if listenAddr == "" {
		listenAddr = defaultListenAddr
	}

	if bootstrapNodesVar != "" {
		bootstrapNodes = strings.Split(bootstrapNodesVar, ",")
	}

	n := node.New(listenAddr)
	grcpServer := node.MakeGRPCTransport(listenAddr, n)
	log.Fatal(n.Start(listenAddr, bootstrapNodes, grcpServer))
}

// for debugging
func debug() {

	var (
		n1 = node.New(":3000")
		n2 = node.New(":3001")
		n3 = node.New(":3002")
		n4 = node.New(":3003")

		grpcServer1 = node.MakeGRPCTransport(n1.ListenAddress, n1)
		grpcServer2 = node.MakeGRPCTransport(n2.ListenAddress, n2)
		grpcServer3 = node.MakeGRPCTransport(n3.ListenAddress, n3)
		grpcServer4 = node.MakeGRPCTransport(n4.ListenAddress, n4)
	)

	go n1.Start(n1.ListenAddress, []string{}, grpcServer1)
	go n2.Start(n2.ListenAddress, []string{":3000"}, grpcServer2)
	go n3.Start(n3.ListenAddress, []string{":3000"}, grpcServer3)
	go n4.Start(n4.ListenAddress, []string{":3001"}, grpcServer4)

	go func(stopTime int, server node.Server) {
		time.Sleep(time.Duration(stopTime) * time.Second)
		n1.Stop(server)
	}(10, grpcServer1)

	select {}
}
