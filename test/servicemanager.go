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
	"github.com/icon-project/goloop/chain/base"
	"github.com/icon-project/goloop/common/db"
	"github.com/icon-project/goloop/common/log"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service"
	"github.com/icon-project/goloop/service/contract"
	"github.com/icon-project/goloop/service/eeproxy"
	"github.com/icon-project/goloop/service/scoredb"
	"github.com/icon-project/goloop/service/state"
	"github.com/icon-project/goloop/service/transaction"
)

type ServiceManager struct {
	module.ServiceManager
	dbase            db.Database
	logger           log.Logger
	plt              base.Platform
	cm               contract.ContractManager
	em               eeproxy.Manager
	chain            module.Chain
	tsc              *service.TxTimestampChecker
	emptyTXs         module.TransactionList
	nextBlockVersion int
	pool             []module.Transaction
}

func NewServiceManager(
	c *Chain,
	plt base.Platform,
	cm contract.ContractManager,
	em eeproxy.Manager,
) *ServiceManager {
	dbase := c.Database()
	return &ServiceManager{
		dbase:            dbase,
		logger:           c.Logger(),
		plt:              plt,
		cm:               cm,
		em:               em,
		chain:            c,
		tsc:              service.NewTimestampChecker(),
		emptyTXs:         transaction.NewTransactionListFromSlice(dbase, nil),
		nextBlockVersion: module.BlockVersion2,
	}
}

func (sm *ServiceManager) TransactionFromBytes(b []byte, blockVersion int) (module.Transaction, error) {
	return transaction.NewTransaction(b)
}

func (sm *ServiceManager) ProposeTransition(parent module.Transition, bi module.BlockInfo, csi module.ConsensusInfo) (module.Transition, error) {
	txs := transaction.NewTransactionListFromSlice(sm.dbase, sm.pool)
	sm.pool = nil
	return service.NewTransition(
		parent,
		sm.emptyTXs,
		txs,
		bi,
		csi,
		true,
	), nil
}

func (sm *ServiceManager) CreateInitialTransition(
	result []byte,
	nextValidators module.ValidatorList,
) (module.Transition, error) {
	return service.NewInitTransition(
		sm.dbase,
		result,
		nextValidators,
		sm.cm,
		sm.em,
		sm.chain,
		sm.logger,
		sm.plt,
		sm.tsc,
	)
}

func (sm *ServiceManager) CreateTransition(parent module.Transition, txs module.TransactionList, bi module.BlockInfo, csi module.ConsensusInfo, validated bool) (module.Transition, error) {
	return service.NewTransition(
		parent,
		sm.emptyTXs,
		txs,
		bi,
		csi,
		validated,
	), nil
}

func (sm *ServiceManager) GetPatches(parent module.Transition, bi module.BlockInfo) module.TransactionList {
	return sm.emptyTXs
}

func (sm *ServiceManager) PatchTransition(transition module.Transition, patches module.TransactionList, bi module.BlockInfo) module.Transition {
	return transition
}

func (sm *ServiceManager) CreateSyncTransition(transition module.Transition, result []byte, vlHash []byte, noBuffer bool) module.Transition {
	panic("implement me")
}

func (sm *ServiceManager) Finalize(transition module.Transition, opt int) error {
	return service.FinalizeTransition(transition, opt, false)
}

func (sm *ServiceManager) WaitForTransaction(parent module.Transition, bi module.BlockInfo, cb func()) bool {
	panic("implement me")
}

func (sm *ServiceManager) GetChainID(result []byte) (int64, error) {
	return int64(sm.chain.CID()), nil
}

func (sm *ServiceManager) GetNetworkID(result []byte) (int64, error) {
	return int64(sm.chain.NID()), nil
}

func (sm *ServiceManager) GetMembers(result []byte) (module.MemberList, error) {
	return nil, nil
}

func (sm *ServiceManager) GetRoundLimit(result []byte, vl int) int64 {
	ws, err := service.NewWorldSnapshot(sm.dbase, sm.plt, result, nil)
	if err != nil {
		return 0
	}
	ass := ws.GetAccountSnapshot(state.SystemID)
	as := scoredb.NewStateStoreWith(ass)
	if as == nil {
		return 0
	}
	factor := scoredb.NewVarDB(as, state.VarRoundLimitFactor).Int64()
	if factor == 0 {
		return 0
	}
	limit := contract.RoundLimitFactorToRound(vl, factor)
	return limit
}

func (sm *ServiceManager) GetMinimizeBlockGen(result []byte) bool {
	ws, err := service.NewWorldSnapshot(sm.dbase, sm.plt, result, nil)
	if err != nil {
		return false
	}
	ass := ws.GetAccountSnapshot(state.SystemID)
	as := scoredb.NewStateStoreWith(ass)
	if as == nil {
		return false
	}
	return scoredb.NewVarDB(as, state.VarMinimizeBlockGen).Bool()
}

func (sm *ServiceManager) GetNextBlockVersion(result []byte) int {
	if result == nil {
		return sm.plt.DefaultBlockVersionFor(sm.chain.CID())
	}
	ws, err := service.NewWorldSnapshot(sm.dbase, sm.plt, result, nil)
	if err != nil {
		return -1
	}
	ass := ws.GetAccountSnapshot(state.SystemID)
	if ass == nil {
		return sm.plt.DefaultBlockVersionFor(sm.chain.CID())
	}
	as := scoredb.NewStateStoreWith(ass)
	v := int(scoredb.NewVarDB(as, state.VarNextBlockVersion).Int64())
	if v == 0 {
		return sm.plt.DefaultBlockVersionFor(sm.chain.CID())
	}
	return v
}

func (sm *ServiceManager) ImportResult(result []byte, vh []byte, src db.Database) error {
	panic("implement me")
}

func (sm *ServiceManager) GenesisTransactionFromBytes(b []byte, blockVersion int) (module.Transaction, error) {
	return transaction.NewGenesisTransaction(b)
}

func (sm *ServiceManager) TransactionListFromHash(hash []byte) module.TransactionList {
	return transaction.NewTransactionListFromHash(sm.dbase, hash)
}

func (sm *ServiceManager) ReceiptListFromResult(result []byte, g module.TransactionGroup) (module.ReceiptList, error) {
	panic("implement me")
}

func (sm *ServiceManager) SendTransaction(result []byte, height int64, tx interface{}) ([]byte, error) {
	t, err := transaction.NewTransactionFromJSON(([]byte)(tx.(string)))
	if err != nil {
		return nil, err
	}
	sm.pool = append(sm.pool, t)
	return t.ID(), nil
}

func (sm *ServiceManager) ValidatorListFromHash(hash []byte) module.ValidatorList {
	vl, err := state.ValidatorSnapshotFromHash(sm.dbase, hash)
	if err != nil {
		return nil
	}
	return vl
}

func (sm *ServiceManager) TransactionListFromSlice(txs []module.Transaction, version int) module.TransactionList {
	switch version {
	case module.BlockVersion0:
		return transaction.NewTransactionListV1FromSlice(txs)
	case module.BlockVersion1, module.BlockVersion2:
		return transaction.NewTransactionListFromSlice(sm.chain.Database(), txs)
	default:
		return nil
	}
}

func (sm *ServiceManager) SendTransactionAndWait(result []byte, height int64, tx interface{}) ([]byte, <-chan interface{}, error) {
	panic("implement me")
}

func (sm *ServiceManager) WaitTransactionResult(id []byte) (<-chan interface{}, error) {
	panic("implement me")
}

func (sm *ServiceManager) ExportResult(result []byte, vh []byte, dst db.Database) error {
	panic("implement me")
}
