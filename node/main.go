package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yuriykis/microblocknet/node/boot"
)

const (
	defaultListenAddr = ":4000"
	defaultAPIAddr    = ":8000"
	defaultGateway    = "http://localhost:6000"
	defaultConsulAddr = "127.0.0.1:10000"
	defaultStoreType  = "memory"
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
		storeType         = os.Getenv("STORE_TYPE")
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

	if storeType == "" {
		storeType = defaultStoreType
	}

	isMiner, err := strconv.ParseBool(isMinerStr)
	if err != nil {
		log.Fatal(err)
	}

	if gatewayAddress == "" {
		gatewayAddress = "http://localhost:6000"
	}

	nb := NewNodeBuilder(
		listenAddr,
		apiListenAddr,
		gatewayAddress,
		consulServiceAddr,
		bootstrapNodes,
		storeType,
		isMiner,
	)
	err = nb.Build()
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(boot.BootNode(nb.bootOpts, nb.node, nb.apiServer, nb.grpcServer))
}

// for debugging
func debug() {

	var (
		nb1 = NewNodeBuilder(
			"localhost:4000",
			"localhost:8000",
			"http://localhost:6000",
			"localhost:10000",
			[]string{},
			"mongo",
			true,
		)
		nb2 = NewNodeBuilder(
			"localhost:4001",
			"localhost:8001",
			"http://localhost:6000",
			"localhost:10001",
			[]string{"localhost:4000"},
			"mongo",
			false,
		)
		nb3 = NewNodeBuilder(
			"localhost:4002",
			"localhost:8002",
			"http://localhost:6000",
			"localhost:10002",
			[]string{"localhost:4000"},
			"mongo",
			false,
		)
		nb4 = NewNodeBuilder(
			"localhost:4003",
			"localhost:8003",
			"http://localhost:6000",
			"localhost:10003",
			[]string{"localhost:4000"},
			"mongo",
			false,
		)
	)
	for _, nb := range []*NodeBuilder{nb1, nb2, nb3, nb4} {
		err := nb.Build()
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, nb := range []*NodeBuilder{nb1, nb2, nb3, nb4} {
		go func(nb *NodeBuilder) {
			err := boot.BootNode(nb.bootOpts, nb.node, nb.apiServer, nb.grpcServer)
			if err != nil {
				log.Fatal(err)
			}
		}(nb)
	}

	// go sendTransaction(n1, 5, 0, 99900)
	// go sendTransaction(n2, 20, 1, 98000)

	// go stop(n1, grpcServer1, 10)
	// go stop(n2, grpcServer2, 30)

	select {}
}

func stop(nb *NodeBuilder, duration time.Duration) {
	time.Sleep(duration * time.Second)
	boot.StopNode(nb.node, nb.apiServer, nb.grpcServer)
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
