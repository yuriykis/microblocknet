package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePrivateKey(t *testing.T) {
	privKey := GeneratePrivateKey()
	assert.NotNil(t, privKey)
	assert.Equal(t, PrivateKeyLength, len(privKey.key))

	pubKey := privKey.PublicKey()
	assert.NotNil(t, pubKey)
	assert.Equal(t, PublicKeyLength, len(pubKey.key))
}

func TestPrivateKeyFromBytes(t *testing.T) {
	privKey := GeneratePrivateKey()
	assert.NotNil(t, privKey)
	assert.Equal(t, PrivateKeyLength, len(privKey.key))

	privKey2 := PrivateKeyFromBytes(privKey.key)
	assert.NotNil(t, privKey2)
	assert.Equal(t, PrivateKeyLength, len(privKey2.key))
	assert.Equal(t, privKey.key, privKey2.key)
}

func TestPrivateKeyFromString(t *testing.T) {
	privKey := GeneratePrivateKey()
	assert.NotNil(t, privKey)
	assert.Equal(t, PrivateKeyLength, len(privKey.Bytes()))

	privKey2 := PrivateKeyFromString(privKey.String())
	assert.NotNil(t, privKey2)
	assert.Equal(t, PrivateKeyLength, len(privKey2.Bytes()))
	assert.Equal(t, privKey.key, privKey2.key)
}

func TestPrivateKeySign(t *testing.T) {
	privKey := GeneratePrivateKey()
	assert.NotNil(t, privKey)
	assert.Equal(t, PrivateKeyLength, len(privKey.Bytes()))

	msg := "Hello, World!"
	sig := privKey.Sign(msg)
	assert.NotNil(t, sig)
	assert.Equal(t, SignatureLength, len(sig.Bytes()))
}
