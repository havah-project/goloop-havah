/*
 * Copyright 2022 ICON Foundation
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

package chain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/codec"
	"github.com/icon-project/goloop/common/containerdb"
	"github.com/icon-project/goloop/common/db"
	"github.com/icon-project/goloop/common/errors"
	"github.com/icon-project/goloop/common/intconv"
	"github.com/icon-project/goloop/common/log"
	"github.com/icon-project/goloop/common/trie"
	"github.com/icon-project/goloop/common/trie/trie_manager"
	"github.com/icon-project/goloop/havah/hvh/hvhstate"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service"
	"github.com/icon-project/goloop/service/state"
	"github.com/icon-project/goloop/service/txresult"
)

type verifyParams struct {
	Start  int64 `json:"start"`
	End    int64 `json:"end"`
	Detail bool  `json:"detail"`
}

type taskReplay struct {
	chain  *singleChain
	tmpDB  db.LayerDB
	result resultStore
	height int64
	start  int64
	end    int64
	detail bool
	stop   chan error
}

func (t *taskReplay) Stop() {
	t.stop <- errors.ErrInterrupted
}

func (t *taskReplay) Wait() error {
	return t.result.Wait()
}

func (t *taskReplay) String() string {
	return fmt.Sprintf("Replay(start=%d,end=%d,detail=%v)",
		t.start, t.end, t.detail)
}

func (t *taskReplay) DetailOf(s State) string {
	switch s {
	case Started:
		return fmt.Sprintf("replay started height=%d", t.height)
	default:
		return "replay " + s.String()
	}
}

func (t *taskReplay) initTransition() (module.Block, module.Transition, error) {
	sm := t.chain.ServiceManager()
	bm := t.chain.BlockManager()
	blk, err := bm.GetBlockByHeight(t.height)
	if err != nil {
		return nil, nil, err
	}
	tr, err := sm.CreateInitialTransition(blk.Result(), blk.NextValidators())
	return blk, tr, err
}

type transitionCallback chan error

func (t transitionCallback) OnValidate(transition module.Transition, err error) {
	t <- err
}

func (t transitionCallback) OnExecute(transition module.Transition, err error) {
	t <- err
}

type ResultValues struct {
	State          common.HexBytes  `json:"stateHash"`
	PatchReceipts  common.HexBytes  `json:"patchReceipts"`
	NormalReceipts common.HexBytes  `json:"normalReceipts"`
	ExtensionData  *ExtensionValues `json:"extensionData,omitempty"`
}

type ExtensionValues struct {
	State  common.HexBytes `json:"state"`
}

func (ev *ExtensionValues) RLPDecodeSelf(d codec.Decoder) error {
	var bs []byte
	if err := d.Decode(&bs); err != nil {
		return err
	}
	_, err := codec.BC.UnmarshalFromBytes(bs, &ev.State)
	return err
}

type Storer interface {
	Store() trie.Immutable
}

func getBytesDiffHandlerFor(logger log.Logger, name string) func(op int, key []byte, exp, real []byte) {
	return func(op int, key []byte, exp, real []byte) {
		switch op {
		case -1:
			logger.Errorf("%s [-] key=%#x value=<%#x>\n", name, key, exp)
		case 0:
			logger.Errorf("%s [=] key=%#x exp=<%#x> real=<%#x>\n", name, key, exp, real)
		case 1:
			logger.Errorf("%s [+] key=%#x value=<%#x>\n", name, key, real)
		}
	}
}

func getAccountDiffHandlerFor(logger log.Logger, name string) func(op int, key []byte, exp, real trie.Object) {
	return func(op int, key []byte, exp, real trie.Object) {
		switch op {
		case -1:
			logger.Errorf("%s [-] key=%#x value=%+v\n", name, key, exp)
		case 0:
			if exp.Equal(real) {
				logger.Errorf("%s [=] key=%#x exp=<%#x> real=<%#x>\n", name, key, exp.Bytes(), real.Bytes())
			} else {
				logger.Errorf("%s [=] key=%#x exp=%+v real=%+v\n", name, key, exp, real)
				var eStore, rStore trie.Immutable
				if eASS, ok := exp.(Storer); ok {
					eStore = eASS.Store()
				}
				if rASS, ok := real.(Storer); ok {
					rStore = rASS.Store()
				}
				if eStore == rStore {
					return
				}
				if eStore == nil {
					logger.Errorf("%s [+] key=%#x real=%+v", name+".store", key, rStore)
					return
				} else if rStore == nil {
					logger.Errorf("%s [-] key=%#x exp=%+v", name+".store", key, eStore)
					return
				}
				accountHash := fmt.Sprintf("%#x", key)
				err := trie_manager.CompareImmutable(eStore, rStore, getBytesDiffHandlerFor(logger, accountHash))
				if err != nil {
					logger.Errorf("%s fail to compare store", name)
				}
			}
		case 1:
			logger.Errorf("%s [+] key=%#x value=%+v", name, key, real)
		}
	}
}

func getExtensionDiffHandlerFor(logger log.Logger, name string) func(op int, key []byte, exp, real []byte) {
	return func(op int, key []byte, exp, real []byte) {
		if varDBName, ok := getVarDBNameFromKey(key); ok {
			switch op {
			case -1:
				ev := varDBValueToString(varDBName, exp)
				logger.Errorf("%s [-] varDB=%s exp=%s", name, varDBName, ev)
			case 0:
				ev := varDBValueToString(varDBName, exp)
				rv := varDBValueToString(varDBName, real)
				logger.Errorf("%s [=] varDB=%s exp=%s real=%s", name, varDBName, ev, rv)
			case 1:
				rv := varDBValueToString(varDBName, real)
				logger.Errorf("%s [+] varDB=%s real=%s", name, varDBName, rv)
			}
		}

		switch op {
		case -1:
			logger.Errorf("%s [-] key=%#x value=<%#x>\n", name, key, exp)
		case 0:
			logger.Errorf("%s [=] key=%#x exp=<%#x> real=<%#x>\n", name, key, exp, real)
		case 1:
			logger.Errorf("%s [+] key=%#x value=<%#x>\n", name, key, real)
		}
	}
}

func showExtensionDiff(dbase db.Database, logger log.Logger, e, r *ExtensionValues) {
	if !bytes.Equal(e.State.Bytes(), r.State.Bytes()) {
		et := trie_manager.NewImmutable(dbase, e.State.Bytes())
		rt := trie_manager.NewImmutable(dbase, r.State.Bytes())
		if err := trie_manager.CompareImmutable(
			et, rt, getExtensionDiffHandlerFor(logger, "ext.state")); err != nil {
			logger.Errorf("Fail to compare ext.state err=%+v", err)
		}
	}
}

type ToJSONer interface {
	ToJSON(version module.JSONVersion) (interface{}, error)
}

func JSONMarshalIndent(obj interface{}) ([]byte, error) {
	if jsoner, ok := obj.(ToJSONer); ok {
		if jso, err := jsoner.ToJSON(module.JSONVersionLast); err == nil {
			obj = jso
		} else {
			log.Warnf("Failure in ToJSON err=%+v", err)
		}
	}
	return json.MarshalIndent(obj, "", "  ")
}

func showResultDiff(dbase db.Database, logger log.Logger, e, r *ResultValues) {
	logger.Tracef("showResultDiff() start")
	defer logger.Tracef("showResultDiff() end")

	if !bytes.Equal(e.State.Bytes(), r.State.Bytes()) {
		et := trie_manager.NewImmutableForObject(dbase, e.State.Bytes(), state.AccountType)
		rt := trie_manager.NewImmutableForObject(dbase, r.State.Bytes(), state.AccountType)
		trie_manager.CompareImmutableForObject(et, rt, getAccountDiffHandlerFor(logger, "world"))
	}

	if !bytes.Equal(e.NormalReceipts.Bytes(), r.NormalReceipts.Bytes()) {
		el := txresult.NewReceiptListFromHash(dbase, e.NormalReceipts.Bytes())
		rl := txresult.NewReceiptListFromHash(dbase, r.NormalReceipts.Bytes())
		idx := 0
		for expect, result := el.Iterator(), rl.Iterator(); expect.Has() && result.Has(); _, _, idx = expect.Next(), result.Next(), idx+1 {
			rct1, _ := expect.Get()
			rct2, _ := result.Get()
			if err := rct1.Check(rct2); err != nil {
				rct1js, _ := JSONMarshalIndent(rct1)
				rct2js, _ := JSONMarshalIndent(rct2)
				logger.Errorf("Expected Receipt[%d]:%s", idx, rct1js)
				logger.Errorf("Returned Receipt[%d]:%s", idx, rct2js)
			}
		}
	}

	showExtensionDiff(dbase, logger, e.ExtensionData, r.ExtensionData)
}

func (t *taskReplay) showResultDiff(logger log.Logger, exp, real []byte) error {
	var eResult, rResult ResultValues
	if _, err := codec.BC.UnmarshalFromBytes(exp, &eResult); err != nil {
		return err
	}
	if _, err := codec.BC.UnmarshalFromBytes(real, &rResult); err != nil {
		return err
	}
	showResultDiff(t.tmpDB, logger, &eResult, &rResult)
	return nil
}

func (t *taskReplay) doReplay() error {
	defer func() {
		t.chain.releaseManagers()
		t.chain.database = t.tmpDB.Unwrap()
	}()

	t.height = t.start

	bm := t.chain.BlockManager()
	sm := t.chain.ServiceManager()
	logger := t.chain.Logger()

	end := t.end
	if last, err := bm.GetLastBlock(); err != nil {
		return err
	} else {
		lastHeight := last.Height()
		if end == 0 || end > lastHeight-1 {
			end = lastHeight - 1
		}
	}

	blk, ptr, err := t.initTransition()
	if err != nil {
		return err
	}
	var nblk module.Block
	var tr module.Transition
	for t.height <= end {
		logger.Tracef("loop: height=%d", t.height)

		// next block for votes and consensus information
		nblk, err = bm.GetBlockByHeight(t.height + 1)
		if err != nil {
			return err
		}
		csi, err := bm.NewConsensusInfo(blk)
		if err != nil {
			return err
		}
		tr, err = sm.CreateTransition(ptr, blk.NormalTransactions(), blk, csi, true)
		if err != nil {
			return err
		}
		ptxs := nblk.PatchTransactions()
		if len(ptxs.Hash()) > 0 {
			tr = sm.PatchTransition(tr, ptxs, nblk)
		}
		cb := make(chan error, 2)
		cancel, err := tr.Execute(transitionCallback(cb))
		if err != nil {
			return err
		}

		// wait for OnValidate
		select {
		case err := <-t.stop:
			cancel()
			return err
		case err := <-cb:
			if err != nil {
				return err
			}
		}

		// wait for OnExecute
		select {
		case err := <-t.stop:
			cancel()
			return err
		case err := <-cb:
			if err != nil {
				return err
			}
		}

		// check the result
		if !bytes.Equal(tr.Result(), nblk.Result()) {
			logger.Errorf("INVALID RESULT res=%#x exp=%#x",
				tr.Result(), nblk.Result())
			if t.detail {
				_ = sm.Finalize(tr, module.FinalizeResult)
				if err := t.showResultDiff(logger, nblk.Result(), tr.Result()); err != nil {
					logger.Errorf("FAIL to show diff err=%+v", err)
				}
			}
			return errors.InvalidStateError.New("InvalidResult")
		} else {
			if err := service.FinalizeTransition(tr,
				module.FinalizeNormalTransaction|module.FinalizePatchTransaction|module.FinalizeResult,
				false); err != nil {
				return err
			}
			_ = t.tmpDB.Flush(false)
		}
		t.height += 1
		ptr, tr = tr, nil
		blk, nblk = nblk, nil
	}
	return nil
}

func (t *taskReplay) Start() error {
	t.tmpDB = db.NewLayerDB(t.chain.database)
	t.chain.database = t.tmpDB

	if err := t.chain.prepareManagers(); err != nil {
		t.chain.database = t.tmpDB.Unwrap()
		t.result.SetValue(err)
		return err
	}
	t.stop = make(chan error, 1)
	go func() {
		err := t.doReplay()
		defer t.result.SetValue(err)
	}()
	return nil
}

func taskReplayFactory(chain *singleChain, param json.RawMessage) (chainTask, error) {
	var p verifyParams
	if err := json.Unmarshal(param, &p); err != nil {
		return nil, err
	}
	if (p.End != 0 && p.End < p.Start) || p.Start < 0 {
		return nil, errors.IllegalArgumentError.Errorf(
			"InvalidParameter(start=%d,end=%d)", p.Start, p.End)
	}
	task := &taskReplay{
		chain:  chain,
		start:  p.Start,
		end:    p.End,
		detail: p.Detail,
	}
	return task, nil
}

var keyToVarDBNameMap map[string]string
func initVarDBNames() map[string]string {
	keyToName := make(map[string]string)
	varDBNames := []string{
		hvhmodule.VarIssueAmount,
		hvhmodule.VarIssueStart,
		hvhmodule.VarIssueLimit,
		hvhmodule.VarIssueReductionCycle,
		hvhmodule.VarHooverBudget,
		hvhmodule.VarUSDTPrice,
		hvhmodule.VarActiveUSDTPrice,
		hvhmodule.VarAllPlanet,
		hvhmodule.VarActivePlanet,
		hvhmodule.VarWorkingPlanet,
		hvhmodule.VarRewardTotal,
		hvhmodule.VarRewardRemain,
		hvhmodule.VarEcoReward,
		hvhmodule.VarPrivateClaimableRate,
		hvhmodule.VarLost,
		hvhmodule.VarBlockVoteCheckPeriod,
		hvhmodule.VarNonVoteAllowance,
		hvhmodule.VarActiveValidatorCount,
		hvhmodule.VarSubValidatorsIndex,
		hvhmodule.VarNetworkStatus,
	}

	for _, name := range varDBNames {
		key := containerdb.ToKey(containerdb.HashBuilder, name).Build()
		keyToName[string(key)] = name
	}
	return keyToName
}

func getVarDBNameFromKey(key []byte) (string, bool) {
	if keyToVarDBNameMap == nil {
		keyToVarDBNameMap = initVarDBNames()
	}
	name, ok := keyToVarDBNameMap[string(key)]
	return name, ok
}

func varDBValueToString(varDBName string, bs []byte) string {
	switch varDBName {
	case hvhmodule.VarNetworkStatus:
		ns, _ := hvhstate.NewNetworkStatusFromBytes(bs)
		return ns.String()
	default:
		return fmt.Sprintf("%d", intconv.BigIntSetBytes(new(big.Int), bs))
	}
}

func init() {
	registerTaskFactory("replay", taskReplayFactory)
}