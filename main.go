package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/yuriykis/microblocknet/crypto"
	"github.com/yuriykis/microblocknet/node"
	"github.com/yuriykis/microblocknet/proto"
	"github.com/yuriykis/microblocknet/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	go sendTransaction(n1, 5)

	// go stop(n1, grpcServer1, 10)
	// go stop(n2, grpcServer2, 30)

	select {}
}

func stop(n *node.NetNode, server node.Server, duration time.Duration) {
	time.Sleep(duration * time.Second)
	n.Stop(server)
}

func sendTransaction(n *node.NetNode, duration time.Duration) {
	time.Sleep(duration * time.Second)
	makeTransaction(n.ListenAddress)
}

func makeTransaction(endpoint string) {
	conn, err := grpc.Dial(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	c := proto.NewNodeClient(conn)

	myPrivKey := crypto.GeneratePrivateKey()
	myAddress := myPrivKey.PublicKey().Address()

	receiverAddress := crypto.GeneratePrivateKey().PublicKey().Address()

	txInput := &proto.TxInput{
		PrevTxHash: util.RandomHash(),
		PublicKey:  myPrivKey.PublicKey().Bytes(),
	}
	txOutput1 := &proto.TxOutput{
		Value:   100,
		Address: receiverAddress.Bytes(),
	}
	txOutput2 := &proto.TxOutput{
		Value:   900,
		Address: myAddress.Bytes(),
	}
	tx := &proto.Transaction{
		Inputs:  []*proto.TxInput{txInput},
		Outputs: []*proto.TxOutput{txOutput1, txOutput2},
	}
	ctx := context.Background()
	c.NewTransaction(ctx, tx)
}
