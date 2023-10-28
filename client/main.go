package main

import (
	"context"
	"fmt"

	"github.com/yuriykis/microblocknet/common/crypto"
	"github.com/yuriykis/microblocknet/common/requests"
	"github.com/yuriykis/microblocknet/common/types"
	gateway "github.com/yuriykis/microblocknet/gateway/client"
	"github.com/yuriykis/microblocknet/node/secure"
)

const godSeed = "41b84a2eff9a47393471748fbbdff9d20c14badab3d2de59fd8b5e98edd34d1c577c4c3515c6c19e5b9fdfba39528b1be755aae4d6a75fc851d3a17fbf51f1bc"

func main() {
	myKey := crypto.PrivateKeyFromString(godSeed)
	myAddr := myKey.PublicKey().Address()
	receiverAdd := crypto.GeneratePrivateKey().PublicKey().Address()

	bc := newBlockchainClient(gateway.NewHTTPClient("http://localhost:6000"))
	t := &types.Transaction{
		FromAddress: myAddr.Bytes(),
		ToAddress:   receiverAdd.Bytes(),
		Amount:      100,
	}
	tResp, err := bc.InitTransaction(context.Background(), t)
	if err != nil {
		panic(err)
	}
	tx := tResp.Transaction
	secure.SignTransaction(tx, myKey)
	fmt.Printf("tx: %+v\n", tx)
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
	t *types.Transaction,
) (*requests.CreateTransactionResponse, error) {
	req := requests.CreateTransactionRequest{
		FromAddress: t.FromAddress,
		ToAddress:   t.ToAddress,
		Amount:      t.Amount,
	}
	return bc.client.InitTransaction(ctx, req)
}
