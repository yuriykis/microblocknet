package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yuriykis/microblocknet/common/crypto"
	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/node/secure"
	"github.com/yuriykis/microblocknet/node/service/client"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	grpcPeer "google.golang.org/grpc/peer"
)

const (
	miningInterval         = 5 * time.Second
	maxMiningDuration      = 10 * time.Second
	syncBlockchainInterval = 5 * time.Second
)

type Service interface {
	Start(ctx context.Context, bootstrapNodes []string, isMiner bool) error
	Stop(ctx context.Context) error
}

type Node interface {
	Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error)
	NewTransaction(ctx context.Context, t *proto.Transaction) (*proto.Transaction, error)
	NewBlock(ctx context.Context, b *proto.Block) (*proto.Block, error)
	GetBlocks(ctx context.Context, v *proto.Version) (*proto.Blocks, error)
}

type ServerConfig struct {
	Version       string
	ListenAddress string
	ApiListenAddr string
	PrivateKey    *crypto.PrivateKey
}
type node struct {
	ServerConfig

	logger *zap.SugaredLogger

	dr DataRetriever
	nm *networkManager

	isMiner bool

	transportServer TransportServer
	apiServer       ApiServer

	quitNode
}

type quitNode struct {
	showNodeInfoQuitCh   chan struct{}
	syncBlockchainQuitCh chan struct{}
}

func (n *node) shutdown() {
	close(n.showNodeInfoQuitCh)
	close(n.syncBlockchainQuitCh)
}

func New(listenAddress string, apiListenAddress string) Service {
	logger := makeLogger()
	return &node{
		ServerConfig: ServerConfig{
			Version:       "0.0.1",
			ListenAddress: listenAddress,
			ApiListenAddr: apiListenAddress,
			PrivateKey:    nil,
		},

		logger: logger,
		dr:     NewDataRetriever(),
		nm:     NewNetworkManager(listenAddress, logger),
	}
}

func (n *node) Start(ctx context.Context, bootstrapNodes []string, isMiner bool) error {

	n.nm.start(bootstrapNodes)
	go n.syncBlockchainLoop(n.syncBlockchainQuitCh)
	go n.showNodeInfo(n.showNodeInfoQuitCh, false, true)

	if isMiner {
		n.isMiner = isMiner
		n.PrivateKey = crypto.GeneratePrivateKey()
		go n.minerLoop()
	}

	n.transportServer = NewGRPCNodeServer(n, n.ListenAddress)
	go n.transportServer.Start()

	api, err := NewApiServer(n.dr, n.ListenAddress, n.ApiListenAddr)
	if err != nil {
		return err
	}
	n.apiServer = api
	return n.apiServer.Start(context.TODO())
}

func (n *node) Stop(ctx context.Context) error {
	n.shutdown()
	n.nm.stop()
	n.apiServer.Stop(context.TODO())
	return n.transportServer.Stop()
}

func (n *node) String() string {
	return n.ListenAddress
}

func (n *node) Address() string {
	return n.ListenAddress
}

func (n *node) Handshake(
	ctx context.Context,
	v *proto.Version,
) (*proto.Version, error) {
	c, err := client.NewGRPCClient(v.ListenAddress)
	if err != nil {
		return nil, err
	}
	n.nm.addPeer(c, v)
	n.logger.Infof("node: %s, sending handshake to %s", n, v.ListenAddress)
	return n.nm.version(), nil
}

func (n *node) NewTransaction(
	ctx context.Context,
	t *proto.Transaction,
) (*proto.Transaction, error) {
	peer, ok := grpcPeer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("node: %s, failed to get peer from context", n)
	}
	n.logger.Infof("node: %s, received transaction from %s", n, peer.Addr.String())

	if n.dr.Mempool().Contains(t) {
		return nil, fmt.Errorf("node: %s, transaction already exists in mempool", n)
	}
	n.dr.Mempool().Add(t)
	n.logger.Infof("node: %s, transaction added to mempool", n)

	// check how to broadcast transaction when peer is not available
	go n.nm.broadcast(t)

	return t, nil
}

func (n *node) NewBlock(ctx context.Context, b *proto.Block) (*proto.Block, error) {
	peer, ok := grpcPeer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("node: %s, failed to get peer from context", n)
	}
	n.logger.Infof("node: %s, received block from %s", n, peer.Addr.String())

	if err := n.dr.Chain().AddBlock(b); err != nil {
		return nil, err
	}
	n.logger.Infof("node: %s, block with height %d added to blockchain", n, b.Header.Height)

	n.clearMempool(b)

	// check how to broadcast block when peer is not available
	go n.nm.broadcast(b)

	return b, nil
}

func (n *node) GetBlocks(ctx context.Context, v *proto.Version) (*proto.Blocks, error) {
	blocks := &proto.Blocks{}
	for i := 0; i < n.dr.Chain().Height(); i++ {
		block, err := n.dr.Chain().GetBlockByHeight(i)
		if err != nil {
			return nil, err
		}
		blocks.Blocks = append(blocks.Blocks, block)
	}
	return blocks, nil
}

func (n *node) addMempoolToBlock(block *proto.Block) {
	for _, tx := range n.dr.Mempool().list() {
		block.Transactions = append(block.Transactions, tx)
	}

	// probably we should clear mempool after the block is mined
	n.dr.Mempool().Clear()
}

func (n *node) clearMempool(b *proto.Block) {
	for _, tx := range b.Transactions {
		n.dr.Mempool().Remove(tx)
	}
}

func (n *node) mineBlock(newBlockCh chan<- *proto.Block, stopMineBlockCh <-chan struct{}) {
	n.logger.Infof("node: %s, starting mining block\n", n)

	lastBlock, err := n.dr.Chain().GetBlockByHeight(n.dr.Chain().Height())
	if err != nil {
		n.logger.Errorf("node: %s, failed to get last block: %v", n, err)
		return
	}

	nonce := uint64(0)
	block := &proto.Block{
		Header: &proto.Header{
			PrevBlockHash: []byte(secure.HashBlock(lastBlock)), // TODO: check if this is correct
			Timestamp:     time.Now().Unix(),
			Height:        lastBlock.Header.Height + 1,
		},
	}
	n.addMempoolToBlock(block)
mine:
	for {
		select {
		case <-stopMineBlockCh:
			n.logger.Infof("node: %s, stopping minerLoop\n", n)
			return
		default:
			if block.GetTransactions() == nil {
				n.logger.Infof("node: %s, no transactions in mempool, block will not be mined\n", n)
				break mine
			}
			n.logger.Infof("node: %s, mining block\n", n)
			block.Header.Nonce = nonce
			blockHash := secure.HashBlock(block)
			fmt.Printf("blockHash: %s\n", blockHash)
			if secure.VerifyBlockHash(block) {
				n.logger.Infof("node: %s, mined block: %s\n", n, blockHash)
				newBlockCh <- block
				return
			}
			nonce++
		}
	}
	newBlockCh <- nil
}

func (n *node) minerLoop() {
	for {
		n.logger.Infof("node: %s, starting minerLoop\n", n)
		time.Sleep(miningInterval)
		newBlockCh := make(chan *proto.Block)
		stopMineBlockCh := make(chan struct{})

		go n.mineBlock(newBlockCh, stopMineBlockCh)
		ticker := time.NewTicker(maxMiningDuration)

	mining:
		for {
			select {
			case <-ticker.C:
				n.logger.Infof("node: %s, stopping minerLoop\n", n)
				close(stopMineBlockCh)
				break mining
			case block := <-newBlockCh:
				if block == nil {
					n.logger.Infof("node: %s, block is nil, will not be added to blockchain\n", n)
					break mining
				}
				block.PublicKey = n.PrivateKey.PublicKey().Bytes()
				secure.SignBlock(block, n.PrivateKey)

				n.dr.Chain().AddBlock(block)
				n.logger.Infof("node: %s, broadcast block: %s\n", n, secure.HashBlock(block))
				n.nm.broadcast(block)
				break mining
			default:
				continue
			}
		}
	}
}

func (n *node) showNodeInfo(quitCh chan struct{}, netLogging bool, blockchainLogging bool) {
	if netLogging || blockchainLogging {
		n.logger.Infof("node: %s, starting showPeers\n", n)
	}
	for {
		select {
		case <-quitCh:
			if netLogging || blockchainLogging {
				n.logger.Infof("node: %s, stopping showPeers", n)
			}
			return
		default:
			if netLogging {
				n.logger.Infof("Node %s, peers: %v", n, n.nm.peersAddrs(context.TODO()))
				// n.logger.Infof("Node %s, knownAddrs: %v", n, n.knownAddrs.list())
				n.logger.Infof("Node %s, mempool: %v", n, n.dr.Mempool().list())
			}
			if blockchainLogging {
				n.logger.Infof("Node %s, blockchain height: %d", n, n.dr.Chain().Height())
				n.logger.Infof("Node %s, blocks in blockchain: %v", n, len(n.dr.Chain().blockStore.List()))
				n.logger.Infof("Node %s, transactions in blockchain: %v", n, len(n.dr.Chain().txStore.List()))
				n.logger.Infof("Node %s, utxos in blockchain: %v", n, len(n.dr.Chain().utxoStore.List()))
			}
			time.Sleep(3 * time.Second)
		}
	}
}

func (n *node) processBlocks(blocks *proto.Blocks) error {
	for _, block := range blocks.Blocks {
		if err := n.dr.Chain().AddBlock(block); err != nil {
			return err
		}
	}
	return nil
}

func (n *node) syncBlockchainLoop(quit chan struct{}) {
	for {
		time.Sleep(syncBlockchainInterval)
		select {
		case <-quit:
			n.logger.Infof("node: %s, stopping syncBlockchainLoop", n)
			return
		default:
			for c := range n.nm.peers.peersForPing() {
				blocks, err := c.GetBlocks(context.Background(), n.nm.version())
				if err != nil {
					n.logger.Errorf("node: %s, failed to get blocks from %s: %v", n, c, err)
					continue
				}
				go n.processBlocks(blocks)
			}
		}
	}
}

func makeLogger() *zap.SugaredLogger {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.TimeKey = "timestamp"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339Nano)
	logger, err := loggerConfig.Build()
	if err != nil {
		log.Fatal(err)
	}
	return logger.Sugar()
}
