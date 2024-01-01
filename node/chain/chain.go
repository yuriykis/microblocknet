package chain

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/yuriykis/microblocknet/common/crypto"
	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/node/secure"
	"github.com/yuriykis/microblocknet/node/store"
)

type HeadersList struct {
	headers []*proto.Header
}

func NewHeadersList() *HeadersList {
	return &HeadersList{
		headers: make([]*proto.Header, 0),
	}
}

func (l *HeadersList) Height() int {
	return len(l.headers) - 1
}

func (l *HeadersList) Add(header *proto.Header) {
	l.headers = append(l.headers, header)
}

func (l *HeadersList) Get(index int) (*proto.Header, error) {
	if index > l.Height() {
		return nil, fmt.Errorf("index %d is greater than height %d", index, l.Height())
	}
	return l.headers[index], nil
}

// -----------------------------------------------------------------------------

const godSeed = "41b84a2eff9a47393471748fbbdff9d20c14badab3d2de59fd8b5e98edd34d1c577c4c3515c6c19e5b9fdfba39528b1be755aae4d6a75fc851d3a17fbf51f1bc"

type Chain struct {
	store   store.Storer
	headers *HeadersList
}

func New(s store.Storer) *Chain {
	chain := &Chain{
		store:   s,
		headers: NewHeadersList(),
	}
	chain.addBlock(genesisBlock())
	return chain
}

func (c *Chain) Store() store.Storer {
	return c.store
}

func (c *Chain) Height() int {
	return c.headers.Height()
}

func (c *Chain) AddBlock(block *proto.Block) error {
	if err := c.ValidateBlock(block); err != nil {
		return err
	}
	return c.addBlock(block)
}

func (c *Chain) addBlock(block *proto.Block) error {
	ctx := context.Background()
	if err := c.store.BlockStore(ctx).Put(ctx, block); err != nil {
		return err
	}
	c.headers.Add(block.Header)

	for _, tx := range block.Transactions {
		if err := c.store.TxStore(ctx).Put(ctx, tx); err != nil {
			return err
		}
		if err := c.makeUTXOs(tx); err != nil {
			return err
		}
	}
	return c.store.BlockStore(ctx).Put(ctx, block)
}

func (c *Chain) makeUTXOs(tx *proto.Transaction) error {
	ctx := context.Background()
	txHash := secure.HashTransaction(tx)
	for index, output := range tx.Outputs {
		utxo := &proto.UTXO{
			TxHash:   []byte(txHash),
			OutIndex: int32(index),
			Output:   output,
			Spent:    false,
		}
		if err := c.store.UTXOStore(ctx).Put(ctx, utxo); err != nil {
			return err
		}
	}
	for _, input := range tx.Inputs {
		utxoKey := secure.MakeUTXOKey(input.PrevTxHash, int(input.OutIndex))
		utxo, err := c.store.UTXOStore(ctx).Get(ctx, utxoKey)
		if err != nil {
			return err
		}
		utxo.Spent = true
		if err := c.store.UTXOStore(ctx).Put(ctx, utxo); err != nil {
			return err
		}
	}
	return nil
}

func (c *Chain) ValidateBlock(b *proto.Block) error {
	if !secure.VerifyBlock(b) {
		return fmt.Errorf("block is not valid")
	}

	currentBlock, err := c.GetBlockByHeight(c.Height())
	if err != nil {
		return err
	}
	if b.Header.Height != currentBlock.Header.Height+1 {
		return fmt.Errorf(
			"block height %d is not equal to current height %d + 1",
			b.Header.Height,
			currentBlock.Header.Height,
		)
	}
	currentBlockHash := secure.HashBlock(currentBlock)
	if !bytes.Equal(b.Header.PrevBlockHash, []byte(currentBlockHash)) {
		return fmt.Errorf(
			"block prev hash %s is not equal to current hash %s",
			b.Header.PrevBlockHash,
			currentBlockHash,
		)
	}
	for _, tx := range b.Transactions {
		if err := c.ValidateTransaction(tx); err != nil {
			return err
		}
	}
	return nil
}

func (c *Chain) ValidateTransaction(tx *proto.Transaction) error {
	ctx := context.Background()
	if !secure.VerifyTransaction(tx) {
		return fmt.Errorf("transaction is not valid")
	}
	inputsSum := int64(0)
	for _, input := range tx.Inputs {
		utxoKey := secure.MakeUTXOKey(input.PrevTxHash, int(input.OutIndex))
		utxo, err := c.store.UTXOStore(ctx).Get(ctx, utxoKey)
		if err != nil {
			return err
		}
		if utxo == nil {
			return fmt.Errorf("utxo %s not found", utxoKey)
		}
		if utxo.Spent {
			return fmt.Errorf("utxo %s is already spent", utxoKey)
		}
		inputsSum += utxo.Output.Value
	}
	outputsSum := int64(0)
	for _, output := range tx.Outputs {
		outputsSum += output.Value
	}
	if inputsSum < outputsSum {
		return fmt.Errorf("inputs sum %d is less than outputs sum %d", inputsSum, outputsSum)
	}
	return nil
}

func (c *Chain) GetBlockByHeight(height int) (*proto.Block, error) {
	ctx := context.Background()
	if height > c.Height() {
		return nil, fmt.Errorf("height %d is greater than chain height %d", height, c.Height())
	}
	header, err := c.headers.Get(height)
	if err != nil {
		return nil, err
	}
	hash := secure.HashHeader(header)
	block, err := c.store.BlockStore(ctx).Get(ctx, hash)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (c *Chain) GetBlockByHash(hash string) (*proto.Block, error) {
	ctx := context.Background()
	return c.store.BlockStore(ctx).Get(ctx, hash)
}

func genesisBlock() *proto.Block {
	privKey := crypto.PrivateKeyFromString(godSeed)

	firstBlock := &proto.Block{
		Header: &proto.Header{
			Height:    0,
			Timestamp: time.Now().Unix(),
		},
	}
	firstTx := &proto.Transaction{
		Inputs: []*proto.TxInput{},
		Outputs: []*proto.TxOutput{
			{
				Value:   100000,
				Address: privKey.PublicKey().Address().Bytes(),
			},
		},
	}
	firstBlock.Transactions = append(firstBlock.Transactions, firstTx)

	secure.SignBlock(firstBlock, privKey)

	return firstBlock
}
