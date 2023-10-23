package types

import (
	"bytes"
	"crypto/sha256"

	"github.com/cbergoon/merkletree"
	"github.com/yuriykis/microblocknet/crypto"
	"github.com/yuriykis/microblocknet/proto"
	pb "google.golang.org/protobuf/proto"
)

const blockHashDifficulty = 1

func HashBlock(block *proto.Block) string {
	return HashHeader(block.Header)
}

func HashHeader(header *proto.Header) string {
	b, err := pb.Marshal(header)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(b)
	return string(hash[:])
}

func SignBlock(block *proto.Block, privKey *crypto.PrivateKey) *crypto.Signature {
	if len(block.GetTransactions()) > 0 {
		t, err := makeMerkleTree(block)
		if err != nil {
			panic(err)
		}
		block.Header.MerkleRoot = t.MerkleRoot()
	}
	sig := privKey.Sign(HashBlock(block))
	block.Signature = sig.Bytes()
	block.PublicKey = privKey.PublicKey().Bytes()
	return sig
}

func VerifyBlock(block *proto.Block) bool {
	if len(block.GetTransactions()) > 0 {
		if !VerifyMerkleTree(block) {
			return false
		}
	}
	sig := crypto.SignatureFromBytes(block.Signature)
	pubKey := crypto.PublicKeyFromBytes(block.PublicKey)
	return pubKey.Verify(HashBlock(block), sig)
}

func VerifyBlockHash(block *proto.Block) bool {
	hash := HashBlock(block)
	return hash[:blockHashDifficulty] == string(bytes.Repeat([]byte{0}, blockHashDifficulty))
}

func VerifyMerkleTree(block *proto.Block) bool {
	hash := block.Header.MerkleRoot
	t, err := makeMerkleTree(block)
	if err != nil {
		return false
	}
	return bytes.Equal(t.MerkleRoot(), hash)
}

type MerkleTreeContent struct {
	Transaction *proto.Transaction
}

func (m *MerkleTreeContent) CalculateHash() ([]byte, error) {
	return []byte(HashTransaction(m.Transaction)), nil
}

func (m *MerkleTreeContent) Equals(other merkletree.Content) (bool, error) {
	return HashTransaction(m.Transaction) == HashTransaction(other.(*MerkleTreeContent).Transaction), nil
}

func makeMerkleTree(block *proto.Block) (*merkletree.MerkleTree, error) {
	txs := block.GetTransactions()
	treeContent := make([]merkletree.Content, len(txs))
	for i, tx := range txs {
		treeContent[i] = &MerkleTreeContent{Transaction: tx}
	}
	t, err := merkletree.NewTree(treeContent)
	if err != nil {
		return nil, err
	}
	return t, nil
}
