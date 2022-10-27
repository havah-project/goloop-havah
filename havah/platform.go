package havah

import (
	"encoding/json"

	"github.com/icon-project/goloop/chain/base"
	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/db"
	"github.com/icon-project/goloop/common/errors"
	"github.com/icon-project/goloop/common/log"
	"github.com/icon-project/goloop/common/merkle"
	"github.com/icon-project/goloop/consensus"
	"github.com/icon-project/goloop/havah/hvh"
	"github.com/icon-project/goloop/havah/hvh/hvhstate"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/contract"
	"github.com/icon-project/goloop/service/platform/basic"
	"github.com/icon-project/goloop/service/state"
	"github.com/icon-project/goloop/service/transaction"
	"github.com/icon-project/goloop/service/txresult"
)

type platform struct {
	base string
	cid  int
}

func (p *platform) NewContractManager(
	dbase db.Database, dir string, logger log.Logger) (contract.ContractManager, error) {
	return newContractManager(p, dbase, dir, logger)
}

func (p *platform) NewExtensionSnapshot(dbase db.Database, raw []byte) state.ExtensionSnapshot {
	// TODO return valid ExtensionSnapshot(not nil) which can return valid ExtensionState.
	//  with that state, we may change state of extension.
	//  For initial state, the snapshot returns nil for Bytes() method.
	if len(raw) == 0 {
		return nil
	}
	return hvh.NewExtensionSnapshot(dbase, raw)
}

func (p *platform) NewExtensionWithBuilder(builder merkle.Builder, raw []byte) state.ExtensionSnapshot {
	return hvh.NewExtensionSnapshotWithBuilder(builder, raw)
}

func (p *platform) ToRevision(value int) module.Revision {
	return hvhmodule.ValueToRevision(value)
}

func (p *platform) NewBaseTransaction(wc state.WorldContext) (module.Transaction, error) {
	height := wc.BlockHeight()
	es := hvh.GetExtensionStateFromWorldContext(wc, nil)
	if es == nil {
		return nil, nil
	}

	// The block height when coin issuing is started
	issueStart := es.GetIssueStart()
	if !hvhstate.IsIssueStarted(height, issueStart) {
		return nil, nil
	}

	termPeriod := es.GetTermPeriod()
	_, blockIndex := hvhstate.GetTermSequenceAndBlockIndex(height, issueStart, termPeriod)
	if blockIndex != 0 {
		return nil, nil
	}

	t := common.HexInt64{Value: wc.BlockTimeStamp()}
	v := common.HexUint16{Value: module.TransactionVersion3}
	mtx := map[string]interface{}{
		"timestamp": t,
		"version":   v,
		"dataType":  "base",
		"data":      es.NewBaseTransactionData(height, issueStart),
	}
	bs, err := json.Marshal(mtx)
	if err != nil {
		return nil, err
	}
	tx, err := transaction.NewTransactionFromJSON(bs)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (p *platform) OnExtensionSnapshotFinalization(state.ExtensionSnapshot, log.Logger) {
	// Do nothing
}

func checkBaseTX(txs module.TransactionList) bool {
	tx, err := txs.Get(0)
	if err == nil {
		return hvh.CheckBaseTX(tx)
	} else {
		return false
	}
}

func (p *platform) OnValidateTransactions(wc state.WorldContext, patches, txs module.TransactionList) error {
	needBaseTX := false
	es := hvh.GetExtensionStateFromWorldContext(wc, nil)
	if es != nil {
		issueStart := es.GetIssueStart()
		termPeriod := es.GetTermPeriod()
		termSeq, blockIndex := hvhstate.GetTermSequenceAndBlockIndex(wc.BlockHeight(), issueStart, termPeriod)
		needBaseTX = termSeq >= 0 && blockIndex == 0
		log.GlobalLogger().Debugf(
			"is=%d tperiod=%d termSeq=%d blockIndex=%d needBaseTX=%t",
			issueStart, termPeriod, termSeq, blockIndex, needBaseTX,
		)
	}
	if hasBaseTX := checkBaseTX(txs); needBaseTX == hasBaseTX {
		return nil
	} else {
		if needBaseTX {
			return errors.IllegalArgumentError.New("NoBaseTransaction")
		} else {
			return errors.IllegalArgumentError.New("InvalidBaseTransaction")
		}
	}
}

func (p *platform) OnExecutionBegin(wc state.WorldContext, logger log.Logger) error {
	return nil
}

func (p *platform) OnExecutionEnd(wc state.WorldContext, _ base.ExecutionResult, logger log.Logger) error {
	return nil
}

func (p *platform) OnTransactionEnd(wc state.WorldContext, logger log.Logger, rct txresult.Receipt) error {
	return nil
}

// Term means 'Terminate'
func (p *platform) Term() {
	// Terminate
}

func (p *platform) DefaultBlockVersionFor(cid int) int {
	return basic.Platform.DefaultBlockVersionFor(cid)
}

func (p *platform) NewBlockHandlers(c base.Chain) []base.BlockHandler {
	return basic.Platform.NewBlockHandlers(c)
}

func (p *platform) NewConsensus(c base.Chain, walDir string) (module.Consensus, error) {
	return basic.Platform.NewConsensus(c, walDir)
}

func (p *platform) CommitVoteSetDecoder() module.CommitVoteSetDecoder {
	return func(bytes []byte) module.CommitVoteSet {
		return consensus.NewCommitVoteSetFromBytes(bytes)
	}
}

func NewPlatform(base string, cid int) (base.Platform, error) {
	return &platform{
		base: base,
		cid:  cid,
	}, nil
}

func init() {
	hvh.RegisterBaseTx()
}
