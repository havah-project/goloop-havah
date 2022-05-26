/*
 * Copyright 2020 ICON Foundation
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

package hvh

import (
	"math/big"

	"github.com/icon-project/goloop/common/codec"
	"github.com/icon-project/goloop/common/db"
	"github.com/icon-project/goloop/common/log"
	"github.com/icon-project/goloop/common/merkle"
	"github.com/icon-project/goloop/havah/hvh/hvhstate"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/havah/hvhutils"
	"github.com/icon-project/goloop/service/state"
)

type ExtensionSnapshotImpl struct {
	dbase db.Database
	state *hvhstate.Snapshot
}

func (ess *ExtensionSnapshotImpl) Bytes() []byte {
	return codec.BC.MustMarshalToBytes(ess)
}

func (ess *ExtensionSnapshotImpl) RLPEncodeSelf(e codec.Encoder) error {
	return e.Encode(ess.state.Bytes())
}

func (ess *ExtensionSnapshotImpl) RLPDecodeSelf(d codec.Decoder) error {
	var stateHash []byte
	if err := d.Decode(&stateHash); err != nil {
		return err
	}
	ess.state = hvhstate.NewSnapshot(ess.dbase, stateHash)
	return nil
}

func (ess *ExtensionSnapshotImpl) Flush() error {
	if err := ess.state.Flush(); err != nil {
		return err
	}
	return nil
}

func (ess *ExtensionSnapshotImpl) NewState(readonly bool) state.ExtensionState {
	logger := hvhutils.NewLogger(nil)

	return &ExtensionStateImpl{
		dbase:  ess.dbase,
		logger: logger,
		state:  hvhstate.NewStateFromSnapshot(ess.state, readonly, logger),
	}
}

func NewExtensionSnapshot(dbase db.Database, hash []byte) state.ExtensionSnapshot {
	if hash == nil {
		return &ExtensionSnapshotImpl{
			dbase: dbase,
			state: hvhstate.NewSnapshot(dbase, nil),
		}
	}
	s := &ExtensionSnapshotImpl{
		dbase: dbase,
	}
	if _, err := codec.BC.UnmarshalFromBytes(hash, s); err != nil {
		return nil
	}
	return s
}

func NewExtensionSnapshotWithBuilder(builder merkle.Builder, raw []byte) state.ExtensionSnapshot {
	var hashes [5][]byte
	if _, err := codec.BC.UnmarshalFromBytes(raw, &hashes); err != nil {
		return nil
	}
	return &ExtensionSnapshotImpl{
		dbase: builder.Database(),
		state: hvhstate.NewSnapshotWithBuilder(builder, hashes[0]),
	}
}

// ==================================================================

type ExtensionStateImpl struct {
	dbase db.Database

	logger log.Logger
	state  *hvhstate.State
}

func (es *ExtensionStateImpl) Logger() log.Logger {
	return es.logger
}

func (es *ExtensionStateImpl) SetLogger(logger log.Logger) {
	if logger != nil {
		es.logger = logger
	}
}

func (es *ExtensionStateImpl) GetSnapshot() state.ExtensionSnapshot {
	return &ExtensionSnapshotImpl{
		dbase: es.dbase,
		state: es.state.GetSnapshot(),
	}
}

func (es *ExtensionStateImpl) Reset(ess state.ExtensionSnapshot) {
	snapshot := ess.(*ExtensionSnapshotImpl)
	if err := es.state.Reset(snapshot.state); err != nil {
		panic(err)
	}
}

// ClearCache is called before executing the first transaction in a block and at the end of base transaction
func (es *ExtensionStateImpl) ClearCache() {
	//es.state.ClearCache()
}

func (es *ExtensionStateImpl) OnExecutionBegin(wc hvhmodule.WorldContext) error {
	return nil
}

func (es *ExtensionStateImpl) OnExecutionEnd(wc hvhmodule.WorldContext) error {
	return nil
}

func (es *ExtensionStateImpl) OnTransactionEnd(blockHeight int64, success bool) error {
	return nil
}

func (es *ExtensionStateImpl) GetUSDTPrice() (*big.Int, error) {
	return es.state.GetUSDTPrice(), nil
}

func (es *ExtensionStateImpl) SetUSDTPrice(price *big.Int) error {
	return es.state.SetUSDTPrice(price)
}

func (es *ExtensionStateImpl) GetIssueInfo(cc hvhmodule.CallContext) (map[string]interface{}, error) {
	height := cc.BlockHeight()
	issueStart := es.state.GetIssueStart() // in height
	termPeriod := es.state.GetTermPeriod() // in height

	jso := map[string]interface{}{
		"height": height,
	}
	if issueStart > 0 {
		termSeq := (height - issueStart) / termPeriod
		jso["termSequence"] = termSeq
	}
	return jso, nil
}

func (es *ExtensionStateImpl) StartRewardIssue(height int64) error {
	return es.state.SetIssueStart(height)
}
