package server

import (
	"fmt"

	"github.com/yuriykis/microblocknet/node/middleware"
)

func StartGRPCTransport(n middleware.NodeServer) error {
	s, ok := n.(*middleware.GRPCNodeServer)
	if !ok {
		return fmt.Errorf("invalid GRPCNodeServer")
	}
	return s.Serve()
}

func StopGRPCTransport(n middleware.NodeServer) error {
	s, ok := n.(*middleware.GRPCNodeServer)
	if !ok {
		return fmt.Errorf("invalid GRPCNodeServer")
	}
	s.Stop()
	return nil
}
