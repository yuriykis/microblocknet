package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/yuriykis/microblocknet/common/crypto"
	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/node/service"
	"github.com/yuriykis/microblocknet/node/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultListenAddr = ":3000"
	defaultAPIAddr    = ":4000"
)

const godSeed = "41b84a2eff9a47393471748fbbdff9d20c14badab3d2de59fd8b5e98edd34d1c577c4c3515c6c19e5b9fdfba39528b1be755aae4d6a75fc851d3a17fbf51f1bc"

func main() {

	if os.Getenv("DEBUG") != "" {
		debug()
		return
	}

	var (
		listenAddr        = os.Getenv("LISTEN_ADDR")
		apiListenAddr     = os.Getenv("API_LISTEN_ADDR")
		bootstrapNodesVar = os.Getenv("BOOTSTRAP_NODES")
		bootstrapNodes    []string
	)
	if listenAddr == "" {
		listenAddr = defaultListenAddr
	}
	if apiListenAddr == "" {
		apiListenAddr = defaultAPIAddr
	}

	if bootstrapNodesVar != "" {
		bootstrapNodes = strings.Split(bootstrapNodesVar, ",")
	}

	n := service.New(listenAddr, apiListenAddr)

	log.Fatal(n.Start(bootstrapNodes, false))
}

// for debugging
func debug() {

	var (
		n1 = service.New(":3000", ":4000")
		n2 = service.New(":3001", ":4001")
		n3 = service.New(":3002", ":4002")
		n4 = service.New(":3003", ":4003")
	)

	go n1.Start([]string{}, true)
	go n2.Start([]string{":3000"}, false)
	go n3.Start([]string{":3000"}, false)
	go n4.Start([]string{":3001"}, false)

	go sendTransaction(n1, 3, 0, 99000)
	// go sendTransaction(n2, 20, 1, 98000)

	// go stop(n1, grpcServer1, 10)
	// go stop(n2, grpcServer2, 30)

	select {}
}

func stop(n service.Service, duration time.Duration) {
	time.Sleep(duration * time.Second)
	n.Stop()
}

func sendTransaction(n service.Service, duration time.Duration, height int, currentValue int64) {
	time.Sleep(duration * time.Second)
	makeTransaction(":3000", n, height, currentValue)
}

func makeTransaction(endpoint string, n service.Service, height int, currentValue int64) {
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
