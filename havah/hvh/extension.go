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

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/codec"
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
	// es.state.ClearCache()
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
func (es *ExtensionStateImpl) NewBaseTransactionData(height, issueStart int64) map[string]interface{} {
	es.Logger().Debugf("NewBaseTransactionData() start: height=%d istart=%d", height, issueStart)

	issueAmount := es.state.GetIssueAmount(height, issueStart)
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
	return es.state.RegisterPlanet(id, isPrivate, isCompany, owner, usdt, price, height)
}

func (es *ExtensionStateImpl) UnregisterPlanet(cc hvhmodule.CallContext, id int64) error {
	es.Logger().Debugf("UnregisterPlanet() start: height=%d id=%d", id, cc.BlockHeight())
	err := es.state.UnregisterPlanet(id)
	es.Logger().Debugf("UnregisterPlanet() end: err=%#v", err)
	return err
}

func (es *ExtensionStateImpl) SetPlanetOwner(cc hvhmodule.CallContext, id int64, owner module.Address) error {
	height := cc.BlockHeight()
	es.Logger().Debugf("SetPlanetOwner() start: height=%d id=%d owner=%s", height, id, owner)
	err := es.state.SetPlanetOwner(id, owner)
	es.Logger().Debugf("SetPlanetOwner() end: err=%#v", err)
	return err
}

func (es *ExtensionStateImpl) GetPlanetInfo(cc hvhmodule.CallContext, id int64) (map[string]interface{}, error) {
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
		return nil
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

	reward := es.state.GetRewardPerActivePlanet()
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
		hooverGuide := es.calcHooverGuide(p)
		es.Logger().Debugf("hooverGuide=%d", hooverGuide)
		if reward.Cmp(hooverGuide) < 0 {
			hooverBalance := cc.GetBalance(hvhmodule.HooverFund)
			if hooverBalance.Sign() > 0 {
				hooverRequest = es.calcSubsidyFromHooverFund(hooverLimit, hooverGuide, hooverBalance, reward)
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
func (es *ExtensionStateImpl) calcHooverGuide(p *hvhstate.Planet) *big.Int {
	hooverGuide := new(big.Int).Mul(p.USDT(), es.state.GetActiveUSDTPrice())
	hooverGuide.Div(hooverGuide, hvhmodule.BigIntUSDTDecimal)
	hooverGuide.Div(hooverGuide, DividerFor10Percent)
	hooverGuide.Div(hooverGuide, hvhmodule.BigIntDayPerYear)
	return hooverGuide
}

func (es *ExtensionStateImpl) calcSubsidyFromHooverFund(
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
	return es.state.GetRewardInfoOf(height, id)
}

func (es *ExtensionStateImpl) GetRewardInfo(cc hvhmodule.CallContext) (map[string]interface{}, error) {
	es.Logger().Debugf("GetRewardInfo() start")

	height := cc.BlockHeight()
	issueStart := es.state.GetIssueStart()
	if !hvhstate.IsIssueStarted(height, issueStart) {
		return nil, scoreresult.Errorf(hvhmodule.StatusNotReady, "TermNotReady")
	}

	reward := es.state.GetRewardPerActivePlanet()
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
