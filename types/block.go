package types

import (
	"crypto/sha256"

	"github.com/yuriykis/microblocknet/crypto"
	"github.com/yuriykis/microblocknet/proto"
	pb "google.golang.org/protobuf/proto"
)

func HashBlock(block *proto.Block) []byte {
	// hash only block header
	b, err := pb.Marshal(block.Header)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(b)
	return hash[:]
}

func SignBlock(block *proto.Block, privKey *crypto.PrivateKey) *crypto.Signature {
	sig := privKey.Sign(HashBlock(block))
	block.Signature = sig.String()
	block.PublicKey = privKey.PublicKey().String()
	return sig
}

func VerifyBlock(block *proto.Block) bool {
	sig := crypto.SignatureFromString(block.Signature)
	pubKey := crypto.PublicKeyFromString(block.PublicKey)
	return pubKey.Verify(HashBlock(block), sig)
}
