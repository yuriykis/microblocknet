package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yuriykis/microblocknet/node/crypto"
	"github.com/yuriykis/microblocknet/node/proto"
	"github.com/yuriykis/microblocknet/node/service/client"
	"github.com/yuriykis/microblocknet/node/store"
	"github.com/yuriykis/microblocknet/node/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	grpcPeer "google.golang.org/grpc/peer"
)

const (
	connectInterval    = 5 * time.Second
	pingInterval       = 6 * time.Second
	maxConnectAttempts = 100
	miningInterval     = 5 * time.Second
	maxMiningDuration  = 10 * time.Second
)

type Service interface {
	Start(bootstrapNodes []string, isMiner bool) error
	Stop() error
	GetBlockByHeight(height int) (*proto.Block, error)
	GetUTXOsByAddress(address []byte) ([]*proto.UTXO, error)
	BootstrapNetwork(addrs []string) error
	PeersAddrs() []string
}

type Node interface {
	Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error)
	NewTransaction(ctx context.Context, t *proto.Transaction) (*proto.Transaction, error)
	NewBlock(ctx context.Context, b *proto.Block) (*proto.Block, error)
	GetBlocks(ctx context.Context, v *proto.Version) (*proto.Blocks, error)
}

type quit struct {
	tryConnectQuitCh   chan struct{}
	pingQuitCh         chan struct{}
	showNodeInfoQuitCh chan struct{}
}

func (q *quit) shutdown() {
	close(q.tryConnectQuitCh)
	close(q.pingQuitCh)
	close(q.showNodeInfoQuitCh)
}

type ServerConfig struct {
	Version       string
	ListenAddress string
	ApiListenAddr string
	PrivateKey    *crypto.PrivateKey
}
type node struct {
	ServerConfig

	logger     *zap.SugaredLogger
	peers      *peersMap
	knownAddrs *knownAddrs

	mempool *Mempool
	chain   *Chain

	isMiner bool

	transportServer TransportServer
	apiServer       ApiServer

	quit
}

func New(listenAddress string, apiListenAddress string) Service {
	var (
		txStore    = store.NewMemoryTxStore()
		blockStore = store.NewMemoryBlockStore()
		utxoStore  = store.NewMemoryUTXOStore()
	)
	chain := NewChain(txStore, blockStore, utxoStore)

	return &node{
		ServerConfig: ServerConfig{
			Version:       "0.0.1",
			ListenAddress: listenAddress,
			ApiListenAddr: apiListenAddress,
			PrivateKey:    nil,
		},

		peers:      NewPeersMap(),
		logger:     makeLogger(),
		knownAddrs: newKnownAddrs(),
		mempool:    NewMempool(),
		chain:      chain,

		quit: quit{
			tryConnectQuitCh:   make(chan struct{}),
			pingQuitCh:         make(chan struct{}),
			showNodeInfoQuitCh: make(chan struct{}),
		},
	}
}

func (n *node) Start(bootstrapNodes []string, isMiner bool) error {

	// TODO: adjust logger itself to show specific info, eg. node info, blockchain info, etc.
	go n.tryConnect(n.tryConnectQuitCh, false)
	go n.ping(n.pingQuitCh, false)
	go n.showNodeInfo(n.showNodeInfoQuitCh, false, true)

	if len(bootstrapNodes) > 0 {
		go func() {
			if err := n.BootstrapNetwork(bootstrapNodes); err != nil {
				n.logger.Errorf("node: %s, failed to bootstrap network: %v", n, err)
			}
		}()
	}

	if isMiner {
		n.isMiner = isMiner
		n.PrivateKey = crypto.GeneratePrivateKey()
		go n.minerLoop()
	}

	n.transportServer = NewGRPCNodeServer(n, n.ListenAddress)
	go n.transportServer.Start()

	n.apiServer = NewApiServer(n, n.ApiListenAddr)
	return n.apiServer.Start()
}

func (n *node) Stop() error {
	n.shutdown()
	n.apiServer.Stop()
	return n.transportServer.Stop()
}

func (n *node) String() string {
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
	n.addPeer(c, v)
	n.logger.Infof("node: %s, sending handshake to %s", n, v.ListenAddress)
	return n.Version(), nil
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

	if n.mempool.Contains(t) {
		return nil, fmt.Errorf("node: %s, transaction already exists in mempool", n)
	}
	n.mempool.Add(t)
	n.logger.Infof("node: %s, transaction added to mempool", n)

	// check how to broadcast transaction when peer is not available
	go n.broadcast(t)

	return t, nil
}

func (n *node) NewBlock(ctx context.Context, b *proto.Block) (*proto.Block, error) {
	peer, ok := grpcPeer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("node: %s, failed to get peer from context", n)
	}
	n.logger.Infof("node: %s, received block from %s", n, peer.Addr.String())

	if err := n.chain.AddBlock(b); err != nil {
		return nil, err
	}
	n.logger.Infof("node: %s, block with height %d added to blockchain", n, b.Header.Height)

	n.clearMempool(b)

	// check how to broadcast block when peer is not available
	go n.broadcast(b)

	return b, nil
}

func (n *node) GetBlocks(ctx context.Context, v *proto.Version) (*proto.Blocks, error) {
	blocks := &proto.Blocks{}
	for i := 0; i < n.chain.Height(); i++ {
		block, err := n.chain.GetBlockByHeight(i)
		if err != nil {
			return nil, err
		}
		blocks.Blocks = append(blocks.Blocks, block)
	}
	return blocks, nil
}

func (n *node) GetBlockByHeight(height int) (*proto.Block, error) {
	return n.chain.GetBlockByHeight(height)
}

func (n *node) GetUTXOsByAddress(address []byte) ([]*proto.UTXO, error) {
	return n.chain.utxoStore.GetByAddress(address)
}

func (n *node) Version() *proto.Version {
	return &proto.Version{
		Version:       "0.0.1",
		ListenAddress: n.ListenAddress,
		Peers:         n.PeersAddrs(),
	}
}

func (n *node) BootstrapNetwork(addrs []string) error {
	for _, addr := range addrs {
		if !n.canConnectWith(addr) {
			continue
		}
		n.knownAddrs.append(addr, 0)
	}
	return nil
}

func (n *node) PeersAddrs() []string {
	return n.peers.Addresses()
}

func (n *node) addMempoolToBlock(block *proto.Block) {
	for _, tx := range n.mempool.list() {
		block.Transactions = append(block.Transactions, tx)
	}

	// probably we should clear mempool after the block is mined
	n.mempool.Clear()
}

func (n *node) clearMempool(b *proto.Block) {
	for _, tx := range b.Transactions {
		n.mempool.Remove(tx)
	}
}

func (n *node) mineBlock(newBlockCh chan<- *proto.Block, stopMineBlockCh <-chan struct{}) {
	n.logger.Infof("node: %s, starting mining block\n", n)

	lastBlock, err := n.chain.GetBlockByHeight(n.chain.Height())
	if err != nil {
		n.logger.Errorf("node: %s, failed to get last block: %v", n, err)
		return
	}

	nonce := uint64(0)
	block := &proto.Block{
		Header: &proto.Header{
			PrevBlockHash: []byte(types.HashBlock(lastBlock)), // TODO: check if this is correct
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
			blockHash := types.HashBlock(block)
			fmt.Printf("blockHash: %s\n", blockHash)
			if types.VerifyBlockHash(block) {
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
				types.SignBlock(block, n.PrivateKey)

				n.chain.AddBlock(block)
				n.logger.Infof("node: %s, broadcast block: %s\n", n, types.HashBlock(block))
				n.broadcast(block)
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
				n.logger.Infof("Node %s, peers: %v", n, n.PeersAddrs())
				n.logger.Infof("Node %s, knownAddrs: %v", n, n.knownAddrs.list())
				n.logger.Infof("Node %s, mempool: %v", n, n.mempool.list())
			}
			if blockchainLogging {
				n.logger.Infof("Node %s, blockchain height: %d", n, n.chain.Height())
				n.logger.Infof("Node %s, blocks in blockchain: %v", n, len(n.chain.blockStore.List()))
				n.logger.Infof("Node %s, transactions in blockchain: %v", n, len(n.chain.txStore.List()))
				n.logger.Infof("Node %s, utxos in blockchain: %v", n, len(n.chain.utxoStore.List()))
			}
			time.Sleep(3 * time.Second)
		}
	}
}

func (n *node) processBlocks(blocks *proto.Blocks) error {
	for _, block := range blocks.Blocks {
		if err := n.chain.AddBlock(block); err != nil {
			return err
		}
	}
	return nil
}
func (n *node) syncBlockchain(c client.Client) error {
	blocks, err := c.GetBlocks(context.Background(), n.Version())
	if err != nil {
		return err
	}
	return n.processBlocks(blocks)
}

func (n *node) addPeer(c client.Client, v *proto.Version) {
	if !n.canConnectWith(v.ListenAddress) {
		return
	}
	n.peers.addPeer(c, v)

	if len(v.Peers) > 0 {
		go func() {
			if err := n.BootstrapNetwork(v.Peers); err != nil {
				n.logger.Errorf("node: %s, failed to bootstrap network: %v", n, err)
			}
		}()
	}
}

func (n *node) dialRemote(address string) (client.Client, *proto.Version, error) {
	client, err := client.NewGRPCClient(address)
	if err != nil {
		return nil, nil, err
	}
	version, err := n.handshakeClient(client)
	if err != nil {
		return nil, nil, err
	}
	return client, version, nil
}

func (n *node) handshakeClient(c client.Client) (*proto.Version, error) {
	version, err := c.Handshake(context.Background(), n.Version())
	if err != nil {
		return nil, err
	}
	n.logger.Infof("node: %s, handshake with %s, version: %v", n, c, version)
	return version, nil
}

func (n *node) broadcast(msg any) {
	for c := range n.Peers() {
		if err := n.sendMsg(c, msg); err != nil {
			n.logger.Errorf("node: %s, failed to send message to %s: %v", n, c, err)
		}
	}
}

func (n *node) sendMsg(c client.Client, msg any) error {
	switch m := msg.(type) {
	case *proto.Transaction:
		_, err := c.NewTransaction(context.Background(), m)
		if err != nil {
			return err
		}
	case *proto.Block:
		_, err := c.NewBlock(context.Background(), m)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("node: %s, unknown message type: %v", n, msg)
	}
	return nil
}

// TryConnect tries to connect to known addresses
func (n *node) tryConnect(quitCh chan struct{}, logging bool) {
	if logging {
		n.logger.Infof("node: %s, starting tryConnect\n", n)
	}
	for {
		select {
		case <-quitCh:
			if logging {
				n.logger.Infof("node: %s, stopping tryConnect\n", n)
			}
			return
		default:
			updatedKnownAddrs := make(map[string]int, 0)
			for addr, connectAttempts := range n.knownAddrs.list() {
				if !n.canConnectWith(addr) {
					continue
				}
				client, version, err := n.dialRemote(addr)
				if err != nil {
					errMsg := fmt.Sprintf(
						"node: %s, failed to connect to %s, will retry later: %v\n",
						n,
						addr,
						err)
					if connectAttempts >= maxConnectAttempts {
						errMsg = fmt.Sprintf(
							"node: %s, failed to connect to %s, reached maxConnectAttempts, removing from knownAddrs: %v\n",
							n,
							addr,
							err,
						)
					}
					if logging {
						n.logger.Errorf(errMsg)
					}
					// if the connection attemps is less than maxConnectAttempts, we will try to connect again
					// otherwise we will remove the address from the known addresses list
					// by not adding it to the updatedKnownAddrs map
					if connectAttempts < maxConnectAttempts {
						updatedKnownAddrs[addr] = connectAttempts + 1
					}
					continue
				}
				n.addPeer(client, version)
				n.syncBlockchain(client)
			}
			n.knownAddrs.update(updatedKnownAddrs)
			time.Sleep(connectInterval)
		}
	}
}

// Ping pings all known peers, if peer is not available,
// it will be removed from the peers list and added to the known addresses list
func (n *node) ping(quitCh chan struct{}, logging bool) {
	if logging {
		n.logger.Infof("node: %s, starting ping\n", n)
	}
	for {
		select {
		case <-quitCh:
			if logging {
				n.logger.Infof("node: %s, stopping ping\n", n)
			}
			return
		default:
			for c, p := range n.peers.peersForPing() {
				_, err := n.handshakeClient(c)
				if err != nil {
					if logging {
						n.logger.Errorf("node: %s, failed to ping %s: %v\n", n, c, err)
					}
					n.knownAddrs.append(p.ListenAddress, 0)
					n.peers.removePeer(c)
					continue
				}
				n.peers.updateLastPingTime(c)
				n.syncBlockchain(c)
			}
			time.Sleep(pingInterval)
		}
	}
}

func (n *node) canConnectWith(addr string) bool {
	if addr == n.ListenAddress {
		return false
	}
	for _, peer := range n.PeersAddrs() {
		if peer == addr {
			return false
		}
	}
	return true
}

func (n *node) Peers() map[client.Client]*peer {
	return n.peers.list()
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
