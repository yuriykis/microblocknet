package node

import (
	"fmt"
	"time"

	"github.com/yuriykis/microblocknet/crypto"
	"github.com/yuriykis/microblocknet/proto"
	"github.com/yuriykis/microblocknet/store"
	"github.com/yuriykis/microblocknet/types"
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

type Chain struct {
	txStore    store.TxStorer
	blockStore store.BlockStorer
	headers    *HeadersList
}

func NewChain(txStore store.TxStorer, blockStore store.BlockStorer) *Chain {
	chain := &Chain{
		txStore:    txStore,
		blockStore: blockStore,
		headers:    NewHeadersList(),
	}
	chain.addBlock(genesisBlock())
	return chain
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
	if err := c.blockStore.Put(block); err != nil {
		return err
	}
	c.headers.Add(block.Header)

	for _, tx := range block.Transactions {
		if err := c.txStore.Put(tx); err != nil {
			return err
		}
	}
	return c.blockStore.Put(block)
}

func (c *Chain) ValidateBlock(b *proto.Block) error {
	if !types.VerifyBlock(b) {
		return fmt.Errorf("block is not valid")
	}
	return nil
}

func (c *Chain) GetBlockByHeight(height int) (*proto.Block, error) {
	if height > c.Height() {
		return nil, fmt.Errorf("height %d is greater than chain height %d", height, c.Height())
	}
	return nil, nil // TODO: implement
}

func genesisBlock() *proto.Block {
	privKey := crypto.GeneratePrivateKey()

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

	types.SignBlock(firstBlock, privKey)

	return firstBlock
}
