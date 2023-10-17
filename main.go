package main

import (
	"fmt"
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
	log.Fatal(n.Start(listenAddr, bootstrapNodes))
}

// for debugging
func debug() {
	var (
		n1ch = make(chan *node.NetNode)
		n2ch = make(chan *node.NetNode)
		n3ch = make(chan *node.NetNode)
		n4ch = make(chan *node.NetNode)
	)
	go start(":3000", []string{}, n1ch)
	go start(":3001", []string{":3000"}, n2ch)
	go start(":3002", []string{":3000"}, n3ch)
	go start(":3003", []string{":3001"}, n4ch)

	for {
		select {
		case n1 := <-n1ch:
			fmt.Printf("Node %s, peers: %v\n", n1, n1.Peers())
		case n2 := <-n2ch:
			fmt.Printf("Node %s, peers: %v\n", n2, n2.Peers())
		case n3 := <-n3ch:
			fmt.Printf("Node %s, peers: %v\n", n3, n3.Peers())
		case n4 := <-n4ch:
			fmt.Printf("Node %s, peers: %v\n", n4, n4.Peers())
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
	go func() {
		time.Sleep(15 * time.Second)
		nodech <- n
	}()
	return n.Start(listenAddr, bootstrapNodes)
}

// TODO:
func stop() {

}
