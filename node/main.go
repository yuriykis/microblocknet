package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yuriykis/microblocknet/node/service"
)

const (
	defaultListenAddr = ":4000"
	defaultAPIAddr    = ":8000"
	defaultGateway    = "http://localhost:6000"
	defaultConsulAddr = "127.0.0.1:10000"
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
		gatewayAddress    = os.Getenv("GATEWAY_ADDR")
		consulServiceAddr = os.Getenv("CONSUL_SERVICE_ADDR")
		bootstrapNodesVar = os.Getenv("BOOTSTRAP_NODES")
		isMinerStr        = os.Getenv("IS_MINER")
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

	if consulServiceAddr == "" {
		consulServiceAddr = defaultConsulAddr
	}

	isMiner, err := strconv.ParseBool(isMinerStr)
	if err != nil {
		log.Fatal(err)
	}

	if gatewayAddress == "" {
		gatewayAddress = "http://localhost:6000"
	}

	n := service.New(listenAddr, apiListenAddr, gatewayAddress, consulServiceAddr)

	log.Fatal(n.Start(context.TODO(), bootstrapNodes, isMiner))
}

// for debugging
func debug() {

	var (
		n1 = service.New(":4000", ":8000", "http://localhost:6000", "127.0.0.1:10000")
		n2 = service.New(":4001", ":8001", "http://localhost:6000", "127.0.0.1:10001")
		n3 = service.New(":4002", ":8002", "http://localhost:6000", "127.0.0.1:10002")
		n4 = service.New(":4003", ":8003", "http://localhost:6000", "127.0.0.1:10003")
	)

	go func() {
		err := n1.Start(context.TODO(), []string{}, true)
		if err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		err := n2.Start(context.TODO(), []string{":4000"}, false)
		if err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		err := n3.Start(context.TODO(), []string{":4000"}, false)
		if err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		err := n4.Start(context.TODO(), []string{":4000"}, false)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// go sendTransaction(n1, 5, 0, 99900)
	// go sendTransaction(n2, 20, 1, 98000)

	// go stop(n1, grpcServer1, 10)
	// go stop(n2, grpcServer2, 30)

	select {}
}

func stop(n service.Service, duration time.Duration) {
	time.Sleep(duration * time.Second)
	n.Stop(context.TODO())
}

// func sendTransaction(n service.Service, duration time.Duration, height int, currentValue int64) {
// 	time.Sleep(duration * time.Second)
// 	makeTransaction(":3000", n, height, currentValue)
// }

// func makeTransaction(endpoint string, n service.Service, height int, currentValue int64) {
// 	conn, err := grpc.Dial(
// 		endpoint,
// 		grpc.WithTransportCredentials(insecure.NewCredentials()),
// 	)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	c := proto.NewNodeClient(conn)
// 	defer conn.Close()

// 	// myPrivKey := crypto.GeneratePrivateKey()
// 	myPrivKey := crypto.PrivateKeyFromString(godSeed)
// 	myAddress := myPrivKey.PublicKey().Address()

// 	receiverAddress := crypto.GeneratePrivateKey().PublicKey().Address()

// 	// TODO: we ned to wait for the previous block to be mined
// 	prevBlock, err := n.GetBlockByHeight(context.TODO(), height)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	prevBlockTx := prevBlock.GetTransactions()[len(prevBlock.GetTransactions())-1]
// 	myUTXOs, err := n.GetUTXOsByAddress(
// 		context.TODO(),
// 		crypto.PrivateKeyFromString(godSeed).PublicKey().Address().Bytes(),
// 	)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	txInput := &proto.TxInput{
// 		PrevTxHash: []byte(secure.HashTransaction(prevBlockTx)),
// 		PublicKey:  myPrivKey.PublicKey().Bytes(),
// 		OutIndex:   myUTXOs[0].OutIndex,
// 	}
// 	txOutput1 := &proto.TxOutput{
// 		Value:   100,
// 		Address: receiverAddress.Bytes(),
// 	}
// 	txOutput2 := &proto.TxOutput{
// 		Value:   currentValue,
// 		Address: myAddress.Bytes(),
// 	}
// 	tx := &proto.Transaction{
// 		Inputs:  []*proto.TxInput{txInput},
// 		Outputs: []*proto.TxOutput{txOutput1, txOutput2},
// 	}
// 	sig := secure.SignTransaction(tx, myPrivKey)
// 	tx.Inputs[0].Signature = sig.Bytes()

// 	ctx := context.Background()
// 	c.NewTransaction(ctx, tx)
// }
