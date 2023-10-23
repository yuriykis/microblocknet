package node

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yuriykis/microblocknet/node/crypto"
	"github.com/yuriykis/microblocknet/node/node/client"
	"github.com/yuriykis/microblocknet/node/proto"
	"github.com/yuriykis/microblocknet/node/store"
	"github.com/yuriykis/microblocknet/node/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	gprcPeer "google.golang.org/grpc/peer"
)

const (
	connectInterval    = 5 * time.Second
	pingInterval       = 20 * time.Second
	maxConnectAttempts = 100
	miningInterval     = 5 * time.Second
	maxMiningDuration  = 10 * time.Second
)

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
	PrivateKey    *crypto.PrivateKey
}
type NetNode struct {
	ServerConfig

	logger     *zap.SugaredLogger
	peers      *peersMap
	knownAddrs *knownAddrs

	mempool *Mempool
	chain   *Chain

	isMiner bool

	transportServer TransportServer
	quit
}

func New(listenAddress string) *NetNode {
	var (
		txStore    = store.NewMemoryTxStore()
		blockStore = store.NewMemoryBlockStore()
		utxoStore  = store.NewMemoryUTXOStore()
	)
	chain := NewChain(txStore, blockStore, utxoStore)

	return &NetNode{
		ServerConfig: ServerConfig{
			Version:       "0.0.1",
			ListenAddress: listenAddress,
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

func (n *NetNode) Start(bootstrapNodes []string, isMiner bool) error {

	// TODO: adjust logger itself to show specific info, eg. node info, blockchain info, etc.
	go n.tryConnect(n.tryConnectQuitCh, true)
	go n.ping(n.pingQuitCh, true)
	go n.showNodeInfo(n.showNodeInfoQuitCh, false, false)

	if len(bootstrapNodes) > 0 {
		go func() {
			if err := n.BootstrapNetwork(bootstrapNodes); err != nil {
				n.logger.Errorf("NetNode: %s, failed to bootstrap network: %v", n, err)
			}
		}()
	}

	if isMiner {
		n.isMiner = isMiner
		n.PrivateKey = crypto.GeneratePrivateKey()
		go n.minerLoop()
	}

	n.transportServer = NewGRPCNodeServer(n, n.ListenAddress)
	return n.transportServer.Start()
}

func (n *NetNode) addMempoolToBlock(block *proto.Block) {
	for _, tx := range n.mempool.list() {
		block.Transactions = append(block.Transactions, tx)
	}

	// probably we should clear mempool after the block is mined
	n.mempool.Clear()
}

func (n *NetNode) mineBlock(newBlockCh chan<- *proto.Block, stopMineBlockCh <-chan struct{}) {
	n.logger.Infof("NetNode: %s, starting mining block\n", n)

	lastBlock, err := n.chain.GetBlockByHeight(n.chain.Height())
	if err != nil {
		n.logger.Errorf("NetNode: %s, failed to get last block: %v", n, err)
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
			n.logger.Infof("NetNode: %s, stopping minerLoop\n", n)
			return
		default:
			if block.GetTransactions() == nil {
				n.logger.Infof("NetNode: %s, no transactions in mempool, block will not be mined\n", n)
				break mine
			}
			n.logger.Infof("NetNode: %s, mining block\n", n)
			block.Header.Nonce = nonce
			blockHash := types.HashBlock(block)
			fmt.Printf("blockHash: %s\n", blockHash)
			if types.VerifyBlockHash(block) {
				n.logger.Infof("NetNode: %s, mined block: %s\n", n, blockHash)
				newBlockCh <- block
				return
			}
			nonce++
		}
	}
	newBlockCh <- nil
}

func (n *NetNode) minerLoop() {
	for {
		n.logger.Infof("NetNode: %s, starting minerLoop\n", n)
		time.Sleep(miningInterval)
		newBlockCh := make(chan *proto.Block)
		stopMineBlockCh := make(chan struct{})

		go n.mineBlock(newBlockCh, stopMineBlockCh)
		ticker := time.NewTicker(maxMiningDuration)

	mining:
		for {
			select {
			case <-ticker.C:
				n.logger.Infof("NetNode: %s, stopping minerLoop\n", n)
				close(stopMineBlockCh)
				break mining
			case block := <-newBlockCh:
				if block == nil {
					n.logger.Infof("NetNode: %s, block is nil, will not be added to blockchain\n", n)
					break mining
				}
				block.PublicKey = n.PrivateKey.PublicKey().Bytes()
				types.SignBlock(block, n.PrivateKey)

				n.chain.AddBlock(block)
				n.logger.Infof("NetNode: %s, broadcast block: %s\n", n, types.HashBlock(block))
				n.broadcast(block)
				break mining
			default:
				continue
			}
		}
	}
}

func (n *NetNode) showNodeInfo(quitCh chan struct{}, netLogging bool, blockchainLogging bool) {
	if netLogging || blockchainLogging {
		n.logger.Infof("NetNode: %s, starting showPeers\n", n)
	}
	for {
		select {
		case <-quitCh:
			if netLogging || blockchainLogging {
				n.logger.Infof("NetNode: %s, stopping showPeers", n)
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

func (n *NetNode) Stop() {
	n.shutdown()
	n.transportServer.Stop()
}

func (n *NetNode) String() string {
	return n.ListenAddress
}

func (n *NetNode) Handshake(
	ctx context.Context,
	v *proto.Version,
) (*proto.Version, error) {
	c, err := client.NewGRPCClient(v.ListenAddress)
	if err != nil {
		return nil, err
	}
	n.addPeer(c, v)
	n.logger.Infof("NetNode: %s, sending handshake to %s", n, v.ListenAddress)
	return n.Version(), nil
}

func (n *NetNode) NewTransaction(
	ctx context.Context,
	t *proto.Transaction,
) (*proto.Transaction, error) {
	peer, ok := gprcPeer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("NetNode: %s, failed to get peer from context", n)
	}
	n.logger.Infof("NetNode: %s, received transaction from %s", n, peer.Addr.String())

	if n.mempool.Contains(t) {
		return nil, fmt.Errorf("NetNode: %s, transaction already exists in mempool", n)
	}
	n.mempool.Add(t)
	n.logger.Infof("NetNode: %s, transaction added to mempool", n)

	// check how to broadcast transaction when peer is not available
	go n.broadcast(t)

	return t, nil
}

func (n *NetNode) NewBlock(ctx context.Context, b *proto.Block) (*proto.Block, error) {
	peer, ok := gprcPeer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("NetNode: %s, failed to get peer from context", n)
	}
	n.logger.Infof("NetNode: %s, received block from %s", n, peer.Addr.String())

	if err := n.chain.AddBlock(b); err != nil {
		return nil, err
	}
	n.logger.Infof("NetNode: %s, block with height %d added to blockchain", n, b.Header.Height)

	n.ClearMempool(b)

	// check how to broadcast block when peer is not available
	go n.broadcast(b)

	return b, nil
}

func (n *NetNode) ClearMempool(b *proto.Block) {
	for _, tx := range b.Transactions {
		n.mempool.Remove(tx)
	}
}

func (n *NetNode) GetBlocks(ctx context.Context, v *proto.Version) (*proto.Blocks, error) {
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

func (n *NetNode) GetBlockByHeight(height int) (*proto.Block, error) {
	return n.chain.GetBlockByHeight(height)
}

func (n *NetNode) GetUTXOsByAddress(address []byte) ([]*proto.UTXO, error) {
	return n.chain.utxoStore.GetByAddress(address)
}

func (n *NetNode) Version() *proto.Version {
	return &proto.Version{
		Version:       "0.0.1",
		ListenAddress: n.ListenAddress,
		Peers:         n.PeersAddrs(),
	}
}

func (n *NetNode) processBlocks(blocks *proto.Blocks) error {
	for _, block := range blocks.Blocks {
		if err := n.chain.AddBlock(block); err != nil {
			return err
		}
	}
	return nil
}
func (n *NetNode) syncBlockchain(c client.Client) error {
	blocks, err := c.GetBlocks(context.Background(), n.Version())
	if err != nil {
		return err
	}
	return n.processBlocks(blocks)
}

func (n *NetNode) addPeer(c client.Client, v *proto.Version) {
	if !n.canConnectWith(v.ListenAddress) {
		return
	}
	n.peers.addPeer(c, v)

	if len(v.Peers) > 0 {
		go func() {
			if err := n.BootstrapNetwork(v.Peers); err != nil {
				n.logger.Errorf("NetNode: %s, failed to bootstrap network: %v", n, err)
			}
		}()
	}
}

func (n *NetNode) dialRemote(address string) (client.Client, *proto.Version, error) {
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

func (n *NetNode) handshakeClient(c client.Client) (*proto.Version, error) {
	version, err := c.Handshake(context.Background(), n.Version())
	if err != nil {
		return nil, err
	}
	n.logger.Infof("NetNode: %s, handshake with %s, version: %v", n, c, version)
	return version, nil
}

func (n *NetNode) broadcast(msg any) {
	for c := range n.Peers() {
		if err := n.sendMsg(c, msg); err != nil {
			n.logger.Errorf("NetNode: %s, failed to send message to %s: %v", n, c, err)
		}
	}
}

func (n *NetNode) sendMsg(c client.Client, msg any) error {
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
		return fmt.Errorf("NetNode: %s, unknown message type: %v", n, msg)
	}
	return nil
}

func (n *NetNode) BootstrapNetwork(addrs []string) error {
	for _, addr := range addrs {
		if !n.canConnectWith(addr) {
			continue
		}
		n.knownAddrs.append(addr, 0)
	}
	return nil
}

// TryConnect tries to connect to known addresses
func (n *NetNode) tryConnect(quitCh chan struct{}, logging bool) {
	if logging {
		n.logger.Infof("NetNode: %s, starting tryConnect\n", n)
	}
	for {
		select {
		case <-quitCh:
			if logging {
				n.logger.Infof("NetNode: %s, stopping tryConnect\n", n)
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
						"NetNode: %s, failed to connect to %s, will retry later: %v\n",
						n,
						addr,
						err)
					if connectAttempts >= maxConnectAttempts {
						errMsg = fmt.Sprintf(
							"NetNode: %s, failed to connect to %s, reached maxConnectAttempts, removing from knownAddrs: %v\n",
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
func (n *NetNode) ping(quitCh chan struct{}, logging bool) {
	if logging {
		n.logger.Infof("NetNode: %s, starting ping\n", n)
	}
	for {
		select {
		case <-quitCh:
			if logging {
				n.logger.Infof("NetNode: %s, stopping ping\n", n)
			}
			return
		default:
			for c, p := range n.peers.peersForPing() {
				_, err := n.handshakeClient(c)
				if err != nil {
					if logging {
						n.logger.Errorf("NetNode: %s, failed to ping %s: %v\n", n, c, err)
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

func (n *NetNode) canConnectWith(addr string) bool {
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

func (n *NetNode) PeersAddrs() []string {
	return n.peers.Addresses()
}

func (n *NetNode) Peers() map[client.Client]*peer {
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
