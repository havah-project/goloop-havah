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
	if logger != nil {
		es.logger = logger
	}
}

func (es *ExtensionStateImpl) State() *hvhstate.State {
	return es.state
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

func (es *ExtensionStateImpl) InitPlatformConfig(cfg *PlatformConfig) error {
	var err error

	if cfg.TermPeriod != nil {
		if err = es.state.SetInt64(hvhmodule.VarTermPeriod, cfg.TermPeriod.Value); err != nil {
			return err
		}
	}
	if cfg.IssueReductionCycle != nil {
		if err = es.state.SetInt64(hvhmodule.VarIssueReductionCycle, cfg.IssueReductionCycle.Value); err != nil {
			return err
		}
	}
	if cfg.PrivateReleaseCycle != nil {
		if err = es.state.SetInt64(hvhmodule.VarPrivateReleaseCycle, cfg.PrivateReleaseCycle.Value); err != nil {
			return err
		}
	}
	if cfg.PrivateLockup != nil {
		if err = es.state.SetInt64(hvhmodule.VarPrivateLockup, cfg.PrivateLockup.Value); err != nil {
			return err
		}
	}

	if cfg.IssueAmount != nil {
		if err = es.state.SetBigInt(hvhmodule.VarIssueAmount, cfg.IssueAmount.Value()); err != nil {
			return err
		}
	}
	if cfg.HooverBudget != nil {
		if err = es.state.SetBigInt(hvhmodule.VarHooverBudget, cfg.HooverBudget.Value()); err != nil {
			return err
		}
	}
	if cfg.USDTPrice != nil {
		if err = es.state.SetBigInt(hvhmodule.VarUSDTPrice, cfg.USDTPrice.Value()); err != nil {
			return err
		}
	} else {
		return scoreresult.InvalidParameterError.New("USDTPrice not found")
	}

	return nil
}

func (es *ExtensionStateImpl) IsIssueStarted(height int64) bool {
	issueStart := es.state.GetIssueStart()
	return issueStart > 0 && height >= issueStart
}

// NewBaseTransactionData creates data part of a baseTransaction
func (es *ExtensionStateImpl) NewBaseTransactionData() map[string]interface{} {
	return nil
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
		"height": height,
	}
	if issueStart > 0 {
		jso["termSequence"] = (height - issueStart) / termPeriod
	}
	return jso, nil
}

func (es *ExtensionStateImpl) StartRewardIssue(cc hvhmodule.CallContext, startBH int64) error {
	return es.state.SetIssueStart(cc.BlockHeight(), startBH)
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
	return es.state.RegisterPlanet(id, isPrivate, isCompany, owner, usdt, price, cc.BlockHeight())
}

func (es *ExtensionStateImpl) UnregisterPlanet(cc hvhmodule.CallContext, id int64) error {
	return es.state.UnregisterPlanet(id)
}

func (es *ExtensionStateImpl) SetPlanetOwner(cc hvhmodule.CallContext, id int64, owner module.Address) error {
	return es.state.SetPlanetOwner(id, owner)
}

func (es *ExtensionStateImpl) GetPlanetInfo(cc hvhmodule.CallContext, id int64) (map[string]interface{}, error) {
	p, err := es.state.GetPlanet(id)
	if err != nil {
		return nil, err
	}
	return p.ToJSON(), nil
}

func (es *ExtensionStateImpl) ReportPlanetWork(cc hvhmodule.CallContext, id int64) error {
	from := cc.From()
	if ok, err := es.state.IsPlanetManager(from); err != nil {
		return err
	} else {
		if !ok {
			return scoreresult.AccessDeniedError.Errorf(
				"Only planetManager can report planet work: %s", from)
		}
	}

	// Check if a planet exists
	p, err := es.state.GetPlanet(id)
	if err != nil {
		return err
	}

	height := cc.BlockHeight()
	issueStart := es.state.GetIssueStart()
	termPeriod := es.state.GetTermPeriod()
	termSequence := (height - issueStart) / termPeriod
	termStart := termSequence*termPeriod + issueStart

	if p.Height() >= termStart {
		// If a planet is registered in this term, ignore its work report
		return nil
	}

	reward := new(big.Int).Div(
		es.state.GetBigInt(hvhmodule.VarRewardTotal),
		es.state.GetBigInt(hvhmodule.VarActivePlanet))
	rewardWithHoover := reward

	if err = es.state.DecreaseRewardRemain(reward); err != nil {
		return err
	}

	// All planets have their own planetReward info
	pr, err := es.state.GetPlanetReward(id)
	if err != nil {
		return err
	}

	// hooverLimit = planetReward.total + reward - planet.price
	hooverLimit := pr.Total()
	hooverLimit.Add(hooverLimit, reward)
	hooverLimit.Sub(hooverLimit, p.Price())
	hooverRequest := new(big.Int)

	if hooverLimit.Sign() > 0 {
		hooverGuide := es.calcHooverGuide(p)

		// if reward < hooverGuide
		if reward.Cmp(hooverGuide) < 0 {
			hooverRequest.Sub(hooverGuide, reward)
			// if hooverRequest > hooverLimit
			if hooverRequest.Cmp(hooverLimit) > 0 {
				hooverRequest = hooverLimit
			}
			// HooverFund provides the subsidy for reward to PublicTreasury
			if hooverRequest, err = es.transferSubsidyFromHooverFund(cc, hooverRequest); err != nil {
				return err
			}
			rewardWithHoover = new(big.Int).Add(reward, hooverRequest)
		}
	}
	if err = es.state.OfferReward(termSequence+1, id, rewardWithHoover); err != nil {
		return err
	}
	onRewardOfferedEvent(cc, termSequence, id, rewardWithHoover, hooverRequest)
	return nil
}

func (es *ExtensionStateImpl) calcHooverGuide(p *hvhstate.Planet) *big.Int {
	hooverGuide := new(big.Int).Mul(p.USDT(), es.state.GetBigInt(hvhmodule.VarActiveUSDTPrice))
	hooverGuide.Div(hooverGuide, hvhmodule.BigIntUSDTDecimal)
	hooverGuide.Div(hooverGuide, big.NewInt(10))
	hooverGuide.Div(hooverGuide, es.state.GetBigInt(hvhmodule.VarIssueReductionCycle))
	return hooverGuide
}

// transferSubsidyFromHooverFund() transfers coins from HooverFund to PublicTreasury to support a planet reward
func (es *ExtensionStateImpl) transferSubsidyFromHooverFund(
	cc hvhmodule.CallContext, amount *big.Int) (*big.Int, error) {
	if amount.Sign() > 0 {
		balance := cc.GetBalance(hvhmodule.HooverFund)
		if balance.Cmp(amount) < 0 {
			// In the case where there is not enough balance in HooverFund to pay subsidy
			amount = balance
		}

		// HooverFund provides the subsidy for reward to PublicTreasury
		if err := cc.Transfer(hvhmodule.HooverFund, hvhmodule.PublicTreasury, amount); err != nil {
			return nil, err
		}
	}
	return amount, nil
}

// ClaimPlanetReward is used by a planet owner
// who wants to transfer a reward from system treasury to owner account
func (es *ExtensionStateImpl) ClaimPlanetReward(cc hvhmodule.CallContext, ids []int64) error {
	if len(ids) > hvhmodule.MaxCountToClaim {
		return scoreresult.Errorf(
			hvhmodule.StatusIllegalArgument,
			"Too many ids to claim: %d > max(%d)", len(ids), hvhmodule.MaxCountToClaim)
	}

	owner := cc.From()
	height := cc.BlockHeight()
	for _, id := range ids {
		reward, err := es.state.ClaimPlanetReward(id, height, owner)
		if err != nil {
			es.Logger().Warnf("Failed to claim a reward for %d", id)
		}
		if reward != nil && reward.Sign() > 0 {
			if err = cc.Transfer(hvhmodule.PublicTreasury, owner, reward); err != nil {
				return nil
			}
			onRewardClaimedEvent(cc, id, owner, reward)
		}
	}
	return nil
}

func (es *ExtensionStateImpl) GetRewardInfo(cc hvhmodule.CallContext, id int64) (map[string]interface{}, error) {
	height := cc.BlockHeight()
	return es.state.GetRewardInfo(height, id)
}
