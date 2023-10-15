package main

import (
	"time"

	"github.com/yuriykis/microblocknet/node"
)

func main() {
	makeNode(":3000", []string{})
	time.Sleep(1 * time.Second)
	makeNode(":3001", []string{":3000"})
	time.Sleep(1 * time.Second)
	makeNode(":3002", []string{":3001"})
	// time.Sleep(1 * time.Second)
	// makeNode(":3003", []string{":3000", ":3001", ":3002"})

	select {}
}

func makeNode(listenAddr string, bootstrapNodes []string) *node.Node {
	n := node.NewNode(listenAddr)
	go func() {
		if err := n.Start(bootstrapNodes); err != nil {
			panic(err)
		}
	}()
	return n
}
