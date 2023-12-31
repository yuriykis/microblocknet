package secure

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuriykis/microblocknet/common/crypto"
	"github.com/yuriykis/microblocknet/node/util"
)

func TestHashBlock(t *testing.T) {
	block := util.RandomBlock()
	hash := HashBlock(block)
	if len(hash) != 32 {
		t.Errorf("HashBlock() = %v, want %v", len([]byte(hash)), 32)
	}
}
func TestSignBlock(t *testing.T) {
	block := util.RandomBlock()
	privKey := crypto.GeneratePrivateKey()
	pubKey := privKey.PublicKey()
	sig := SignBlock(block, privKey)
	assert.True(t, sig.Verify(HashBlock(block), pubKey))
}
