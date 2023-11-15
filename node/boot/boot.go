package boot

import (
	"github.com/yuriykis/microblocknet/node/middleware"
	"github.com/yuriykis/microblocknet/node/server"
	"github.com/yuriykis/microblocknet/node/service"
)

type BootOpts struct {
	BootstrapNodes []string
	IsMiner        bool
}

func BootNode(opts BootOpts, n *service.Node, apiServer *server.ApiNodeServer, grpcServer middleware.NodeServer) error {

	nodeOpts := service.NodeOpts{
		BootstrapNodes: opts.BootstrapNodes,
		IsMiner:        opts.IsMiner,
	}
	n.Start(nodeOpts)

	go server.StartApiTrasport(apiServer)

	return server.StartGRPCTransport(grpcServer)
}

func StopNode(n *service.Node, apiServer *server.ApiNodeServer, grpcServer middleware.NodeServer) error {
	if err := server.StopGRPCTransport(grpcServer); err != nil {
		return err
	}
	if err := server.StopApiTransport(apiServer); err != nil {
		return err
	}
	n.Stop()
	return nil
}
