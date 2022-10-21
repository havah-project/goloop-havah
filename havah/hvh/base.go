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

// -build base

package hvh

import (
	"bytes"
	"encoding/json"
	"math/big"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/codec"
	"github.com/icon-project/goloop/common/crypto"
	"github.com/icon-project/goloop/common/errors"
	"github.com/icon-project/goloop/havah/hvh/hvhstate"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/contract"
	"github.com/icon-project/goloop/service/scoredb"
	"github.com/icon-project/goloop/service/scoreresult"
	"github.com/icon-project/goloop/service/state"
	"github.com/icon-project/goloop/service/transaction"
	"github.com/icon-project/goloop/service/txresult"
)

type baseDataJSON struct {
	IssueAmount common.HexInt `json:"issueAmount"`
}

func parseBaseData(data []byte) (*baseDataJSON, error) {
	if data == nil {
		return nil, nil
	}
	jso := new(baseDataJSON)
	jd := json.NewDecoder(bytes.NewBuffer(data))
	jd.DisallowUnknownFields()
	if err := jd.Decode(jso); err != nil {
		return nil, err
	}
	return jso, nil
}

type baseV3Data struct {
	Version   common.HexUint16 `json:"version"`
	From      *common.Address  `json:"from,omitempty"` // it should be nil
	TimeStamp common.HexInt64  `json:"timestamp"`
	DataType  string           `json:"dataType,omitempty"`
	Data      json.RawMessage  `json:"data,omitempty"`
}

func (tx *baseV3Data) calcHash() ([]byte, error) {
	sha := bytes.NewBuffer(nil)
	sha.Write([]byte("icx_sendTransaction"))

	// data
	if tx.Data != nil {
		sha.Write([]byte(".data."))
		if len(tx.Data) > 0 {
			var obj interface{}
			if err := json.Unmarshal(tx.Data, &obj); err != nil {
				return nil, err
			}
			if bs, err := transaction.SerializeValue(obj); err != nil {
				return nil, err
			} else {
				sha.Write(bs)
			}
		}
	}

	// dataType
	sha.Write([]byte(".dataType."))
	sha.Write([]byte(tx.DataType))

	// timestamp
	sha.Write([]byte(".timestamp."))
	sha.Write([]byte(tx.TimeStamp.String()))

	// version
	sha.Write([]byte(".version."))
	sha.Write([]byte(tx.Version.String()))

	return crypto.SHA3Sum256(sha.Bytes()), nil
}

type baseV3 struct {
	baseV3Data

	id    []byte
	hash  []byte
	bytes []byte
}

func (tx *baseV3) Version() int {
	return module.TransactionVersion3
}

func (tx *baseV3) Prepare(ctx contract.Context) (state.WorldContext, error) {
	lq := []state.LockRequest{
		{state.WorldIDStr, state.AccountWriteLock},
	}
	wc := ctx.GetFuture(lq)
	wc.WorldVirtualState().Ensure()

	return wc, nil
}

func (tx *baseV3) Execute(ctx contract.Context, _ state.WorldSnapshot, estimate bool) (txresult.Receipt, error) {
	if estimate {
		return nil, errors.InvalidStateError.New("EstimationNotAllowed")
	}
	info := ctx.TransactionInfo()
	if info == nil {
		return nil, errors.InvalidStateError.New("TransactionInfoUnavailable")
	}
	if info.Index != 0 {
		return nil, errors.CriticalFormatError.New("BaseMustBeTheFirst")
	}

	cc := contract.NewCallContext(ctx, ctx.GetStepLimit(state.StepLimitTypeInvoke), false)
	defer cc.Dispose()

	icc := NewCallContext(cc, tx.From())
	es := GetExtensionStateFromWorldContext(cc, cc.Logger())
	if es == nil {
		return nil, errors.InvalidStateError.New("ExtensionIsNotReady")
	}
	if err := es.OnBaseTx(icc, tx.Data); err != nil {
		return nil, err
	}

	// Make a receipt
	r := txresult.NewReceipt(ctx.Database(), ctx.Revision(), cc.Treasury())
	cc.GetEventLogs(r)
	r.SetResult(module.StatusSuccess, hvhmodule.BigIntZero, hvhmodule.BigIntZero, nil)
	es.ClearCache()
	return r, nil
}

func (tx *baseV3) Dispose() {
	// panic("implement me")
}

func (tx *baseV3) Group() module.TransactionGroup {
	return module.TransactionGroupNormal
}

func (tx *baseV3) ID() []byte {
	if tx.id == nil {
		if bs, err := tx.baseV3Data.calcHash(); err != nil {
			panic(err)
		} else {
			tx.id = bs
		}
	}
	return tx.id
}

func (tx *baseV3) From() module.Address {
	return state.SystemAddress
}

func (tx *baseV3) Bytes() []byte {
	if tx.bytes == nil {
		if bs, err := codec.BC.MarshalToBytes(&tx.baseV3Data); err != nil {
			panic(err)
		} else {
			tx.bytes = bs
		}
	}
	return tx.bytes
}

func (tx *baseV3) Hash() []byte {
	if tx.hash == nil {
		tx.hash = crypto.SHA3Sum256(tx.Bytes())
	}
	return tx.hash
}

func (tx *baseV3) Verify() error {
	return nil
}

func (tx *baseV3) ToJSON(version module.JSONVersion) (interface{}, error) {
	jso := map[string]interface{}{
		"version":   &tx.baseV3Data.Version,
		"timestamp": &tx.baseV3Data.TimeStamp,
		"dataType":  tx.baseV3Data.DataType,
		"data":      tx.baseV3Data.Data,
	}
	jso["txHash"] = common.HexBytes(tx.ID())
	return jso, nil
}

func (tx *baseV3) ValidateNetwork(nid int) bool {
	return true
}

func (tx *baseV3) PreValidate(wc state.WorldContext, update bool) error {
	return nil
}

func (tx *baseV3) GetHandler(cm contract.ContractManager) (transaction.Handler, error) {
	return tx, nil
}

func (tx *baseV3) Timestamp() int64 {
	return tx.baseV3Data.TimeStamp.Value
}

func (tx *baseV3) Nonce() *big.Int {
	return nil
}

func (tx *baseV3) To() module.Address {
	return state.SystemAddress
}

func (tx *baseV3) IsSkippable() bool {
	return false
}

func checkBaseV3JSON(jso map[string]interface{}) bool {
	if d, ok := jso["dataType"]; !ok || d != "base" {
		return false
	}
	if v, ok := jso["version"]; !ok || v != "0x3" {
		return false
	}
	return true
}

func parseBaseV3JSON(bs []byte, raw bool) (transaction.Transaction, error) {
	tx := new(baseV3)
	if err := json.Unmarshal(bs, &tx.baseV3Data); err != nil {
		return nil, transaction.InvalidFormat.Wrap(err, "InvalidJSON")
	}
	if tx.baseV3Data.From != nil {
		return nil, transaction.InvalidFormat.New("InvalidFromValue(NonNil)")
	}
	return tx, nil
}

type baseV3Header struct {
	Version common.HexUint16 `json:"version"`
	From    *common.Address  `json:"from"` // it should be nil
}

func checkBaseV3Bytes(bs []byte) bool {
	var vh baseV3Header
	if _, err := codec.BC.UnmarshalFromBytes(bs, &vh); err != nil {
		return false
	}
	return vh.From == nil
}

func parseBaseV3Bytes(bs []byte) (transaction.Transaction, error) {
	tx := new(baseV3)
	if _, err := codec.BC.UnmarshalFromBytes(bs, &tx.baseV3Data); err != nil {
		return nil, err
	}
	return tx, nil
}

func RegisterBaseTx() {
	transaction.RegisterFactory(&transaction.Factory{
		Priority:    15,
		CheckJSON:   checkBaseV3JSON,
		ParseJSON:   parseBaseV3JSON,
		CheckBinary: checkBaseV3Bytes,
		ParseBinary: parseBaseV3Bytes,
	})
}

func (es *ExtensionStateImpl) OnBaseTx(cc hvhmodule.CallContext, data []byte) error {
	height := cc.BlockHeight()
	es.Logger().Debugf("OnBaseTx() start: height=%d", height)

	issueStart := es.state.GetIssueStart()
	if !hvhstate.IsIssueStarted(height, issueStart) {
		return errors.InvalidStateError.Errorf(
			"IssueDoesntStarted(height=%d,issueStart=%d)", height, issueStart)
	}

	baseData, err := parseBaseData(data)
	if err != nil {
		return transaction.InvalidFormat.Wrap(err, "Failed to parse baseData")
	}

	termPeriod := es.state.GetTermPeriod()
	termSeq, blockIndexInTerm := hvhstate.GetTermSequenceAndBlockIndex(height, issueStart, termPeriod)
	es.Logger().Debugf(
		"height=%d istart=%d tperiod=%d blockIndex=%d tseq=%d issue=%v",
		height, issueStart, termPeriod, blockIndexInTerm, termSeq, baseData.IssueAmount.Value())

	if blockIndexInTerm != 0 {
		return errors.InvalidStateError.Errorf("InvalidBaseTx")
	}

	// The code below should be executed only at the first block of each term
	issueLimit := es.state.GetIssueLimit()
	if termSeq > 0 && (issueLimit == 0 || termSeq <= issueLimit) {
		if err = es.onTermEnd(cc); err != nil {
			return err
		}
	}
	if issueLimit == 0 || termSeq < issueLimit {
		if err = es.onTermStart(cc, termSeq, baseData); err != nil {
			return err
		}
	}
	es.Logger().Debugf("OnBaseTx() end: height=%d", height)
	return nil
}

func (es *ExtensionStateImpl) onTermEnd(cc hvhmodule.CallContext) error {
	height := cc.BlockHeight()
	es.Logger().Debugf("onTermEnd() start: height=%d", height)

	var err error
	if err = es.TransferEcoSystemReward(cc); err != nil {
		return err
	}

	if err = es.distributeFees(cc); err != nil {
		return err
	}

	if err = es.TransferMissedReward(cc); err != nil {
		return err
	}

	if err = es.refillHooverFund(cc); err != nil {
		return err
	}
	if err = es.state.OnTermEnd(); err != nil {
		return err
	}

	es.Logger().Debugf("onTermEnd() end: height=%d", height)
	return nil
}

func (es *ExtensionStateImpl) TransferEcoSystemReward(cc hvhmodule.CallContext) error {
	es.Logger().Debugf("TransferEcoSystemReward() start")

	reward, err := es.state.ClaimEcoSystemReward()
	if err != nil {
		return err
	}
	if err = cc.Transfer(hvhmodule.PublicTreasury, hvhmodule.EcoSystem, reward, module.Transfer); err != nil {
		return err
	}

	es.Logger().Debugf("TransferEcoSystemReward() end: reward=%d", reward)
	return nil
}

func (es *ExtensionStateImpl) TransferMissedReward(cc hvhmodule.CallContext) error {
	es.Logger().Debugf("TransferMissedReward() start")

	missed, err := es.state.ClaimMissedReward()
	if err != nil {
		return err
	}
	balance := cc.GetBalance(hvhmodule.PublicTreasury)
	if balance.Cmp(missed) < 0 {
		return scoreresult.Errorf(hvhmodule.StatusCriticalError,
			"Invalid PublicTreasury balance=%d missed=%d",
			balance, missed)
	}
	if err = cc.Transfer(hvhmodule.PublicTreasury, hvhmodule.SustainableFund, missed, module.Transfer); err != nil {
		return err
	}
	if err = increaseVarDBInSustainableFund(cc, hvhmodule.VarMissingReward, missed); err != nil {
		return err
	}

	es.Logger().Debugf(
		"TransferMissedReward() end: missed=%d publicTreasury=%d",
		missed, cc.GetBalance(hvhmodule.PublicTreasury))
	return nil
}

func (es *ExtensionStateImpl) distributeFees(cc hvhmodule.CallContext) error {
	var err error

	// TxFee Distribution
	if err = es.distributeTxFee(cc, hvhmodule.BigRatEcoSystemToFee); err != nil {
		return err
	}
	// ServiceFee Distribution
	if err = es.distributeServiceFee(cc, hvhmodule.BigRatEcoSystemToFee); err != nil {
		return err
	}

	return nil
}

func (es *ExtensionStateImpl) refillHooverFund(cc hvhmodule.CallContext) error {
	es.Logger().Debugf("refillHooverFund() start")

	sf := hvhmodule.SustainableFund
	hf := hvhmodule.HooverFund
	hooverBudget := es.state.GetHooverBudget() // unit: hvh

	// amount = original HooverFund Budget - hfBalance
	hfBalance := cc.GetBalance(hf)
	amount := new(big.Int).Sub(hooverBudget, hfBalance)

	if amount.Sign() > 0 {
		sfBalance := cc.GetBalance(sf)
		if sfBalance.Cmp(amount) < 0 {
			amount.Set(sfBalance)
		}
		if err := cc.Transfer(sf, hf, amount, module.Transfer); err != nil {
			return err
		}
		if err := increaseVarDBInSustainableFund(cc, hvhmodule.VarHooverRefill, amount); err != nil {
			return err
		}

		hfBalance = cc.GetBalance(hf)
		sfBalance = cc.GetBalance(sf)
		es.Logger().Debugf(
			"onHooverRefilledEvent(): amount=%d hfBalance=%d sfBalance=%d",
			amount, hfBalance, sfBalance)
		onHooverRefilledEvent(cc, amount, hfBalance, sfBalance)
	}

	es.Logger().Debugf("refillHooverFund() end")
	return nil
}

func (es *ExtensionStateImpl) onTermStart(cc hvhmodule.CallContext, termSeq int64, baseTx *baseDataJSON) error {
	var err error
	es.Logger().Debugf("onTermStart() start: termSeq=%d", termSeq)

	issueAmount, update := es.state.GetIssueAmountByTS(termSeq)
	if update {
		es.Logger().Infof("IssueAmount is reduced to=%d", issueAmount)
		if err = es.state.SetIssueAmount(issueAmount); err != nil {
			return err
		}
	}

	// verify baseTx
	if baseTx.IssueAmount.Value().Cmp(issueAmount) != 0 {
		return transaction.InvalidTxValue.Errorf("Invalid issueAmount exp=%d real=%d",
			issueAmount, baseTx.IssueAmount.Value())
	}

	if err = es.issueCoin(cc, termSeq, issueAmount); err != nil {
		return err
	}

	// Reset reward-related states
	if err = es.state.OnTermStart(issueAmount); err != nil {
		return err
	}

	es.Logger().Debugf("onTermStart() end: termSeq=%d", termSeq)
	return nil
}

func (es *ExtensionStateImpl) issueCoin(cc hvhmodule.CallContext, termSeq int64, amount *big.Int) error {
	es.Logger().Debugf("issueCoin() start: termSeq=%d amount=%d", termSeq, amount)

	totalSupply, err := cc.Issue(hvhmodule.PublicTreasury, amount)
	if err != nil {
		return err
	}
	onIssuedEvent(cc, termSeq, amount, totalSupply)

	es.Logger().Debugf("issueCoin() end: termSeq=%d", termSeq)
	return nil
}

func (es *ExtensionStateImpl) distributeServiceFee(cc hvhmodule.CallContext, proportion *big.Rat) error {
	es.Logger().Debugf("distributeServiceFee() start: proportion: %#v", proportion)

	from := hvhmodule.ServiceTreasury
	_, sfAmount, err := es.distributeFee(cc, from, proportion)
	if err != nil {
		return err
	}
	if err = increaseVarDBInSustainableFund(cc, hvhmodule.VarServiceFee, sfAmount); err != nil {
		return err
	}

	es.Logger().Debugf("distributeServiceFee() end: from=%s sfAmount=%d", from, sfAmount)
	return nil
}

func (es *ExtensionStateImpl) distributeTxFee(cc hvhmodule.CallContext, proportion *big.Rat) error {
	es.Logger().Debugf("distributeTxFee() start: proportion=%#v", proportion)

	from := cc.Treasury()
	_, sfAmount, err := es.distributeFee(cc, from, proportion)
	if err != nil {
		return err
	}
	if err = increaseVarDBInSustainableFund(cc, hvhmodule.VarTxFee, sfAmount); err != nil {
		return err
	}

	es.Logger().Debugf("distributeTxFee() end: from=%s sfAmount=%d", from, sfAmount)
	return nil
}

func (es *ExtensionStateImpl) distributeFee(
	cc hvhmodule.CallContext, from module.Address, proportion *big.Rat,
) (*big.Int, *big.Int, error) {
	es.Logger().Debugf("distributeFee() start: from=%s proportion=%#v", from, proportion)

	var err error
	balance := cc.GetBalance(from)
	ecoAmount := hvhmodule.BigIntZero
	susAmount := hvhmodule.BigIntZero

	if balance.Sign() > 0 {
		ecoAmount = new(big.Int).Mul(balance, proportion.Num())
		ecoAmount.Div(ecoAmount, proportion.Denom())
		susAmount = new(big.Int).Sub(balance, ecoAmount)

		if err = cc.Transfer(from, hvhmodule.SustainableFund, susAmount, module.Transfer); err != nil {
			return nil, nil, err
		}
		if err = cc.Transfer(from, hvhmodule.EcoSystem, ecoAmount, module.Transfer); err != nil {
			return nil, nil, err
		}
	}

	es.Logger().Debugf("distributeFee() end: from=%s ecoAmount=%d susAmount=%d", from, ecoAmount, susAmount)
	return ecoAmount, susAmount, nil
}

func increaseVarDBInSustainableFund(cc hvhmodule.CallContext, key string, amount *big.Int) error {
	return increaseScoreVarDB(cc, hvhmodule.SustainableFund, key, amount)
}

func increaseScoreVarDB(cc hvhmodule.CallContext, score module.Address, key string, amount *big.Int) error {
	if amount == nil || amount.Sign() <= 0 {
		return nil
	}
	as := cc.GetAccountState(score.ID())
	varDB := scoredb.NewVarDB(as, key)
	value := varDB.BigInt()
	if value == nil {
		value = amount
	} else {
		value = new(big.Int).Add(value, amount)
	}
	return varDB.Set(value)
}

func CheckBaseTX(tx module.Transaction) bool {
	_, ok := transaction.Unwrap(tx).(*baseV3)
	return ok
}
