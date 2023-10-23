package service

import (
	"sync"
	"time"

	"github.com/yuriykis/microblocknet/node/proto"
	"github.com/yuriykis/microblocknet/node/service/client"
)

const checkConnectInterval = 50 * time.Second

type peer struct {
	*proto.Version
	lastPing time.Time
}

func newPeer(v *proto.Version) *peer {
	return &peer{
		Version:  v,
		lastPing: time.Now(),
	}
}

// ---------------------------------------------------------------------------

type peersMap struct {
	peers map[client.Client]*peer
	lock  sync.RWMutex
}

func NewPeersMap() *peersMap {
	return &peersMap{
		peers: make(map[client.Client]*peer),
	}
}
func (pm *peersMap) addPeer(c client.Client, v *proto.Version) {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	pm.peers[c] = newPeer(v)
}

func (pm *peersMap) removePeer(c client.Client) {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	delete(pm.peers, c)
}

func (pm *peersMap) Addresses() []string {
	pm.lock.RLock()
	defer pm.lock.RUnlock()
	peersList := make([]string, 0)
	for _, v := range pm.peers {
		peersList = append(peersList, v.ListenAddress)
	}
	return peersList
}

func (pm *peersMap) list() map[client.Client]*peer {
	pm.lock.RLock()
	defer pm.lock.RUnlock()
	return pm.peers
}

func (pm *peersMap) peersForPing() map[client.Client]*peer {
	pm.lock.RLock()
	defer pm.lock.RUnlock()
	peers := make(map[client.Client]*peer)
	for c, p := range pm.peers {
		if time.Since(p.lastPing) > checkConnectInterval {
			peers[c] = p
		}
	}
	return peers
}

func (pm *peersMap) updateLastPingTime(c client.Client) {
	pm.lock.Lock()
	defer pm.lock.Unlock()
	pm.peers[c].lastPing = time.Now()
}

// ---------------------------------------------------------------------------
type knownAddrs struct {
	addrs map[string]int // [addr]connectAttempts
	lock  sync.RWMutex
}

func newKnownAddrs() *knownAddrs {
	return &knownAddrs{
		addrs: make(map[string]int),
	}
}

func (ka *knownAddrs) append(addr string, connectAttempts int) {
	ka.lock.Lock()
	defer ka.lock.Unlock()
	ka.addrs[addr] = connectAttempts
}

func (ka *knownAddrs) update(addrs map[string]int) {
	ka.lock.Lock()
	defer ka.lock.Unlock()
	ka.addrs = addrs
}

func (ka *knownAddrs) list() map[string]int {
	ka.lock.Lock()
	defer ka.lock.Unlock()
	return ka.addrs
}
