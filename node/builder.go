package main

import (
	"github.com/yuriykis/microblocknet/node/boot"
	"github.com/yuriykis/microblocknet/node/middleware"
	"github.com/yuriykis/microblocknet/node/server"
	"github.com/yuriykis/microblocknet/node/service"
)

type NodeBuilder struct {
	serverConfig service.ServerConfig
	bootOpts     boot.BootOpts
	grpcServer   middleware.NodeServer
	apiServer    *server.ApiNodeServer
	node         *service.Node
}

func NewNodeBuilder(
	nodeListenAddr string,
	apiListenAddr string,
	gatewayAddr string,
	consulServiceAddr string,
	bootstrapNodes []string,
	storeType string,
	isMiner bool,
) *NodeBuilder {
	return &NodeBuilder{
		serverConfig: service.ServerConfig{
			NodeListenAddress:    nodeListenAddr,
			ApiListenAddr:        apiListenAddr,
			GatewayAddress:       gatewayAddr,
			ConsulServiceAddress: consulServiceAddr,
			StoreType:            storeType,
		},
		bootOpts: boot.BootOpts{
			BootstrapNodes: bootstrapNodes,
			IsMiner:        isMiner,
		},
	}
}

func (b *NodeBuilder) Build() error {
	var err error
	n := service.New(b.serverConfig)
	b.node = n
	b.apiServer, err = server.NewApiServer(
		b.serverConfig.NodeListenAddress,
		b.serverConfig.ApiListenAddr,
		n,
	)
	if err != nil {
		return err
	}
	b.grpcServer = middleware.NewGRPCNodeServer(
		n,
		b.serverConfig.NodeListenAddress,
	)
	return nil
}
