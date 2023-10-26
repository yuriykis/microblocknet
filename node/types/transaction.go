package types

import (
	"crypto/sha256"

	"github.com/yuriykis/microblocknet/common/crypto"
	"github.com/yuriykis/microblocknet/common/proto"
	pb "google.golang.org/protobuf/proto"
)

func transactionToHashable(tx *proto.Transaction) *proto.Transaction {
	// we need to copy all tx fields as they are the pointers
	// and we don't want to change the original tx
	// we copy all fields except signature
	txNoSig := &proto.Transaction{
		Inputs:  make([]*proto.TxInput, len(tx.Inputs)),
		Outputs: make([]*proto.TxOutput, len(tx.Outputs)),
	}
	for i, input := range tx.Inputs {
		txNoSig.Inputs[i] = &proto.TxInput{
			PublicKey:  input.PublicKey,
			PrevTxHash: input.PrevTxHash,
			OutIndex:   input.OutIndex,
		}
	}
	for i, output := range tx.Outputs {
		txNoSig.Outputs[i] = &proto.TxOutput{
			Value:   output.Value,
			Address: output.Address,
		}
	}
	return txNoSig
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
