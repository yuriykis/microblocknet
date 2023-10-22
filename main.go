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
	"github.com/yuriykis/microblocknet/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultListenAddr = ":3000"

const godSeed = "41b84a2eff9a47393471748fbbdff9d20c14badab3d2de59fd8b5e98edd34d1c577c4c3515c6c19e5b9fdfba39528b1be755aae4d6a75fc851d3a17fbf51f1bc"

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
	log.Fatal(n.Start(listenAddr, bootstrapNodes, grcpServer, false))
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

	go n1.Start(n1.ListenAddress, []string{}, grpcServer1, true)
	go n2.Start(n2.ListenAddress, []string{":3000"}, grpcServer2, false)
	go n3.Start(n3.ListenAddress, []string{":3000"}, grpcServer3, false)
	go n4.Start(n4.ListenAddress, []string{":3001"}, grpcServer4, false)

	go sendTransaction(n1, 3, 0, 99000)
	go sendTransaction(n2, 10, 1, 98000)
	// go stop(n1, grpcServer1, 10)
	// go stop(n2, grpcServer2, 30)

	select {}
}

func stop(n *node.NetNode, server node.Server, duration time.Duration) {
	time.Sleep(duration * time.Second)
	n.Stop(server)
}

func sendTransaction(n *node.NetNode, duration time.Duration, height int, currentValue int64) {
	time.Sleep(duration * time.Second)
	makeTransaction(n.ListenAddress, n, height, currentValue)
}

func makeTransaction(endpoint string, n *node.NetNode, height int, currentValue int64) {
	conn, err := grpc.Dial(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	c := proto.NewNodeClient(conn)
	defer conn.Close()

	// myPrivKey := crypto.GeneratePrivateKey()
	myPrivKey := crypto.PrivateKeyFromString(godSeed)
	myAddress := myPrivKey.PublicKey().Address()

	receiverAddress := crypto.GeneratePrivateKey().PublicKey().Address()

	// TODO: we ned to wait for the previous block to be mined
	prevBlock, err := n.GetBlockByHeight(height)
	if err != nil {
		log.Fatal(err)
	}
	prevBlockTx := prevBlock.GetTransactions()[len(prevBlock.GetTransactions())-1]
	myUTXOs, err := n.GetUTXOsByAddress(crypto.PrivateKeyFromString(godSeed).PublicKey().Address().Bytes())
	if err != nil {
		log.Fatal(err)
	}

	txInput := &proto.TxInput{
		PrevTxHash: []byte(types.HashTransaction(prevBlockTx)),
		PublicKey:  myPrivKey.PublicKey().Bytes(),
		OutIndex:   myUTXOs[0].OutIndex,
	}
	txOutput1 := &proto.TxOutput{
		Value:   100,
		Address: receiverAddress.Bytes(),
	}
	txOutput2 := &proto.TxOutput{
		Value:   currentValue,
		Address: myAddress.Bytes(),
	}
	tx := &proto.Transaction{
		Inputs:  []*proto.TxInput{txInput},
		Outputs: []*proto.TxOutput{txOutput1, txOutput2},
	}
	sig := types.SignTransaction(tx, myPrivKey)
	tx.Inputs[0].Signature = sig.Bytes()

	ctx := context.Background()
	c.NewTransaction(ctx, tx)
}
