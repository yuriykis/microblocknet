package main

import (
	"fmt"
	"time"

	"github.com/yuriykis/microblocknet/node"
)

func main() {
	n1 := makeNode(":3000", []string{})
	time.Sleep(1 * time.Second)
	n2 := makeNode(":3001", []string{":3000"})
	time.Sleep(1 * time.Second)
	n3 := makeNode(":3002", []string{":3000"})
	time.Sleep(1 * time.Second)
	n4 := makeNode(":3003", []string{":3001"})
	time.Sleep(1 * time.Second)
	n5 := makeNode(":3004", []string{":3002"})
	time.Sleep(5 * time.Second)

	fmt.Println(n1.Peers())
	fmt.Println(n2.Peers())
	fmt.Println(n3.Peers())
	fmt.Println(n4.Peers())
	fmt.Println(n5.Peers())

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
