package types

import (
	"crypto/sha256"

	"github.com/yuriykis/microblocknet/crypto"
	"github.com/yuriykis/microblocknet/proto"
	pb "google.golang.org/protobuf/proto"
)

func transactionToHashable(tx *proto.Transaction) *proto.Transaction {
	// hash transaction without signatures in inputs
	for _, input := range tx.Inputs {
		input.Signature = nil
	}
	return tx
}

func HashTransaction(tx *proto.Transaction) string {
	b, err := pb.Marshal(transactionToHashable(tx))
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(b)
	return string(hash[:])
}

func SignTransaction(tx *proto.Transaction, privKey *crypto.PrivateKey) *crypto.Signature {
	return privKey.Sign(HashTransaction(tx))
}

func VerifyTransaction(tx *proto.Transaction) bool {
	for _, input := range tx.Inputs {
		sig := crypto.SignatureFromBytes(input.Signature)
		pubKey := crypto.PublicKeyFromBytes(input.PublicKey)
		if !pubKey.Verify(HashTransaction(tx), sig) {
			return false
		}
	}
	return true
}
