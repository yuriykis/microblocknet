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
	k, err := hex.DecodeString(key)
	if err != nil {
		panic(err)
	}
	return PrivateKeyFromBytes(k)
}

func (p *PrivateKey) PublicKey() *PublicKey {
	k := p.key.Public().(ed25519.PublicKey)
	paddedKey := make([]byte, PublicKeyLength)
	copy(paddedKey, k)

	return &PublicKey{
		key: paddedKey,
	}
}

func (p *PrivateKey) Sign(message string) *Signature {
	msg := []byte(message)
	return &Signature{
		value: ed25519.Sign(p.key, msg),
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

func PublicKeyFromString(key string) *PublicKey {
	b := []byte(key)
	return PublicKeyFromBytes(b)
}

func (p *PublicKey) Verify(message string, sig *Signature) bool {
	msg := []byte(message)
	return ed25519.Verify(p.key, msg, sig.value)
}

func (p *PublicKey) Address() *Address {
	return &Address{
		value: p.key[PublicKeyLength-AddressLength:],
	}
}

func (p *PublicKey) Bytes() []byte {
	return p.key
}

func (p *PublicKey) String() string {
	return hex.EncodeToString(p.key)
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

func SignatureFromString(value string) *Signature {
	b := []byte(value)
	return SignatureFromBytes(b)
}

func (s *Signature) Bytes() []byte {
	return s.value
}

func (s *Signature) String() string {
	return hex.EncodeToString(s.value)
}

func (s *Signature) Verify(message string, pubKey *PublicKey) bool {
	return pubKey.Verify(message, s)
}

type Address struct {
	value []byte
}

func (a *Address) String() string {
	return hex.EncodeToString(a.value)
}

func (a *Address) Bytes() []byte {
	return a.value
}
