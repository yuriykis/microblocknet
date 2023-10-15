package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
)

const (
	PrivateKeyLength = 64
	PublicKeyLength  = 32
	SignatureLength  = 64
	SeedLength       = 32
	AddressLength    = 20
)

type PrivateKey struct {
	key ed25519.PrivateKey
}

func NewPrivateKey() *PrivateKey {
	return &PrivateKey{}
}

func (p *PrivateKey) Bytes() []byte {
	return p.key
}

func (p *PrivateKey) String() string {
	return hex.EncodeToString(p.key)
}

func GeneratePrivateKey() *PrivateKey {
	seed := make([]byte, SeedLength)
	_, err := rand.Read(seed)
	if err != nil {
		panic(err)
	}
	return &PrivateKey{
		key: ed25519.NewKeyFromSeed(seed),
	}
}

func PrivateKeyFromBytes(key []byte) *PrivateKey {
	paddedKey := make([]byte, PrivateKeyLength)
	copy(paddedKey, key)
	return &PrivateKey{
		key: ed25519.PrivateKey(paddedKey),
	}
}

func PrivateKeyFromString(key string) *PrivateKey {
	b, err := hex.DecodeString(key)
	if err != nil {
		panic(err)
	}
	return PrivateKeyFromBytes(b)
}

func (p *PrivateKey) PublicKey() *PublicKey {
	k := p.key.Public().(ed25519.PublicKey)
	paddedKey := make([]byte, PublicKeyLength)
	copy(paddedKey, k)

	return &PublicKey{
		key: paddedKey,
	}
}

func (p *PrivateKey) Sign(message []byte) *Signature {
	return &Signature{
		value: ed25519.Sign(p.key, message),
	}
}

type PublicKey struct {
	key ed25519.PublicKey
}

func PublicKeyFromBytes(key []byte) *PublicKey {
	paddedKey := make([]byte, PublicKeyLength)
	copy(paddedKey, key)
	return &PublicKey{
		key: paddedKey,
	}
}

func (p *PublicKey) Verify(message []byte, sig *Signature) bool {
	return ed25519.Verify(p.key, message, sig.value)
}

func (p *PublicKey) Address() *Address {
	return &Address{
		value: p.key[PublicKeyLength-AddressLength:],
	}
}

type Signature struct {
	value []byte
}

func SignatureFromBytes(value []byte) *Signature {
	valueCopy := make([]byte, SignatureLength)
	copy(valueCopy, value)
	return &Signature{
		value: valueCopy,
	}
}

func (s *Signature) Bytes() []byte {
	return s.value
}

type Address struct {
	value []byte
}

func (a *Address) Bytes() []byte {
	return a.value
}

func (a *Address) String() string {
	return string(a.value)
}