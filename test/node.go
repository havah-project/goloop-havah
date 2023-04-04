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
	"io"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/icon-project/goloop/chain/base"
	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/codec"
	"github.com/icon-project/goloop/common/log"
	"github.com/icon-project/goloop/common/wallet"
	"github.com/icon-project/goloop/consensus"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/contract"
	"github.com/icon-project/goloop/service/eeproxy"
)

type Node struct {
	*testing.T
	Chain     *Chain
	Base      string
	em        eeproxy.Manager
	NM        *NetworkManager
	SM        module.ServiceManager
	BM        module.BlockManager
	CS        module.Consensus
	PrevBlock module.Block
	LastBlock module.Block
	Platform  base.Platform
}

type NodeContext struct {
	C        *Chain
	Config   *FixtureConfig
	Base     string
	Platform base.Platform
	CM       contract.ContractManager
	EM       eeproxy.Manager
}

func NewNode(t *testing.T, o ...FixtureOption) *Node {
	cf := NewFixtureConfig(t, o...)
	base, err := os.MkdirTemp("", cf.Prefix)
	assert.NoError(t, err)
	dbase := cf.Dbase()
	logger := log.New()
	w := cf.Wallet
	if w == nil {
		w = wallet.New()
	}
	c, err := NewChain(t, w, dbase, logger, cf.CVSD, cf.Genesis)
	assert.NoError(t, err)
	c.Logger().SetLevel(log.TraceLevel)

	// set up sm
	RegisterTransactionFactory()
	const (
		ContractPath = "contract"
		EESocketPath = "ee.sock"
	)
	ctx := &NodeContext{
		C:      c,
		Config: cf,
		Base:   base,
	}
	plt := cf.NewPlatform(ctx)
	ctx.Platform = plt
	cm, err := plt.NewContractManager(c.Database(), path.Join(base, ContractPath), c.Logger())
	assert.NoError(t, err)
	ee, err := eeproxy.AllocEngines(c.Logger(), "python")
	assert.NoError(t, err)
	em, err := eeproxy.NewManager("unix", path.Join(base, EESocketPath), c.Logger(), ee...)
	assert.NoError(t, err)

	go func() {
		_ = em.Loop()
	}()
	err = em.SetInstances(0, 0, 0)
	assert.NoError(t, err)

	ctx.CM = cm
	ctx.EM = em
	c.sm = NewServiceManager(c, plt, cm, em)

	c.bm = cf.NewBM(ctx)
	lastBlk, err := c.bm.GetLastBlock()
	assert.NoError(t, err)

	c.cs = cf.NewCS(ctx)

	return &Node{
		T:         t,
		Chain:     c,
		Base:      base,
		em:        em,
		NM:        c.nm,
		SM:        c.sm,
		BM:        c.bm,
		CS:        c.cs,
		PrevBlock: nil,
		LastBlock: lastBlk,
		Platform:  plt,
	}
}

func (t *Node) Close() {
	err := t.em.Close()
	assert.NoError(t, err)
	err = os.RemoveAll(t.Base)
	assert.NoError(t, err)
	t.CS.Term()
	t.BM.Term()
	t.Chain.Close()
}

func (t *Node) GetLastBlock() module.Block {
	return GetLastBlock(t.T, t.BM)
}

func (t *Node) AssertLastBlock(
	prev module.Block, version int,
) {
	var height int64
	var prevID []byte
	if prev != nil {
		height = prev.Height() + 1
		prevID = prev.ID()
	}
	AssertLastBlock(t.T, t.BM, height, prevID, version)
}

func (t *Node) ProposeBlock(
	votes module.CommitVoteSet,
) module.BlockCandidate {
	blk, err, cbErr := ProposeBlock(t.BM, t.LastBlock.ID(), votes)
	assert.NoError(t, err)
	assert.NoError(t, cbErr)
	assert.Equal(t, t.LastBlock.ID(), blk.PrevID())
	assert.Equal(t, t.LastBlock.Height()+1, blk.Height())
	return blk
}

func (t *Node) ImportBlock(bc module.BlockCandidate, flag int) {
	assert.Equal(t, t.LastBlock.ID(), bc.PrevID())
	assert.Equal(t, t.LastBlock.Height()+1, bc.Height())
	blk, err, cbErr := ImportBlock(t.T, t.BM, bc, flag)
	assert.NoError(t, err)
	assert.NoError(t, cbErr)
	assert.Equal(t, bc.ID(), blk.ID())
}

func (t *Node) ImportBlockByReader(
	r io.Reader, flag int,
) module.BlockCandidate {
	bc, err, cbErr := ImportBlockByReader(t.T, t.BM, r, flag)
	assert.NoError(t, err)
	assert.NoError(t, cbErr)
	assert.Equal(t, t.LastBlock.ID(), bc.PrevID())
	assert.Equal(t, t.LastBlock.Height()+1, bc.Height())
	return bc
}

func (t *Node) UpdateLastBlock() {
	lastBlock, err := t.BM.GetLastBlock()
	assert.NoError(t, err)
	t.LastBlock = lastBlock
}

func (t *Node) FinalizeBlock(bc module.BlockCandidate) {
	prevBlock := t.LastBlock
	FinalizeBlock(t.T, t.BM, bc)
	blk, err := t.BM.GetLastBlock()
	assert.NoError(t, err)
	assert.Equal(t, bc.ID(), blk.ID())
	t.AssertLastBlock(prevBlock, bc.Version())
	t.PrevBlock = t.LastBlock
	t.LastBlock = blk
}

func (t *Node) ProposeFinalizeBlock(votes module.CommitVoteSet) {
	bc := t.ProposeBlock(votes)
	t.FinalizeBlock(bc)
	bc.Dispose()
}

func (t *Node) ProposeImportFinalizeBlock(
	votes module.CommitVoteSet,
) {
	bc := t.ProposeBlock(votes)
	t.ImportBlock(bc, 0)
	t.FinalizeBlock(bc)
	bc.Dispose()
}

func (t *Node) ImportFinalizeBlockByReader(r io.Reader) {
	bc := t.ImportBlockByReader(r, 0)
	t.FinalizeBlock(bc)
	bc.Dispose()
}

func (t *Node) ProposeFinalizeBlockWithTX(
	votes module.CommitVoteSet, txJson string,
) {
	tid, err := t.SM.SendTransaction(nil, 0, txJson)
	assert.NoError(t, err)
	bc := t.ProposeBlock(votes)
	t.FinalizeBlock(bc)
	tx, err := t.LastBlock.NormalTransactions().Get(0)
	assert.NoError(t, err)
	assert.Equal(t, tid, tx.ID())
	bc.Dispose()
}

func (t *Node) ProposeImportFinalizeBlockWithTX(
	votes module.CommitVoteSet, txJson string,
) {
	tid, err := t.SM.SendTransaction(nil, 0, txJson)
	assert.NoError(t, err)
	bc := t.ProposeBlock(votes)
	t.ImportBlock(bc, 0)
	t.FinalizeBlock(bc)
	tx, err := t.LastBlock.NormalTransactions().Get(0)
	assert.NoError(t, err)
	assert.Equal(t, tid, tx.ID())
	bc.Dispose()
}

func (t *Node) NewVoteListForLastBlock() module.CommitVoteSet {
	var pcm module.BTPProofContextMap
	var ntsHashEntries []module.NTSHashEntryFormat
	var ntsdProofParts [][]byte
	var ntsVoteCount int
	if t.LastBlock.Height() > 1 {
		blk, err := t.BM.GetBlockByHeight(t.LastBlock.Height() - 1)
		assert.NoError(t, err)
		pcm, err = blk.NextProofContextMap()
		assert.NoError(t, err)
		bd, err := t.LastBlock.BTPDigest()
		assert.NoError(t, err)
		ntsdProofParts = make([][]byte, 0)
		for _, ntd := range bd.NetworkTypeDigests() {
			if pc, err := pcm.ProofContextFor(ntd.NetworkTypeID()); err == nil {
				ntsd := pc.NewDecision(
					module.GetSourceNetworkUID(t.Chain),
					ntd.NetworkTypeID(),
					t.LastBlock.Height(),
					t.LastBlock.Votes().VoteRound(),
					ntd.NetworkTypeSectionHash(),
				)
				pp, err := pc.NewProofPart(ntsd.Hash(), t.Chain)
				assert.NoError(t, err)
				ntsHashEntries = append(ntsHashEntries, module.NTSHashEntryFormat{
					NetworkTypeID:          ntd.NetworkTypeID(),
					NetworkTypeSectionHash: ntd.NetworkTypeSectionHash(),
				})
				ntsdProofParts = append(ntsdProofParts, pp.Bytes())
			}
		}
		ntsVoteCount, err = bd.NTSVoteCount(pcm)
		assert.NoError(t, err)
	}
	precommit := consensus.NewVoteMessage(
		t.Chain.Wallet(),
		consensus.VoteTypePrecommit,
		t.LastBlock.Height(),
		0,
		t.LastBlock.ID(),
		nil,
		t.LastBlock.Timestamp()+1,
		ntsHashEntries,
		ntsdProofParts,
		ntsVoteCount,
	)
	return consensus.NewCommitVoteList(pcm, precommit)
}

func (t *Node) Address() module.Address {
	return t.Chain.Wallet().Address()
}

func (t *Node) CommonAddress() *common.Address {
	return t.Address().(*common.Address)
}

func (t *Node) WaitForBlock(h int64) module.Block {
	chn, err := t.BM.WaitForBlock(h)
	assert.NoError(t.T, err)
	return <-chn
}

func (t *Node) NewTx() *Transaction {
	blk, err := t.BM.GetLastBlock()
	assert.NoError(t.T, err)
	return NewTx().SetTimestamp(blk.Timestamp())
}

func (t *Node) ProposalBytesFor(blk module.Block, round int32) (pmBS_ []byte, bpmBS_ []byte, bps_ consensus.PartSet) {
	psb := consensus.NewPartSetBuffer(consensus.ConfigBlockPartSize)
	assert.NoError(t, blk.MarshalHeader(psb))
	assert.NoError(t, blk.MarshalBody(psb))
	bps := psb.PartSet()
	msg := consensus.NewProposalMessage()
	msg.Height = blk.Height()
	msg.Round = round
	msg.BlockPartSetID = bps.ID()
	msg.POLRound = -1
	assert.NoError(t, msg.Sign(t.Chain.Wallet()))
	pmBS := codec.MustMarshalToBytes(msg)
	bpm := consensus.BlockPartMessage{}
	assert.Equal(t, 1, bps.Parts())
	bpm.Height = blk.Height()
	bpm.BlockPart = bps.GetPart(0).Bytes()
	bpm.Index = uint16(0)
	bpm.Nonce = 0
	bpmBS := codec.MustMarshalToBytes(bpm)
	return pmBS, bpmBS, bps
}

func (t *Node) InvalidProposalBytesFor(blk module.Block) (pmBS_ []byte, bpmBS_ []byte, bps_ consensus.PartSet) {
	psb := consensus.NewPartSetBuffer(consensus.ConfigBlockPartSize)
	assert.NoError(t, blk.MarshalHeader(psb))
	// write breaking one byte
	_, err := psb.Write([]byte{0})
	assert.NoError(t, err)
	assert.NoError(t, blk.MarshalBody(psb))
	bps := psb.PartSet()
	msg := consensus.NewProposalMessage()
	msg.Height = blk.Height()
	msg.Round = 0
	msg.BlockPartSetID = bps.ID()
	msg.POLRound = -1
	assert.NoError(t, msg.Sign(t.Chain.Wallet()))
	pmBS := codec.MustMarshalToBytes(msg)
	bpm := consensus.BlockPartMessage{}
	assert.Equal(t, 1, bps.Parts())
	bpm.Height = blk.Height()
	bpm.BlockPart = bps.GetPart(0).Bytes()
	bpm.Index = uint16(0)
	bpm.Nonce = 0
	bpmBS := codec.MustMarshalToBytes(bpm)
	return pmBS, bpmBS, bps
}

func (t *Node) VoteFor(vt consensus.VoteType, blk module.Block, bpsID *consensus.PartSetID, round int32) *consensus.VoteMessage {
	return consensus.NewVoteMessage(
		t.Chain.Wallet(),
		vt,
		blk.Height(),
		round,
		blk.ID(),
		bpsID,
		blk.Timestamp()+1,
		nil,
		nil,
		0,
	)
}

func (t *Node) NilVoteFor(vt consensus.VoteType, blk module.Block, r int32) *consensus.VoteMessage {
	return consensus.NewVoteMessage(
		t.Chain.Wallet(),
		vt,
		blk.Height(),
		r,
		codec.MustMarshalToBytes(t.Chain.NID()),
		nil,
		blk.Timestamp()+1,
		nil,
		nil,
		0,
	)
}

func NodeInterconnect(nodes []*Node) {
	l := len(nodes)
	for i := 0; i < l; i++ {
		for j := i + 1; j < l; j++ {
			nodes[i].NM.Connect(nodes[j].NM)
		}
	}
}

func NodeWaitForBlock(nodes []*Node, h int64) module.Block {
	var blk module.Block
	for _, n := range nodes {
		b := n.WaitForBlock(h)
		if blk == nil {
			blk = b
		} else {
			assert.Equal(n.T, blk.ID(), b.ID())
		}
	}
	return blk
}
