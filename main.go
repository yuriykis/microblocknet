package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/yuriykis/microblocknet/node"
	"github.com/yuriykis/microblocknet/proto"
	"google.golang.org/grpc"
)

const defaultListenAddr = ":3000"

func main() {
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
	if len(bootstrapNodes) > 0 {
		go func() {
			if err := n.BootstrapNetwork(bootstrapNodes); err != nil {
				log.Fatalf("NetNode: %s, failed to bootstrap network: %v", n, err)
			}
		}()
	}
	log.Fatal(makeGRPCTransport(n.ListenAddress, n))
}

func makeGRPCTransport(listenAddr string, svc node.Node) error {
	fmt.Printf("Node %s, starting GRPC transport\n", listenAddr)
	var (
		opt        = []grpc.ServerOption{}
		grpcServer = grpc.NewServer(opt...)
	)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()

	grpcNodeServer := node.NewGRPCNodeServer(svc)
	proto.RegisterNodeServer(grpcServer, grpcNodeServer)

	return grpcServer.Serve(ln)
}

// for debugging
func debug() {
	var (
		n1ch = make(chan *node.NetNode)
		n2ch = make(chan *node.NetNode)
	)
	go start(":3000", []string{}, n1ch)

	time.Sleep(1 * time.Second)
	go start(":3001", []string{":3000"}, n2ch)

	for {
		select {
		case n1 := <-n1ch:
			fmt.Printf("Node %s, peers: %v\n", n1, n1.Peers())
		case n2 := <-n2ch:
			fmt.Printf("Node %s, peers: %v\n", n2, n2.Peers())
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func start(listenAddr string, bootstrapNodes []string, nodech chan *node.NetNode) error {
	if listenAddr == "" {
		listenAddr = defaultListenAddr
	}

	n := node.New(listenAddr)
	if len(bootstrapNodes) > 0 {
		go func() {
			if err := n.BootstrapNetwork(bootstrapNodes); err != nil {
				log.Fatalf("NetNode: %s, failed to bootstrap network: %v", n, err)
			}
		}()
	}
	go func() {
		time.Sleep(5 * time.Second)
		nodech <- n
	}()
	return makeGRPCTransport(n.ListenAddress, n)
}

// func main() {
// 	n1 := makeNode(":3000", []string{})
// 	time.Sleep(1 * time.Second)
// 	n2 := makeNode(":3001", []string{":3000"})
// 	time.Sleep(1 * time.Second)
// 	n3 := makeNode(":3002", []string{":3000"})
// 	time.Sleep(1 * time.Second)
// 	n4 := makeNode(":3003", []string{":3001"})
// 	time.Sleep(1 * time.Second)
// 	n5 := makeNode(":3004", []string{":3002"})
// 	time.Sleep(5 * time.Second)

// 	fmt.Println(n1.Peers())
// 	fmt.Println(n2.Peers())
// 	fmt.Println(n3.Peers())
// 	fmt.Println(n4.Peers())
// 	fmt.Println(n5.Peers())

// 	select {}
// }

// func makeNode(listenAddr string, bootstrapNodes []string) *node.NetNode {
// 	n := node.NewNode(listenAddr)
// 	go func() {
// 		if err := n.Start(bootstrapNodes); err != nil {
// 			panic(err)
// 		}
// 	}()
// 	return n
// }
