package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yuriykis/microblocknet/common/crypto"
	"github.com/yuriykis/microblocknet/common/proto"
	"github.com/yuriykis/microblocknet/node/chain"
	"github.com/yuriykis/microblocknet/node/client"
	"github.com/yuriykis/microblocknet/node/secure"
	"github.com/yuriykis/microblocknet/node/store"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	grpcPeer "google.golang.org/grpc/peer"
)

const (
	miningInterval         = 5 * time.Second
	maxMiningDuration      = 10 * time.Second
	syncBlockchainInterval = 5 * time.Second
)

type Noder interface {
	Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error)
	NewTransaction(ctx context.Context, t *proto.Transaction) (*proto.Transaction, error)
	NewBlock(ctx context.Context, b *proto.Block) (*proto.Block, error)
	GetBlocks(ctx context.Context, v *proto.Version) (*proto.Blocks, error)
}

type Api interface {
	Chain() *chain.Chain
	Gate() *gatewayClient
}

type NodeOpts struct {
	BootstrapNodes []string
	IsMiner        bool
}

type ServerConfig struct {
	Version              string
	NodeListenAddress    string
	ApiListenAddr        string
	GatewayAddress       string
	ConsulServiceAddress string
	StoreType            string
}

type Node struct {
	ServerConfig

	isMiner    bool
	PrivateKey *crypto.PrivateKey

	logger *zap.SugaredLogger

	nm *networkManager

	chain   *chain.Chain
	mempool *Mempool

	gate          *gatewayClient
	consulService *ConsulService

	quitNode
}

type quitNode struct {
	showNodeInfoQuitCh   chan struct{}
	syncBlockchainQuitCh chan struct{}
	pingQuitCh           chan struct{}
}

func (n *Node) shutdown() {
	close(n.showNodeInfoQuitCh)
	close(n.syncBlockchainQuitCh)
}

func New(conf ServerConfig) *Node {
	logger := makeLogger()
	st, err := store.NewChainStore(conf.StoreType)
	if err != nil {
		log.Fatal(err)
	}
	return &Node{
		ServerConfig: conf,

		logger: logger,

		nm: NewNetworkManager(conf.NodeListenAddress, logger),

		chain:   chain.New(st),
		mempool: NewMempool(),

		gate:          NewGatewayClient(conf.GatewayAddress, logger),
		consulService: NewConsulService(logger, conf.ConsulServiceAddress),

		quitNode: quitNode{
			showNodeInfoQuitCh:   make(chan struct{}),
			syncBlockchainQuitCh: make(chan struct{}),
		},
	}
}

func (n *Node) Start(opts NodeOpts) error {

	n.nm.start(opts.BootstrapNodes)
	n.consulService.Start()

	go n.syncBlockchainLoop(n.syncBlockchainQuitCh)
	go n.showNodeInfo(n.showNodeInfoQuitCh, false, true)

	if opts.IsMiner {
		n.isMiner = opts.IsMiner
		n.PrivateKey = crypto.GeneratePrivateKey()
		go n.minerLoop()
	}

	go n.Gate().registerGatewayLoop(n.pingQuitCh, n.ApiListenAddr)

	return nil
}

func (n *Node) Stop() error {
	n.shutdown()
	n.nm.stop()
	return nil
}

func (n *Node) Chain() *chain.Chain {
	return n.chain
}

func (n *Node) Gate() *gatewayClient {
	return n.gate
}

func (r *Node) Mempool() *Mempool {
	return r.mempool
}

func (n *Node) String() string {
	return n.NodeListenAddress
}

func (n *Node) Address() string {
	return n.NodeListenAddress
}

func (n *Node) Handshake(
	ctx context.Context,
	v *proto.Version,
) (*proto.Version, error) {
	c, err := client.NewGRPCClient(v.ListenAddress)
	if err != nil {
		return nil, err
	}
	n.nm.addPeer(c, v)
	n.logger.Infof("Node: %s, sending handshake to %s", n, v.ListenAddress)
	return n.nm.version(), nil
}

func (n *Node) NewTransaction(
	ctx context.Context,
	t *proto.Transaction,
) (*proto.Transaction, error) {
	peer, ok := grpcPeer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("Node: %s, failed to get peer from context", n)
	}
	n.logger.Infof("Node: %s, received transaction from %s", n, peer.Addr.String())

	if n.Mempool().Contains(t) {
		return nil, fmt.Errorf("Node: %s, transaction already exists in mempool", n)
	}
	n.Mempool().Add(t)
	n.logger.Infof("Node: %s, transaction added to mempool", n)

	// check how to broadcast transaction when peer is not available
	go n.nm.broadcast(t)

	return t, nil
}

func (n *Node) NewBlock(ctx context.Context, b *proto.Block) (*proto.Block, error) {
	peer, ok := grpcPeer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("Node: %s, failed to get peer from context", n)
	}
	n.logger.Infof("Node: %s, received block from %s", n, peer.Addr.String())

	if err := n.Chain().AddBlock(b); err != nil {
		return nil, err
	}
	n.logger.Infof("Node: %s, block with height %d added to blockchain", n, b.Header.Height)

	n.clearMempool(b)

	// check how to broadcast block when peer is not available
	go n.nm.broadcast(b)

	return b, nil
}

func (n *Node) GetBlocks(ctx context.Context, v *proto.Version) (*proto.Blocks, error) {
	blocks := &proto.Blocks{}
	for i := 0; i < n.Chain().Height(); i++ {
		block, err := n.Chain().GetBlockByHeight(i)
		if err != nil {
			return nil, err
		}
		blocks.Blocks = append(blocks.Blocks, block)
	}
	return blocks, nil
}

func (n *Node) addMempoolToBlock(block *proto.Block) {
	for _, tx := range n.Mempool().List() {
		block.Transactions = append(block.Transactions, tx)
	}

	// probably we should clear mempool after the block is mined
	n.Mempool().Clear()
}

func (n *Node) clearMempool(b *proto.Block) {
	for _, tx := range b.Transactions {
		n.Mempool().Remove(tx)
	}
}

func (n *Node) mineBlock(newBlockCh chan<- *proto.Block, stopMineBlockCh <-chan struct{}) {
	n.logger.Infof("Node: %s, starting mining block\n", n)

	lastBlock, err := n.Chain().GetBlockByHeight(n.Chain().Height())
	if err != nil {
		n.logger.Errorf("Node: %s, failed to get last block: %v", n, err)
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
			n.logger.Infof("Node: %s, stopping minerLoop\n", n)
			return
		default:
			if block.GetTransactions() == nil {
				n.logger.Infof("Node: %s, no transactions in mempool, block will not be mined\n", n)
				break mine
			}
			n.logger.Infof("Node: %s, mining block\n", n)
			block.Header.Nonce = nonce
			blockHash := secure.HashBlock(block)
			fmt.Printf("blockHash: %s\n", blockHash)
			if secure.VerifyBlockHash(block) {
				n.logger.Infof("Node: %s, mined block: %s\n", n, blockHash)
				newBlockCh <- block
				return
			}
			nonce++
		}
	}
	newBlockCh <- nil
}

func (n *Node) minerLoop() {
	for {
		n.logger.Infof("Node: %s, starting minerLoop\n", n)
		time.Sleep(miningInterval)
		newBlockCh := make(chan *proto.Block)
		stopMineBlockCh := make(chan struct{})

		go n.mineBlock(newBlockCh, stopMineBlockCh)
		ticker := time.NewTicker(maxMiningDuration)

	mining:
		for {
			select {
			case <-ticker.C:
				n.logger.Infof("Node: %s, stopping minerLoop\n", n)
				close(stopMineBlockCh)
				break mining
			case block := <-newBlockCh:
				if block == nil {
					n.logger.Infof("Node: %s, block is nil, will not be added to blockchain\n", n)
					break mining
				}
				block.PublicKey = n.PrivateKey.PublicKey().Bytes()
				secure.SignBlock(block, n.PrivateKey)

				n.Chain().AddBlock(block)
				n.logger.Infof("Node: %s, broadcast block: %s\n", n, secure.HashBlock(block))
				n.nm.broadcast(block)
				break mining
			default:
				continue
			}
		}
	}
}

func (n *Node) showNodeInfo(quitCh chan struct{}, netLogging bool, blockchainLogging bool) {
	ctx := context.Background() // TODO: check if this is correct

	if netLogging || blockchainLogging {
		n.logger.Infof("Node: %s, starting showPeers\n", n)
	}
	for {
		select {
		case <-quitCh:
			if netLogging || blockchainLogging {
				n.logger.Infof("Node: %s, stopping showPeers", n)
			}
			return
		default:
			if netLogging {
				n.logger.Infof("Node %s, peers: %v", n, n.nm.peersAddrs(context.TODO()))
				// n.logger.Infof("Node %s, knownAddrs: %v", n, n.knownAddrs.list())
				n.logger.Infof("Node %s, mempool: %v", n, n.Mempool().List())
			}
			if blockchainLogging {
				n.logger.Infof("Node %s, blockchain height: %d", n, n.Chain().Height())
				n.logger.Infof("Node %s, blocks in blockchain: %v", n, len(n.Chain().Store().BlockStore(ctx).List(ctx)))
				n.logger.Infof("Node %s, transactions in blockchain: %v", n, len(n.Chain().Store().TxStore(ctx).List(ctx)))
				n.logger.Infof("Node %s, utxos in blockchain: %v", n, len(n.Chain().Store().UTXOStore(ctx).List(ctx)))
			}
			time.Sleep(3 * time.Second)
		}
	}
}

func (n *Node) processBlocks(blocks *proto.Blocks) error {
	for _, block := range blocks.Blocks {
		if err := n.Chain().AddBlock(block); err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) syncBlockchainLoop(quit chan struct{}) {
	for {
		time.Sleep(syncBlockchainInterval)
		select {
		case <-quit:
			n.logger.Infof("Node: %s, stopping syncBlockchainLoop", n)
			return
		default:
			for c := range n.nm.peers.peersForPing() {
				blocks, err := c.GetBlocks(context.Background(), n.nm.version())
				if err != nil {
					n.logger.Errorf("Node: %s, failed to get blocks from %s: %v", n, c, err)
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
