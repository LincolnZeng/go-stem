/**
*  @file
*  @copyright defined in scdo/LICENSE
 */

package scdo

import (
	"net"
	"testing"

	"github.com/scdoproject/go-scdo/crypto"
	log2 "github.com/scdoproject/go-scdo/log"
	"github.com/scdoproject/go-scdo/p2p"
	"github.com/scdoproject/go-scdo/p2p/discovery"
	"github.com/stretchr/testify/assert"
)

func getTestPeer(shard uint) *peer {
	log := log2.GetLogger("test")
	addr := crypto.MustGenerateRandomAddress()
	node := discovery.NewNodeWithAddr(*addr, &net.UDPAddr{}, shard)
	p2pPeer := p2p.NewPeer(nil, nil, node)
	peer := newPeer(1, p2pPeer, nil, log)

	return peer
}

func Test_PeerSet_Add(t *testing.T) {
	set := newPeerSet()

	peer1 := getTestPeer(0)
	set.Add(peer1)
	assert.Equal(t, len(set.peerMap), 1)
	assert.Equal(t, len(set.shardPeers[0]), 1)

	set.Add(peer1)
	assert.Equal(t, len(set.peerMap), 1)
	assert.Equal(t, len(set.shardPeers[0]), 1)

	peer2 := getTestPeer(1)
	set.Add(peer2)
	assert.Equal(t, len(set.peerMap), 2)
	assert.Equal(t, len(set.shardPeers[1]), 1)
}

func Test_PeerSet_Find(t *testing.T) {
	set := newPeerSet()
	peer1 := getTestPeer(0)
	set.Add(peer1)
	peer2 := getTestPeer(0)
	set.Add(peer2)

	assert.Equal(t, set.Find(peer1.Node.ID), peer1)
	assert.Equal(t, set.Find(peer2.Node.ID), peer2)
}

func Test_PeerSet_Remove(t *testing.T) {
	set := newPeerSet()
	peer1 := getTestPeer(0)
	set.Add(peer1)
	peer2 := getTestPeer(1)
	set.Add(peer2)

	assert.Equal(t, len(set.peerMap), 2)
	set.Remove(peer1.Node.ID)
	assert.Equal(t, len(set.peerMap), 1)
	assert.Equal(t, len(set.shardPeers[0]), 0)
	assert.Equal(t, len(set.shardPeers[1]), 1)
	set.Remove(peer1.Node.ID)
	assert.Equal(t, len(set.peerMap), 1)
	set.Remove(peer2.Node.ID)
	assert.Equal(t, len(set.peerMap), 0)
	assert.Equal(t, len(set.shardPeers[0]), 0)
	assert.Equal(t, len(set.shardPeers[1]), 0)
}
