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
	database db.Database
	state    *hvhstate.Snapshot
}

func (s *ExtensionSnapshotImpl) Bytes() []byte {
	return codec.BC.MustMarshalToBytes(s)
}

func (s *ExtensionSnapshotImpl) RLPEncodeSelf(e codec.Encoder) error {
	return e.EncodeListOf(
		s.state.Bytes(),
	)
}

func (s *ExtensionSnapshotImpl) RLPDecodeSelf(d codec.Decoder) error {
	return nil
}

func (s *ExtensionSnapshotImpl) Flush() error {
	if err := s.state.Flush(); err != nil {
		return err
	}
	return nil
}

func (s *ExtensionSnapshotImpl) NewState(readonly bool) state.ExtensionState {
	logger := hvhutils.NewLogger(nil)

	return &ExtensionStateImpl{
		database: s.database,
		logger:   logger,
		State:    hvhstate.NewStateFromSnapshot(s.state, readonly, logger),
	}
}

func NewExtensionSnapshot(database db.Database, hash []byte) state.ExtensionSnapshot {
	if hash == nil {
		return &ExtensionSnapshotImpl{
			database: database,
			state:    hvhstate.NewSnapshot(database, nil),
		}
	}
	s := &ExtensionSnapshotImpl{
		database: database,
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
		database: builder.Database(),
		state:    hvhstate.NewSnapshotWithBuilder(builder, hashes[0]),
	}
}

type ExtensionStateImpl struct {
	database db.Database

	logger log.Logger
	State  *hvhstate.State
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
		database: es.database,
		state:    es.State.GetSnapshot(),
	}
}

func (es *ExtensionStateImpl) Reset(isnapshot state.ExtensionSnapshot) {
	//snapshot := isnapshot.(*ExtensionSnapshotImpl)
	//if err := es.State.Reset(snapshot.state); err != nil {
	//	panic(err)
	//}
}

// ClearCache clear cache. It's called before executing first transaction
// and also it could be called at the end of base transaction
func (es *ExtensionStateImpl) ClearCache() {
	//es.State.ClearCache()
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
