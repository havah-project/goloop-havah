package network

import (
	"fmt"
	"sync"

	"github.com/icon-project/goloop/common/errors"
	"github.com/icon-project/goloop/common/log"
	"github.com/icon-project/goloop/module"
)

var (
	p2pProtoChanJoinReq  = module.ProtocolInfo(0x0500)
	p2pProtoChanJoinResp = module.ProtocolInfo(0x0600)
)

type ChannelNegotiator struct {
	*peerHandler
	netAddress NetAddress
	m          map[string]*ProtocolInfos
	mtx        sync.RWMutex
}

func newChannelNegotiator(netAddress NetAddress, l log.Logger) *ChannelNegotiator {
	cn := &ChannelNegotiator{
		netAddress:  netAddress,
		peerHandler: newPeerHandler(l.WithFields(log.Fields{LoggerFieldKeySubModule: "negotiator"})),
		m:           make(map[string]*ProtocolInfos),
	}
	return cn
}

func (cn *ChannelNegotiator) onPeer(p *Peer) {
	cn.logger.Traceln("onPeer", p)
	if !p.In() {
		cn.setWaitInfo(p2pProtoChanJoinResp, p)
		cn.sendJoinRequest(p)
	} else {
		cn.setWaitInfo(p2pProtoChanJoinReq, p)
	}
}

func (cn *ChannelNegotiator) onError(err error, p *Peer, pkt *Packet) {
	cn.logger.Infoln("onError", err, p, pkt)
	cn.peerHandler.onError(err, p, pkt)
}

func (cn *ChannelNegotiator) onPacket(pkt *Packet, p *Peer) {
	switch pkt.protocol {
	case p2pProtoControl:
		switch pkt.subProtocol {
		case p2pProtoChanJoinReq:
			cn.handleJoinRequest(pkt, p)
		case p2pProtoChanJoinResp:
			cn.handleJoinResponse(pkt, p)
		default:
			p.CloseByError(ErrNotRegisteredProtocol)
		}
	}
}

type JoinRequest struct {
	Channel   string
	Addr      NetAddress
	Protocols []module.ProtocolInfo
}

type JoinResponse struct {
	Channel   string
	Addr      NetAddress
	Protocols []module.ProtocolInfo
}

var defaultProtocols = []module.ProtocolInfo{
	module.ProtoStateSync,
	module.ProtoTransaction,
	module.ProtoConsensus,
	module.ProtoFastSync,
	module.ProtoConsensusSync,
}

func (cn *ChannelNegotiator) addProtocol(channel string, pi module.ProtocolInfo) {
	cn.mtx.Lock()
	defer cn.mtx.Unlock()

	pis, ok := cn.m[channel]
	if !ok {
		pis = newProtocolInfos()
		cn.m[channel] = pis
	}
	pis.Add(pi)
}

func (cn *ChannelNegotiator) removeProtocol(channel string, pi module.ProtocolInfo) {
	cn.mtx.Lock()
	defer cn.mtx.Unlock()

	if pis, ok := cn.m[channel]; ok {
		pis.Remove(pi)
		if pis.Len() == 0 {
			delete(cn.m, channel)
		}
	}
}

func (cn *ChannelNegotiator) ProtocolInfos(channel string) *ProtocolInfos {
	cn.mtx.RLock()
	defer cn.mtx.RUnlock()

	return cn.m[channel]
}

func (cn *ChannelNegotiator) resolveProtocols(p *Peer, channel string, protocols []module.ProtocolInfo) error {
	if p.Channel() != channel {
		return errors.Errorf("invalid channel")
	}

	pis := cn.ProtocolInfos(channel)
	if pis == nil {
		return errors.Errorf("not exists channel")
	}

	ppis := p.ProtocolInfos()
	if len(protocols) == 0 {
		protocols = defaultProtocols
	}
	ppis.Set(protocols)
	ppis.Resolve(pis)
	if ppis.Len() == 0 {
		return errors.Errorf("not supported protocols")
	}
	return nil
}

func (cn *ChannelNegotiator) sendJoinRequest(p *Peer) {
	pis := cn.ProtocolInfos(p.Channel())
	if pis == nil {
		err := fmt.Errorf("sendJoinRequest error[%v]", "not exists channel")
		cn.logger.Infoln("sendJoinRequest", p.ConnString(), "ChannelNegotiatorError", err)
		p.CloseByError(err)
		return
	}
	m := &JoinRequest{Channel: p.Channel(), Addr: cn.netAddress, Protocols: pis.Array()}
	cn.sendMessage(p2pProtoChanJoinReq, m, p)
	cn.logger.Traceln("sendJoinRequest", m, p)
}

func (cn *ChannelNegotiator) handleJoinRequest(pkt *Packet, p *Peer) {
	if !cn.checkWaitInfo(pkt, p) {
		return
	}

	rm := &JoinRequest{}
	if !cn.decodePeerPacket(p, rm, pkt) {
		return
	}
	cn.logger.Traceln("handleJoinRequest", rm, p)

	if err := cn.resolveProtocols(p, rm.Channel, rm.Protocols); err != nil {
		err = fmt.Errorf("handleJoinRequest error[%v]", err.Error())
		cn.logger.Infoln("handleJoinRequest", p.ConnString(), "ChannelNegotiatorError", err)
		p.CloseByError(err)
		return
	}
	p.setNetAddress(rm.Addr)

	m := &JoinResponse{Channel: p.Channel(), Addr: cn.netAddress, Protocols: p.ProtocolInfos().Array()}
	cn.sendMessage(p2pProtoChanJoinResp, m, p)

	cn.nextOnPeer(p)
}

func (cn *ChannelNegotiator) handleJoinResponse(pkt *Packet, p *Peer) {
	if !cn.checkWaitInfo(pkt, p) {
		return
	}

	rm := &JoinResponse{}
	if !cn.decodePeerPacket(p, rm, pkt) {
		return
	}
	cn.logger.Traceln("handleJoinResponse", rm, p)

	if err := cn.resolveProtocols(p, rm.Channel, rm.Protocols); err != nil {
		err = fmt.Errorf("handleJoinResponse error[%v]", err.Error())
		cn.logger.Infoln("handleJoinResponse", p.ConnString(), "ChannelNegotiatorError", err)
		p.CloseByError(err)
		return
	}
	p.setNetAddress(rm.Addr)

	cn.nextOnPeer(p)
}
