/*
 * Copyright 2021 ICON Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package test

import (
	"sync"
	"testing"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/errors"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/network"
)

var nmMu sync.Mutex

type NetworkManager struct {
	module.NetworkManager
	t      *testing.T
	id     module.PeerID
	rCh    chan packetEntry
	stopCh chan struct{}

	// mutable data
	peers    []Peer
	handlers []*nmHandler
	roles    map[string]module.Role
}

func indexOf(pl []Peer, id module.PeerID) int {
	for i := range pl {
		if pl[i].ID().Equal(id) {
			return i
		}
	}
	return -1
}

func NewNetworkManager(t *testing.T, a module.Address) *NetworkManager {
	const chLen = 1024
	n := &NetworkManager{
		t:      t,
		roles:  make(map[string]module.Role),
		id:     network.NewPeerIDFromAddress(a),
		rCh:    make(chan packetEntry, chLen),
		stopCh: make(chan struct{}),
	}
	go n.handlePacketLoop()
	return n
}

func (n *NetworkManager) handlePacketLoop() {
forLoop:
	for {
		select {
		case <-n.stopCh:
			break forLoop
		case p := <-n.rCh:
			n.handlePacket(p.pk, p.cb)
		}
	}
}

func (n *NetworkManager) Close() {
	n.stopCh <- struct{}{}
}

func (n *NetworkManager) attach(p Peer) {
	al := common.Lock(&nmMu)
	defer al.Unlock()

	if indexOf(n.peers, p.ID()) < 0 {
		n.peers = append(n.peers, p)
		var reactors []module.Reactor
		for _, h := range n.handlers {
			reactors = append(reactors, h.reactor)
		}
		al.Unlock()

		for _, reactor := range reactors {
			reactor.OnJoin(p.ID())
		}
	}
}

func (n *NetworkManager) detach(p Peer) {
	al := common.Lock(&nmMu)
	defer al.Unlock()

	if i := indexOf(n.peers, p.ID()); i >= 0 {
		last := len(n.peers) - 1
		n.peers[i] = n.peers[last]
		n.peers[last] = nil
		n.peers = n.peers[:last]

		var reactors []module.Reactor
		for _, h := range n.handlers {
			reactors = append(reactors, h.reactor)
		}
		al.Unlock()

		for _, reactor := range reactors {
			reactor.OnLeave(p.ID())
		}
	}
}

func (n *NetworkManager) notifyPacket(pk *Packet, cb func(rebroadcast bool, err error)) {
	n.rCh <- packetEntry{pk, cb}
}

func (n *NetworkManager) handlePacket(pk *Packet, cb func(rebroadcast bool, err error)) {
	al := common.Lock(&nmMu)
	defer al.Unlock()

	for _, h := range n.handlers {
		if pk.MPI == h.mpi {
			reactor := h.reactor
			al.Unlock()

			rb, err := reactor.OnReceive(pk.PI, pk.Data, pk.Src)
			if cb != nil {
				cb(rb, err)
			}
			return
		}
	}
}

func (n *NetworkManager) GetPeers() []module.PeerID {
	al := common.Lock(&nmMu)
	defer al.Unlock()

	peerIDs := make([]module.PeerID, len(n.peers))
	for i, p := range n.peers {
		peerIDs[i] = p.ID()
	}
	return peerIDs
}

func (n *NetworkManager) RegisterReactor(name string, mpi module.ProtocolInfo, reactor module.Reactor, piList []module.ProtocolInfo, priority uint8, policy module.NotRegisteredProtocolPolicy) (module.ProtocolHandler, error) {
	al := common.Lock(&nmMu)
	defer al.Unlock()

	h := &nmHandler{
		n,
		mpi,
		name,
		reactor,
		piList,
		priority,
	}
	n.handlers = append(n.handlers, h)
	return h, nil
}

func (n *NetworkManager) RegisterReactorForStreams(name string, pi module.ProtocolInfo, reactor module.Reactor, piList []module.ProtocolInfo, priority uint8, policy module.NotRegisteredProtocolPolicy) (module.ProtocolHandler, error) {
	return n.RegisterReactor(name, pi, reactor, piList, priority, policy)
}

func (n *NetworkManager) UnregisterReactor(reactor module.Reactor) error {
	al := common.Lock(&nmMu)
	defer al.Unlock()

	for i, h := range n.handlers {
		if h.reactor == reactor {
			last := len(n.handlers) - 1
			n.handlers[i] = n.handlers[last]
			n.handlers[last] = nil
			n.handlers = n.handlers[:last]
			return nil
		}
	}
	return nil
}

func (n *NetworkManager) SetRole(version int64, role module.Role, peers ...module.PeerID) {
	al := common.Lock(&nmMu)
	defer al.Unlock()

	for k, v := range n.roles {
		if v == role {
			delete(n.roles, k)
		}
	}
	for _, id := range peers {
		n.roles[string(id.Bytes())] = role
	}
}

func (n *NetworkManager) NewPeerFor(mpi module.ProtocolInfo) (*SimplePeer, *SimplePeerHandler) {
	p := NewPeer(n.t).Connect(n)
	h := p.RegisterProto(mpi)
	return p, h
}

func (n *NetworkManager) NewPeerForWithAddress(mpi module.ProtocolInfo, w module.Wallet) (*SimplePeer, *SimplePeerHandler) {
	p := NewPeerWithAddress(n.t, w).Connect(n)
	h := p.RegisterProto(mpi)
	return p, h
}

func (n *NetworkManager) Connect(n2 *NetworkManager) {
	PeerConnect(n, n2)
}

func (n *NetworkManager) ID() module.PeerID {
	return n.id
}

type nmHandler struct {
	n        *NetworkManager
	mpi      module.ProtocolInfo
	name     string
	reactor  module.Reactor
	piList   []module.ProtocolInfo
	priority uint8
}

func (h *nmHandler) Broadcast(pi module.ProtocolInfo, b []byte, bt module.BroadcastType) error {
	al := common.Lock(&nmMu)
	pk := &Packet{
		SendTypeBroadcast,
		h.n.id,
		bt,
		h.mpi,
		pi,
		b,
	}
	peers := append([]Peer(nil), h.n.peers...)
	al.Unlock()

	for _, p := range peers {
		p.notifyPacket(pk, nil)
	}
	return nil
}

func (h *nmHandler) Multicast(pi module.ProtocolInfo, b []byte, role module.Role) error {
	al := common.Lock(&nmMu)
	defer al.Unlock()

	var peers []Peer
	pk := &Packet{
		SendTypeMulticast,
		h.n.id,
		role,
		h.mpi,
		pi,
		b,
	}
	for _, p := range h.n.peers {
		if h.n.roles[string(p.ID().Bytes())] == role {
			peers = append(peers, p)
		}
	}
	al.Unlock()
	for _, p := range peers {
		p.notifyPacket(pk, nil)
	}
	return nil
}

func (h *nmHandler) Unicast(pi module.ProtocolInfo, b []byte, id module.PeerID) error {
	al := common.Lock(&nmMu)
	defer al.Unlock()

	if idx := indexOf(h.n.peers, id); idx >= 0 {
		pk := &Packet{
			SendTypeUnicast,
			h.n.id,
			id,
			h.mpi,
			pi,
			b,
		}
		p := h.n.peers[idx]
		al.Unlock()

		p.notifyPacket(pk, nil)
		return nil
	}
	return errors.New("no peer")
}

func (h *nmHandler) GetPeers() []module.PeerID {
	return h.n.GetPeers()
}
