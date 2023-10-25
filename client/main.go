package main

import (
	"context"
	"log"
	"time"

	"github.com/yuriykis/microblocknet/client/types"
	"github.com/yuriykis/microblocknet/gateway/client"
	"github.com/yuriykis/microblocknet/node/crypto"
	"github.com/yuriykis/microblocknet/node/proto"
	"github.com/yuriykis/microblocknet/node/service"
	nodetypes "github.com/yuriykis/microblocknet/node/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const godSeed = "41b84a2eff9a47393471748fbbdff9d20c14badab3d2de59fd8b5e98edd34d1c577c4c3515c6c19e5b9fdfba39528b1be755aae4d6a75fc851d3a17fbf51f1bc"

func main() {
	// go sendTransaction(n1, 3, 0, 99000)
	// go sendTransaction(n2, 10, 1, 98000)
	// go stop(n1, grpcServer1, 10)
	// go stop(n2, grpcServer2, 30)

	bc := newBlockchainClient(client.NewHTTPClient("http://localhost:6000"))
	res, err := bc.GetBlockByHeight(context.Background(), 1)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res.Block)
}

type blockchainClient struct {
	client client.Client
}

func newBlockchainClient(c client.Client) *blockchainClient {
	return &blockchainClient{
		client: c,
	}
}

func (bc *blockchainClient) GetBlockByHeight(ctx context.Context, height int) (*types.GetBlockByHeightResponse, error) {
	var req types.GetBlockByHeightRequest
	req.Height = height
	res, err := bc.client.GetBlockByHeight(context.Background(), req.Height)
	if err != nil {
		log.Fatal(err)
	}
	cRes := &types.GetBlockByHeightResponse{
		Block: res.Block,
	}
	return cRes, nil
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
		PrevTxHash: []byte(nodetypes.HashTransaction(prevBlockTx)),
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
	sig := nodetypes.SignTransaction(tx, myPrivKey)
	tx.Inputs[0].Signature = sig.Bytes()

	ctx := context.Background()
	c.NewTransaction(ctx, tx)
}
