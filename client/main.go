package main

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/yuriykis/microblocknet/common/crypto"
	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/common/requests"
	gateway "github.com/yuriykis/microblocknet/gateway/client"
	"github.com/yuriykis/microblocknet/node/secure"
)

const godSeed = "41b84a2eff9a47393471748fbbdff9d20c14badab3d2de59fd8b5e98edd34d1c577c4c3515c6c19e5b9fdfba39528b1be755aae4d6a75fc851d3a17fbf51f1bc"

func main() {
	myKey := crypto.PrivateKeyFromString(godSeed)
	myPubKey := myKey.PublicKey()
	myAddr := myPubKey.Address()
	receiverAdd := crypto.GeneratePrivateKey().PublicKey().Address()

	bc := newBlockchainClient(gateway.NewHTTPClient("http://localhost:6000"))
	t := &Transaction{
		FromAddress: myAddr.Bytes(),
		FromPubKey:  myPubKey.Bytes(),
		ToAddress:   receiverAdd.Bytes(),
		Amount:      100,
	}
	tResp, err := bc.InitTransaction(context.Background(), t)
	if err != nil {
		log.Fatal(err)
	}
	tx := tResp.Transaction
	sig := secure.SignTransaction(tx, myKey)
	tx.Inputs[0].Signature = sig.Bytes()
	txRes, err := bc.NewTransaction(context.Background(), tx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(txRes.Transaction)
}

type blockchainClient struct {
	client gateway.Client
}

func newBlockchainClient(c gateway.Client) *blockchainClient {
	return &blockchainClient{
		client: c,
	}
}

func (bc *blockchainClient) InitTransaction(
	ctx context.Context,
	t *Transaction,
) (*requests.InitTransactionResponse, error) {
	req := requests.InitTransactionRequest{
		FromAddress: t.FromAddress,
		FromPubKey:  t.FromPubKey,
		ToAddress:   t.ToAddress,
		Amount:      t.Amount,
	}
	return bc.client.InitTransaction(ctx, req)
}

func (bc *blockchainClient) NewTransaction(
	ctx context.Context,
	t *proto.Transaction,
) (*requests.NewTransactionResponse, error) {
	req := requests.NewTransactionRequest{
		Transaction: t,
	}
	return bc.client.NewTransaction(ctx, req)
}
