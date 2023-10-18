package util

import (
	randc "crypto/rand"
	"encoding/hex"
	"io"
	"math/rand"
	"time"

	"github.com/yuriykis/microblocknet/proto"
)

func RandomHash() string {
	hash := make([]byte, 32)
	io.ReadFull(randc.Reader, hash)
	return hex.EncodeToString(hash)
}

func RandomBlock() *proto.Block {
	header := &proto.Header{
		Height:        int32(rand.Intn(1000)),
		PrevBlockHash: RandomHash(),
		MerkleRoot:    RandomHash(),
		Timestamp:     time.Now().UnixNano(),
	}
	return &proto.Block{
		Header: header,
	}
}
