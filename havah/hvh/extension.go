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
	"fmt"
	"math/big"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/codec"
	"github.com/icon-project/goloop/common/crypto"
	"github.com/icon-project/goloop/common/db"
	"github.com/icon-project/goloop/common/errors"
	"github.com/icon-project/goloop/common/log"
	"github.com/icon-project/goloop/common/merkle"
	"github.com/icon-project/goloop/havah/hvh/hvhstate"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/havah/hvhutils"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/scoreresult"
	"github.com/icon-project/goloop/service/state"
)

type ValidatorDataType int

const (
	VDataTypeNone ValidatorDataType = 0
	VDataTypeInfo ValidatorDataType = 1
	VDataTypeStatus ValidatorDataType = 2
	VDataTypeAll = VDataTypeInfo | VDataTypeStatus
)

func StringToValidatorDataType(value string) ValidatorDataType {
	switch value {
	case "info":
		return VDataTypeInfo
	case "status":
		return VDataTypeStatus
	case "all":
		return VDataTypeAll
	default:
		return VDataTypeNone
	}
}

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
	var hash []byte
	if _, err := codec.BC.UnmarshalFromBytes(raw, &hash); err != nil {
		return nil
	}
	return &ExtensionSnapshotImpl{
		dbase: builder.Database(),
		state: hvhstate.NewSnapshotWithBuilder(builder, hash),
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
	es.logger = logger
	es.state.SetLogger(es.logger)
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
	es.state.ClearCache()
}

func (es *ExtensionStateImpl) InitPlatformConfig(cfg *PlatformConfig) error {
	if err := es.state.InitState(&cfg.StateConfig); err != nil {
		return err
	}
	return nil
}

func (es *ExtensionStateImpl) GetIssueStart() int64 {
	return es.state.GetIssueStart()
}

func (es *ExtensionStateImpl) GetTermPeriod() int64 {
	return es.state.GetTermPeriod()
}

// NewBaseTransactionData creates data part of a baseTransaction
func (es *ExtensionStateImpl) NewBaseTransactionData(height int64) map[string]interface{} {
	issueStart := es.GetIssueStart()
	es.Logger().Debugf("NewBaseTransactionData() start: height=%d istart=%d", height, issueStart)

	issueAmount := es.state.GetIssueAmount(height, issueStart)
	if issueAmount == nil || issueAmount.Sign() <= 0 {
		termPeriod := es.GetTermPeriod()
		_, blockIndex := hvhstate.GetTermSequenceAndBlockIndex(height, issueStart, termPeriod)
		if !es.IsItTimeToCheckBlockVote(blockIndex) {
			return nil
		}

		if issueAmount == nil {
			issueAmount = hvhmodule.BigIntZero
		} else if issueAmount.Sign() < 0 {
			es.logger.Panicf("Invalid issueAmount: %d", issueAmount)
		}
	}

	jso := map[string]interface{}{
		"issueAmount": new(common.HexInt).SetValue(issueAmount),
	}

	es.Logger().Debugf("NewBaseTransactionData() end: issue=%s", issueAmount)
	return jso
}

func (es *ExtensionStateImpl) GetUSDTPrice() (*big.Int, error) {
	price := es.state.GetUSDTPrice()
	if price == nil || price.Sign() < 0 {
		return nil, scoreresult.RevertedError.New("Invalid USDTPrice")
	}
	return price, nil
}

func (es *ExtensionStateImpl) SetUSDTPrice(price *big.Int) error {
	return es.state.SetUSDTPrice(price)
}

func (es *ExtensionStateImpl) GetIssueInfo(cc hvhmodule.CallContext) (map[string]interface{}, error) {
	height := cc.BlockHeight()
	issueStart := es.state.GetIssueStart() // in height
	termPeriod := es.state.GetTermPeriod() // in height

	jso := map[string]interface{}{
		"height":              height,
		"termPeriod":          termPeriod,
		"issueReductionCycle": es.state.GetIssueReductionCycle(),
	}
	if issueStart > 0 {
		jso["issueStart"] = issueStart
	}
	if hvhstate.IsIssueStarted(height, issueStart) {
		jso["termSequence"] = (height - issueStart) / termPeriod
	}
	return jso, nil
}

func (es *ExtensionStateImpl) StartRewardIssue(cc hvhmodule.CallContext, startBH int64) error {
	height := cc.BlockHeight()
	es.Logger().Debugf("StartRewardIssue() start: height=%d startBH=%d", height, startBH)
	err := es.state.SetIssueStart(height, startBH)
	es.Logger().Debugf("StartRewardIssue() end")
	return err
}

func (es *ExtensionStateImpl) AddPlanetManager(address module.Address) error {
	return es.state.AddPlanetManager(address)
}

func (es *ExtensionStateImpl) RemovePlanetManager(address module.Address) error {
	return es.state.RemovePlanetManager(address)
}

func (es *ExtensionStateImpl) IsPlanetManager(address module.Address) (bool, error) {
	return es.state.IsPlanetManager(address)
}

func (es *ExtensionStateImpl) RegisterPlanet(
	cc hvhmodule.CallContext,
	id int64,
	isPrivate bool, isCompany bool,
	owner module.Address,
	usdt *big.Int, price *big.Int,
) error {
	height := cc.BlockHeight()
	return es.state.RegisterPlanet(
		cc.Revision().Value(),
		id, isPrivate, isCompany, owner, usdt, price, height)
}

func (es *ExtensionStateImpl) UnregisterPlanet(cc hvhmodule.CallContext, id int64) error {
	height := cc.BlockHeight()
	rev := cc.Revision().Value()
	es.Logger().Debugf("UnregisterPlanet() start: id=%d height=%d rev=%d", id, height, rev)

	lostDelta, err := es.state.UnregisterPlanet(rev, id)
	if rev >= hvhmodule.RevisionLostCoin {
		if err == nil && lostDelta != nil && lostDelta.Sign() > 0 {
			lostTotal, _ := es.state.GetLost()
			reason := fmt.Sprintf("PlanetUnregistered(id=%d)", id)
			onLostDepositedEvent(cc, lostDelta, lostTotal, reason)
			es.Logger().Debugf("LostDeposited(lostDelta=%d,lostTotal=%d,reason=%s)", lostDelta, lostTotal, reason)
		}
	}

	es.Logger().Debugf("UnregisterPlanet() end: id=%d height=%d rev=%d err=%#v", id, height, rev, err)
	return err
}

func (es *ExtensionStateImpl) SetPlanetOwner(cc hvhmodule.CallContext, id int64, owner module.Address) error {
	height := cc.BlockHeight()
	es.Logger().Debugf("SetPlanetOwner() start: height=%d id=%d owner=%s", height, id, owner)
	err := es.state.SetPlanetOwner(id, owner)
	es.Logger().Debugf("SetPlanetOwner() end: err=%#v", err)
	return err
}

func (es *ExtensionStateImpl) GetPlanetInfo(_ hvhmodule.CallContext, id int64) (map[string]interface{}, error) {
	p, err := es.state.GetPlanet(id)
	if err != nil {
		return nil, err
	}
	return p.ToJSON(), nil
}

func (es *ExtensionStateImpl) ReportPlanetWork(cc hvhmodule.CallContext, id int64) error {
	height := cc.BlockHeight()
	es.Logger().Debugf("ReportPlanetWork() start: height=%d id=%d", height, id)

	issueStart := es.state.GetIssueStart()
	if !hvhstate.IsIssueStarted(height, issueStart) {
		return errors.InvalidStateError.Errorf(
			"IssueDoesntStarted(height=%d,issueStart=%d)", height, issueStart)
	}

	// Check if a planet exists
	p, err := es.state.GetPlanet(id)
	if err != nil {
		return err
	}

	termPeriod := es.state.GetTermPeriod()
	termSeq := (height - issueStart) / termPeriod
	termStart := termSeq*termPeriod + issueStart
	termNumber := termSeq + 1

	es.Logger().Debugf(
		"planet=%#v height=%d istart=%d tp=%d tseq=%d tstart=%d",
		p, height, issueStart, termPeriod, termSeq, termStart)

	if p.Height() >= termStart {
		// If a planet is registered in this term, ignore its work report
		return scoreresult.Errorf(
			hvhmodule.StatusRewardError,
			"ReportPlanetWork not allowed during this term: tseq=%d id=%d", termSeq, id)
	}

	// All planets have their own planetReward info
	pr, err := es.state.GetPlanetReward(id)
	if err != nil {
		return err
	}
	if termNumber <= pr.LastTermNumber() {
		return scoreresult.Errorf(
			hvhmodule.StatusRewardError,
			"Duplicate reportPlanetWork: tn=%d id=%d", termNumber, id)
	}

	_, reward := es.state.GetActivePlanetCountAndReward()
	rewardWithHoover := reward

	if err = es.state.DecreaseRewardRemain(reward); err != nil {
		return err
	}

	// hooverLimit = planet.price - planetReward.total - reward
	hooverLimit := calcHooverLimit(pr.Total(), reward, p.Price())
	es.Logger().Debugf(
		"pr.total=%d reward=%d price=%d hooverLimit=%d",
		pr.Total(), reward, p.Price(), hooverLimit)

	hooverRequest := hvhmodule.BigIntZero
	if hooverLimit.Sign() > 0 {
		hooverGuide := calcHooverGuide(p.USDT(), es.state.GetActiveUSDTPrice())
		es.Logger().Debugf("hooverGuide=%d", hooverGuide)
		if reward.Cmp(hooverGuide) < 0 {
			hooverBalance := cc.GetBalance(hvhmodule.HooverFund)
			if hooverBalance.Sign() > 0 {
				hooverRequest = calcSubsidyFromHooverFund(hooverLimit, hooverGuide, hooverBalance, reward)
				rewardWithHoover = new(big.Int).Add(rewardWithHoover, hooverRequest)
			}
		}
	}

	if err = cc.Transfer(
		hvhmodule.HooverFund, hvhmodule.PublicTreasury, hooverRequest, module.Transfer); err != nil {
		return err
	}

	if p.IsCompany() {
		// Divide company rewards into ecoSystem and owner in a ratio of 6:4
		proportion := hvhmodule.BigRatEcoSystemToCompanyReward
		ecoReward := new(big.Int).Mul(rewardWithHoover, proportion.Num())
		ecoReward.Div(ecoReward, proportion.Denom())
		planetReward := new(big.Int).Sub(rewardWithHoover, ecoReward)

		if err = es.state.OfferReward(termNumber, id, pr, planetReward, rewardWithHoover); err != nil {
			return err
		}
		if err = es.state.IncreaseEcoSystemReward(ecoReward); err != nil {
			return err
		}
	} else {
		if err = es.state.OfferReward(
			termNumber, id, pr, rewardWithHoover, rewardWithHoover); err != nil {
			return err
		}
	}

	if err = es.state.IncrementWorkingPlanet(); err != nil {
		return err
	}
	onRewardOfferedEvent(cc, termSeq, id, rewardWithHoover, hooverRequest)
	es.Logger().Debugf(
		"ReportPlanetWork() end: height=%d id=%d rewardWithHoover=%d hooverRequest=%d",
		height, id, rewardWithHoover, hooverRequest)
	return nil
}

func calcHooverLimit(total, rewardPerPlanet, planetPrice *big.Int) *big.Int {
	// hooverLimit = planet.price - planetReward.total - reward)
	hooverLimit := new(big.Int).Sub(planetPrice, total)
	return hooverLimit.Sub(hooverLimit, rewardPerPlanet)
}

var DividerFor10Percent = big.NewInt(10)

// calcHooverGuide returns the rewards
// that a planet should receive every term to get 10% of its price in usdt as a reward
func calcHooverGuide(priceInUSDT, activeUSDTPrice *big.Int) *big.Int {
	hooverGuide := new(big.Int).Mul(priceInUSDT, activeUSDTPrice)
	hooverGuide.Div(hooverGuide, hvhmodule.BigIntUSDTDecimal)
	hooverGuide.Div(hooverGuide, DividerFor10Percent)
	hooverGuide.Div(hooverGuide, hvhmodule.BigIntDayPerYear)
	return hooverGuide
}

func calcSubsidyFromHooverFund(
	hooverLimit, hooverGuide, hooverBalance, reward *big.Int) *big.Int {
	hooverRequest := new(big.Int).Sub(hooverGuide, reward)
	// if hooverRequest > hooverLimit
	if hooverRequest.Cmp(hooverLimit) > 0 {
		hooverRequest = hooverLimit
	}
	// if hoooverRequest > hooverBalance
	if hooverRequest.Cmp(hooverBalance) > 0 {
		hooverRequest = hooverBalance
	}

	return hooverRequest
}

// ClaimPlanetReward is used by a planet owner
// who wants to transfer a reward from system treasury to owner account
func (es *ExtensionStateImpl) ClaimPlanetReward(cc hvhmodule.CallContext, ids []int64) error {
	height := cc.BlockHeight()
	es.Logger().Debugf("ClaimPlanetReward() start: height=%d ids=%v", height, ids)

	if len(ids) > hvhmodule.MaxCountToClaim {
		return scoreresult.Errorf(
			hvhmodule.StatusIllegalArgument,
			"Too many ids to claim: %d > max(%d)", len(ids), hvhmodule.MaxCountToClaim)
	}

	issueStart := es.state.GetIssueStart()
	if !hvhstate.IsIssueStarted(height, issueStart) {
		return errors.InvalidStateError.Errorf(
			"IssueDoesntStarted(height=%d,issueStart=%d)", height, issueStart)
	}
	termPeriod := es.state.GetTermPeriod()
	termSeq := (height - issueStart) / termPeriod

	owner := cc.From()
	for _, id := range ids {
		reward, err := es.state.ClaimPlanetReward(id, height, owner)
		if err != nil {
			es.Logger().Warnf("Failed to claim a reward for %d", id)
		}
		if reward != nil && reward.Sign() > 0 {
			if err = cc.Transfer(hvhmodule.PublicTreasury, owner, reward, module.Claim); err != nil {
				return nil
			}
			onRewardClaimedEvent(cc, owner, termSeq, id, reward)
			es.Logger().Debugf("owner=%s termSeq=%d id=%d reward=%d", owner, termSeq, id, reward)
		}
	}

	es.Logger().Debugf("ClaimPlanetReward() end")
	return nil
}

func (es *ExtensionStateImpl) GetRewardInfoOf(cc hvhmodule.CallContext, id int64) (map[string]interface{}, error) {
	height := cc.BlockHeight()
	jso, err := es.state.GetRewardInfoOf(height, id)
	if err == nil {
		jso["height"] = height
	}
	return jso, err
}

func (es *ExtensionStateImpl) GetRewardInfosOf(cc hvhmodule.CallContext, ids []int64) (map[string]interface{}, error) {
	if len(ids) > hvhmodule.MaxCountToClaim {
		return nil, scoreresult.Errorf(
			hvhmodule.StatusIllegalArgument,
			"Too many ids: ids(%d) > max(%d)", len(ids), hvhmodule.MaxCountToClaim)
	}

	height := cc.BlockHeight()
	jso := make(map[string]interface{})
	ris := make([]interface{}, len(ids))
	for i, id := range ids {
		if ri, err := es.state.GetRewardInfoOf(height, id); err == nil {
			ris[i] = ri
		} else {
			ris[i] = nil
		}
	}

	jso["height"] = height
	jso["rewardInfos"] = ris
	return jso, nil
}

func (es *ExtensionStateImpl) GetRewardInfo(cc hvhmodule.CallContext) (map[string]interface{}, error) {
	es.Logger().Debugf("GetRewardInfo() start")

	height := cc.BlockHeight()
	issueStart := es.state.GetIssueStart()
	if !hvhstate.IsIssueStarted(height, issueStart) {
		return nil, scoreresult.Errorf(hvhmodule.StatusNotReady, "TermNotReady")
	}

	_, reward := es.state.GetActivePlanetCountAndReward()
	termPeriod := es.state.GetTermPeriod()
	termSequence := (height - issueStart) / termPeriod
	es.Logger().Infof("GetRewardInfo: height=%d termPeriod=%d termSequence=%d reward=%d",
		height, termPeriod, termSequence, reward)

	es.Logger().Debugf("GetRewardInfo() end")
	return map[string]interface{}{
		"height":                height,
		"termSequence":          termSequence,
		"rewardPerActivePlanet": reward,
	}, nil
}

func (es *ExtensionStateImpl) BurnCoin(cc hvhmodule.CallContext, amount *big.Int) error {
	from := cc.From()
	es.Logger().Debugf("BurnCoin() start: from=%s amount=%d", from, amount)
	totalSupply, err := cc.Burn(amount)
	if err != nil {
		return err
	}
	onBurnedEvent(cc, state.SystemAddress, amount, totalSupply)
	es.Logger().Debugf(
		"BurnCoin() end: from=%s amount=%d ts=%d", from, amount, totalSupply)
	return nil
}

func (es *ExtensionStateImpl) SetPrivateClaimableRate(numerator, denominator int64) error {
	es.Logger().Debugf(
		"SetPrivateClaimableRate() start: numerator=%d denominator=%d",
		numerator, denominator)
	defer es.Logger().Debugf("SetPrivateClaimableRate() end")
	return es.state.SetPrivateClaimableRate(numerator, denominator)
}

func (es *ExtensionStateImpl) GetPrivateClaimableRate() (map[string]interface{}, error) {
	num, denom := es.state.GetPrivateClaimableRate()
	return map[string]interface{}{
		"numerator":   num,
		"denominator": denom,
	}, nil
}

func (es *ExtensionStateImpl) WithdrawLostTo(cc hvhmodule.CallContext, to module.Address) error {
	es.Logger().Debugf("WithdrawLostTo() start: to=%s height=%d", to, cc.BlockHeight())
	defer es.Logger().Debugf("WithdrawLostTo() end")

	var err error
	var lost *big.Int

	if !to.IsContract() {
		if lost, err = es.state.DeleteLost(); err == nil {
			if lost.Sign() > 0 {
				if err = cc.Transfer(hvhmodule.PublicTreasury, to, lost, module.Transfer); err == nil {
					onLostWithdrawnEvent(cc, to, lost)
				}
			}
		}
	} else {
		err = scoreresult.Errorf(
			hvhmodule.StatusIllegalArgument, "ContractNotAllowed(to=%s)", to)
	}

	if err != nil {
		es.Logger().Infof("WithdrawLostTo() is failed: err=%v", err)
	}
	return err
}

func (es *ExtensionStateImpl) GetLost() (*big.Int, error) {
	return es.state.GetLost()
}

func (es *ExtensionStateImpl) SetBlockVoteCheckParameters(cc hvhmodule.CallContext, period, allowance int64) error {
	from := cc.From()
	height := cc.BlockHeight()
	es.logger.Debugf("SetBlockVoteCheckParameters() start: height=%d from=%s", height, from)
	err := es.state.SetBlockVoteCheckParameters(period, allowance)
	es.logger.Debugf("SetBlockVoteCheckParameters() end: height=%d", height)
	return err
}

func (es *ExtensionStateImpl) GetBlockVoteCheckParameters(cc hvhmodule.CallContext) (map[string]interface{}, error) {
	return map[string]interface{}{
		"height":               cc.BlockHeight(),
		"blockVoteCheckPeriod": es.state.GetBlockVoteCheckPeriod(),
		"nonVoteAllowance":     es.state.GetNonVoteAllowance(),
	}, nil
}

func (es *ExtensionStateImpl) RegisterValidator(
	cc hvhmodule.CallContext,
	owner module.Address, nodePublicKey []byte, gradeName, name string,
	urlPtr *string) error {
	height := cc.BlockHeight()
	es.logger.Debugf(
		"RegisterValidator() start: height=%d owner=%s nodePublicKey=%x grade=%s name=%s urlPtr=%v",
		height, owner, nodePublicKey, gradeName, name, urlPtr)

	if err := hvhutils.CheckCompressedPublicKeyFormat(nodePublicKey); err != nil {
		return err
	}
	grade := hvhstate.StringToGrade(gradeName)
	if grade == hvhstate.GradeNone {
		return scoreresult.InvalidParameterError.Errorf("InvalidArgument(grade=%s)", gradeName)
	}
	if err := hvhutils.CheckNameLength(name); err != nil {
		return err
	}
	if urlPtr != nil {
		if err := hvhutils.CheckUrlLength(*urlPtr); err != nil {
			return err
		}
	}
	err := es.state.RegisterValidator(owner, nodePublicKey, grade, name, urlPtr)

	es.logger.Debugf("RegisterValidator() end: height=%d err=%v", height, err)
	return err
}

func (es *ExtensionStateImpl) UnregisterValidator(cc hvhmodule.CallContext, owner module.Address) error {
	from := cc.From()
	height := cc.BlockHeight()
	es.logger.Debugf("UnregisterValidator() start: height=%d from=%s owner=%s", height, from, owner)

	if from == nil {
		return scoreresult.InvalidParameterError.Errorf("InvalidArgument(from=%s)", from)
	}
	if err := hvhutils.CheckAddressArgument(owner, true); err != nil {
		return err
	}
	isCallerGov := from.Equal(cc.Governance())
	if !isCallerGov && !from.Equal(owner) {
		return scoreresult.AccessDeniedError.Errorf("NoPermission(from=%s,owner=%s", from, owner)
	}

	node, err := es.state.UnregisterValidator(owner)
	if err != nil {
		return err
	}
	if node != nil {
		// Change active validator set if it is an active validator
		var changed bool
		validatorState := cc.GetValidatorState()
		changed, err = replaceActiveValidatorAddress(validatorState, node, nil)
		if err != nil {
			return err
		}
		if changed {
			onActiveValidatorRemoved(cc, owner, node, "unregistered")
			nc := int64(validatorState.Len())
			onActiveValidatorCountChanged(cc, nc+1, nc)
		}
	}

	es.logger.Debugf("UnregisterValidator() end: height=%d", height)
	return nil
}

func (es *ExtensionStateImpl) GetNetworkStatus(cc hvhmodule.CallContext) (map[string]interface{}, error) {
	height := cc.BlockHeight()
	es.logger.Debugf("GetNetworkStatus() start: height=%d", height)

	issueStart := es.state.GetIssueStart()
	if issueStart == 0 {
		return nil, scoreresult.Errorf(
			hvhmodule.StatusNotReady, "TermNotReady(issueStart=0)")
	}
	ns, err := es.state.GetNetworkStatus()
	if err != nil {
		return nil, err
	}
	if !ns.IsDecentralized() {
		return nil, scoreresult.Errorf(
			hvhmodule.StatusNotReady, "NotYetDecentralized")
	}

	var termStart int64
	var jso map[string]interface{}
	termPeriod := es.state.GetTermPeriod()
	termStart, _, err = hvhstate.GetTermStartAndIndex(height, issueStart, termPeriod)
	if err != nil {
		return nil, err
	}
	jso = ns.ToJSON()
	jso["height"] = height
	jso["termStart"] = termStart

	es.logger.Debugf("GetNetworkStatus() end: height=%d", height)
	return jso, nil
}

func (es *ExtensionStateImpl) SetValidatorInfo(cc hvhmodule.CallContext, m map[string]string) error {
	var err error
	from := cc.From()
	height := cc.BlockHeight()
	es.logger.Debugf("SetValidatorInfo() start: height=%d from=%s", height, from)

	if err = hvhutils.CheckAddressArgument(from, true); err != nil {
		return err
	}

	if len(m) == 0 {
		es.logger.Infof("Empty argument")
		return nil
	}
	err = es.state.SetValidatorInfo(from, m)

	es.logger.Debugf("SetValidatorInfo() end: height=%d err=%v", height, err)
	return err
}

func (es *ExtensionStateImpl) SetNodePublicKey(cc hvhmodule.CallContext, pubKey []byte) error {
	from := cc.From()
	height := cc.BlockHeight()
	es.logger.Debugf("SetNodePublicKey() start: height=%d from=%s pubKey=%x", height, from, pubKey)

	if err :=  hvhutils.CheckAddressArgument(from, true); err != nil {
		return err
	}
	if err := hvhutils.CheckCompressedPublicKeyFormat(pubKey); err != nil {
		return err
	}
	oldNode, newNode, err := es.state.SetNodePublicKey(from, pubKey)
	if err != nil {
		return err
	}

	var changed bool
	validatorState := cc.GetValidatorState()
	if changed, err = replaceActiveValidatorAddress(validatorState, oldNode, newNode); err == nil {
		if changed {
			onActiveValidatorRemoved(cc, from, oldNode, "pubkeychange")
			onActiveValidatorAdded(cc, from, newNode)
		}
	}

	es.logger.Debugf("SetNodePublicKey() end: height=%d err=%v", height, err)
	return err
}

func replaceActiveValidatorAddress(
	validatorState state.ValidatorState, oldNode, newNode module.Address) (bool, error) {
	if oldNode.Equal(newNode) {
		// No need to replace, because old node is the same as new one
		return false, nil
	}
	idx := validatorState.IndexOf(oldNode)
	if idx < 0 {
		// oldNode is not an active validator
		return false, nil
	}

	if newNode != nil {
		v, err := state.ValidatorFromAddress(newNode)
		if err != nil {
			return false, err
		}
		return true, validatorState.SetAt(idx, v)
	} else {
		v, _ := validatorState.Get(idx)
		validatorState.Remove(v)
		return true, nil
	}
}

func (es *ExtensionStateImpl) EnableValidator(cc hvhmodule.CallContext, owner module.Address) error {
	from := cc.From()
	height := cc.BlockHeight()
	es.logger.Debugf("EnableValidator() start: height=%d from=%s", height, from)

	if from == nil {
		return scoreresult.InvalidParameterError.Errorf("InvalidArgument(from=%s)", from)
	}
	if err := hvhutils.CheckAddressArgument(owner, true); err != nil {
		return err
	}

	isCallerGov := from.Equal(cc.Governance())
	if !isCallerGov && !from.Equal(owner) {
		return scoreresult.AccessDeniedError.Errorf("NoPermission(from=%s,owner=%s)", from, owner)
	}

	err := es.state.EnableValidator(owner, isCallerGov)
	es.logger.Debugf("EnableValidator() end: height=%d err=%v", height, err)
	return err
}

func (es *ExtensionStateImpl) GetValidatorInfo(
	cc hvhmodule.CallContext, owner module.Address) (map[string]interface{}, error) {
	height := cc.BlockHeight()
	es.Logger().Debugf("GetValidatorInfo() start: height=%d owner=%s", height, owner)

	var jso map[string]interface{}
	vi, err := es.state.GetValidatorInfo(owner)
	if err == nil {
		jso = vi.ToJSON()
		jso["height"] = height
	}

	es.Logger().Debugf("GetValidatorInfo() end: owner=%s err=%v", owner, err)
	return jso, err
}

func (es *ExtensionStateImpl) GetValidatorStatus(
	cc hvhmodule.CallContext, owner module.Address) (map[string]interface{}, error) {
	height := cc.BlockHeight()
	es.Logger().Debugf("GetValidatorStatus() start: height=%d owner=%s", height, owner)

	var jso map[string]interface{}
	vs, err := es.state.GetValidatorStatus(owner)
	if err == nil {
		jso = vs.ToJSON()
		jso["height"] = height
	}

	es.Logger().Debugf("GetValidatorStatus() end: owner=%s err=%v", owner, err)
	return jso, err
}

func (es *ExtensionStateImpl) SetActiveValidatorCount(cc hvhmodule.CallContext, count int64) error {
	height := cc.BlockHeight()
	es.Logger().Debugf("SetActiveValidatorCount() start: height=%d count=%d", height, count)

	if !(count > 0 && count <= hvhmodule.MaxValidatorCount) {
		return scoreresult.InvalidParameterError.Errorf("InvalidArgument(%d)", count)
	}

	err := es.state.SetActiveValidatorCount(count)
	es.Logger().Debugf("SetActiveValidatorCount() end: height=%d err=%v", height, err)
	return err
}

func (es *ExtensionStateImpl) GetValidatorCount() (int64, error) {
	return es.state.GetActiveValidatorCount(), nil
}

func (es *ExtensionStateImpl) GetValidatorsOf(
	cc hvhmodule.CallContext, gradeFilterName string) (map[string]interface{}, error) {
	height := cc.BlockHeight()
	gradeFilter := hvhstate.StringToGradeFilter(gradeFilterName)
	if gradeFilter == hvhstate.GradeFilterNone {
		return nil, scoreresult.InvalidParameterError.Errorf("InvalidArgument(%s)", gradeFilterName)
	}

	if validators, err := es.state.GetValidatorsOf(gradeFilter); err == nil {
		addrs := make([]interface{}, len(validators))
		for i, v := range validators {
			addrs[i] = v
		}
		return map[string]interface{}{
			"height":     height,
			"grade":      gradeFilterName,
			"validators": addrs,
		}, nil
	} else {
		return nil, err
	}
}

func (es *ExtensionStateImpl) GetValidatorsInfo(
	cc hvhmodule.CallContext, dataType string, disqualifiedOnly bool) (map[string]interface{}, error) {
	var err error
	var owners []module.Address
	height := cc.BlockHeight()
	es.Logger().Debugf(
		"GetValidatorsInfo() start: height=%d dataType=%s disqualifiedOnly=%t",
		height, dataType, disqualifiedOnly)

	vDataType := StringToValidatorDataType(dataType)
	if vDataType == VDataTypeNone {
		return nil, scoreresult.InvalidParameterError.Errorf("InvalidArgument(dataType=%s)", dataType)
	}

	if disqualifiedOnly {
		owners, err = es.state.GetDisqualifiedValidators()
	} else {
		owners, err = es.state.GetValidatorsOf(hvhstate.GradeFilterAll)
	}
	if err != nil {
		return nil, err
	}

	var jso map[string]interface{}
	validators := make([]interface{}, len(owners))
	ret := make(map[string]interface{})

	for i, owner := range owners {
		if jso, err = es.getValidatorInfoAndStatus(owner, vDataType); err == nil {
			validators[i] = jso
		} else {
			return nil, err
		}
	}

	ret["height"] = height
	ret["validators"] = validators
	es.Logger().Debugf("GetValidatorsInfo() end: height=%d", height)
	return ret, nil
}

func (es *ExtensionStateImpl) getValidatorInfoAndStatus(
	owner module.Address, vDataType ValidatorDataType) (map[string]interface{}, error) {
	var err error
	var vi *hvhstate.ValidatorInfo
	var vs *hvhstate.ValidatorStatus
	var jso map[string]interface{}

	if vDataType & VDataTypeInfo != 0 {
		vi, err = es.state.GetValidatorInfo(owner)
		if err != nil {
			return nil, err
		}
		jso = vi.ToJSON()
	} else {
		jso = nil
	}

	if vDataType & VDataTypeStatus != 0 {
		vs, err = es.state.GetValidatorStatus(owner)
		if err != nil {
			return nil, err
		}
		vsJso := vs.ToJSON()
		if jso == nil {
			jso = vsJso
			jso["owner"] = owner
		} else {
			for k, v := range vsJso {
				jso[k] = v
			}
		}
	}

	return jso, nil
}

// InitBTPPublicKeys registers existing validators public keys to BTPState
// Called only once when the revision is set to RevisionBTP2
func (es *ExtensionStateImpl) InitBTPPublicKeys(btpCtx state.BTPContext, bsi *state.BTPStateImpl) error {
	height := btpCtx.BlockHeight()
	es.logger.Debugf("InitBTPPublicKeys() start: height=%s", height)

	var vi *hvhstate.ValidatorInfo
	var publicKey *crypto.PublicKey
	owners, err := es.state.GetValidatorsOf(hvhstate.GradeFilterAll)
	if err != nil {
		return err
	}

	for _, owner := range owners {
		vi, err = es.state.GetValidatorInfo(owner)
		if err != nil {
			return err
		}
		publicKey = vi.PublicKey()
		if err = bsi.SetPublicKey(
			btpCtx, owner, hvhmodule.DSASecp256k1,
			publicKey.SerializeCompressed()); err != nil {
			return err
		}
	}

	es.logger.Debugf("InitBTPPublicKeys() end: height=%s", height)
	return nil
}

func GetExtensionStateFromWorldContext(wc state.WorldContext, logger log.Logger) *ExtensionStateImpl {
	es := wc.GetExtensionState()
	if es == nil {
		return nil
	}
	if esi, ok := es.(*ExtensionStateImpl); ok {
		esi.SetLogger(hvhutils.NewLogger(logger))
		return esi
	} else {
		return nil
	}
}
