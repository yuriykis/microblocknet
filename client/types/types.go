package types

import "github.com/yuriykis/microblocknet/node/proto"

type GetBlockByHeightRequest struct {
	Height int
}

type GetBlockByHeightResponse struct {
	Block *proto.Block
}
