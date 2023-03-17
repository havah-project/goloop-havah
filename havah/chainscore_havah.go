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

package havah

import (
	"encoding/json"
	"math/big"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/errors"
	"github.com/icon-project/goloop/havah/hvh"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/scoreresult"
)

func (s *chainScore) checkNFT(charge bool) error {
	if !hvhmodule.PlanetNFT.Equal(s.from) {
		if charge {
			if err := s.cc.ApplyCallSteps(); err != nil {
				return err
			}
		}
		return scoreresult.AccessDeniedError.Errorf("NoPermission(from=%s)", s.from)
	}
	return nil
}

func (s *chainScore) getExtensionState() (*hvh.ExtensionStateImpl, error) {
	es := hvh.GetExtensionStateFromWorldContext(s.cc, s.log)
	if es == nil {
		return nil, errors.InvalidStateError.New("ExtensionState is nil")
	}
	return es, nil
}

func (s *chainScore) getExtensionStateAndContext() (*hvh.ExtensionStateImpl, hvhmodule.CallContext, error) {
	s.log.Debug("getExtensionStateAndContext() start")

	es, err := s.getExtensionState()
	if err != nil {
		return nil, nil, err
	}

	ctx := hvh.NewCallContext(s.cc, s.from)
	if err = s.invokeBaseTxOnQueryMode(ctx, es); err != nil {
		return nil, nil, err
	}

	s.log.Debug("getExtensionStateAndContext() end")
	return es, ctx, nil
}

func (s *chainScore) invokeBaseTxOnQueryMode(ctx hvhmodule.CallContext, es *hvh.ExtensionStateImpl) error {
	if ctx.IsBaseTxInvoked() {
		return nil
	}

	height := ctx.BlockHeight()
	if baseData := es.NewBaseTransactionData(height); baseData != nil {
		if bs, err := json.Marshal(baseData); err == nil {
			if err = es.OnBaseTx(ctx, bs); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	ctx.SetBaseTxInvoked()
	return nil
}

func (s *chainScore) checkIntercallOnQueryMode() error {
	s.cc.Logger().Debugf("checkIntercallOnQueryMode(from=%s)", s.from)
	if s.cc.TransactionID() == nil && s.from != nil && s.from.IsContract() {
		return scoreresult.RevertedError.Errorf("IntercallNotAllowedOnQueryMode(from=%s)", s.from)
	}
	return nil
}

func (s *chainScore) Ex_getUSDTPrice() (*big.Int, error) {
	if s.cc.Revision().Value() >= hvhmodule.RevisionFixStepCharge {
		if err := s.tryChargeCall(); err != nil {
			return nil, err
		}
	}
	es, _, err := s.getExtensionStateAndContext()
	if err != nil {
		return nil, err
	}
	return es.GetUSDTPrice()
}

func (s *chainScore) Ex_setUSDTPrice(price *common.HexInt) error {
	if s.cc.Revision().Value() >= hvhmodule.RevisionFixAccessControl {
		if err := s.checkGovernance(true); err != nil {
			return err
		}
	}
	es, _, err := s.getExtensionStateAndContext()
	if err != nil {
		return err
	}
	return es.SetUSDTPrice(price.Value())
}

func (s *chainScore) Ex_getIssueInfo() (map[string]interface{}, error) {
	if s.cc.Revision().Value() >= hvhmodule.RevisionFixStepCharge {
		if err := s.tryChargeCall(); err != nil {
			return nil, err
		}
	}
	es, ctx, err := s.getExtensionStateAndContext()
	if err != nil {
		return nil, err
	}
	return es.GetIssueInfo(ctx)
}

func (s *chainScore) Ex_startRewardIssue(height *common.HexInt) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	es, ctx, err := s.getExtensionStateAndContext()
	if err != nil {
		return err
	}
	startBH := height.Int64()
	if startBH <= ctx.BlockHeight() {
		return scoreresult.RevertedError.New("Invalid height")
	}
	return es.StartRewardIssue(ctx, startBH)
}

func (s *chainScore) Ex_addPlanetManager(address module.Address) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	es, _, err := s.getExtensionStateAndContext()
	if err != nil {
		return err
	}
	return es.AddPlanetManager(address)
}

func (s *chainScore) Ex_removePlanetManager(address module.Address) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	es, _, err := s.getExtensionStateAndContext()
	if err != nil {
		return err
	}
	return es.RemovePlanetManager(address)
}

func (s *chainScore) Ex_isPlanetManager(address module.Address) (bool, error) {
	if s.cc.Revision().Value() >= hvhmodule.RevisionFixStepCharge {
		if err := s.tryChargeCall(); err != nil {
			return false, err
		}
	}
	es, _, err := s.getExtensionStateAndContext()
	if err != nil {
		return false, err
	}
	return es.IsPlanetManager(address)
}

func (s *chainScore) Ex_registerPlanet(
	id *common.HexInt,
	isPrivate, isCompany bool, owner module.Address, usdt, price *common.HexInt) error {
	if err := s.checkNFT(true); err != nil {
		return err
	}
	es, ctx, err := s.getExtensionStateAndContext()
	if err != nil {
		return err
	}
	return es.RegisterPlanet(
		ctx, id.Int64(),
		isPrivate, isCompany, owner, usdt.Value(), price.Value())
}

func (s *chainScore) Ex_unregisterPlanet(id *common.HexInt) error {
	if err := s.checkNFT(true); err != nil {
		return err
	}
	es, ctx, err := s.getExtensionStateAndContext()
	if err != nil {
		return err
	}
	return es.UnregisterPlanet(ctx, id.Int64())
}

func (s *chainScore) Ex_setPlanetOwner(id *common.HexInt, owner module.Address) error {
	if err := s.checkNFT(true); err != nil {
		return err
	}
	es, ctx, err := s.getExtensionStateAndContext()
	if err != nil {
		return err
	}
	return es.SetPlanetOwner(ctx, id.Int64(), owner)
}

func (s *chainScore) Ex_getPlanetInfo(id *common.HexInt) (map[string]interface{}, error) {
	if err := s.tryChargeCall(); err != nil {
		return nil, err
	}
	es, ctx, err := s.getExtensionStateAndContext()
	if err != nil {
		return nil, err
	}
	return es.GetPlanetInfo(ctx, id.Int64())
}

func (s *chainScore) Ex_reportPlanetWork(id *common.HexInt) error {
	if err := s.tryChargeCall(); err != nil {
		return err
	}
	es, ctx, err := s.getExtensionStateAndContext()
	if err != nil {
		return err
	}
	ok, err := es.IsPlanetManager(s.from)
	if err != nil {
		return err
	}
	if !ok {
		return scoreresult.AccessDeniedError.Errorf("NoPermission: %s", s.from)
	}
	return es.ReportPlanetWork(ctx, id.Int64())
}

func (s *chainScore) Ex_claimPlanetReward(ids []interface{}) error {
	if err := s.tryChargeCall(); err != nil {
		return err
	}
	es, ctx, err := s.getExtensionStateAndContext()
	if err != nil {
		return err
	}

	planetIds := make([]int64, len(ids))
	for i := 0; i < len(ids); i++ {
		planetIds[i] = (ids[i].(*common.HexInt)).Int64()
	}
	// PlanetOwner is checked in ExtensionStateImpl.ClaimPlanetReward()
	return es.ClaimPlanetReward(ctx, planetIds)
}

func (s *chainScore) Ex_getRewardInfoOf(id *common.HexInt) (map[string]interface{}, error) {
	if err := s.tryChargeCall(); err != nil {
		return nil, err
	}
	es, ctx, err := s.getExtensionStateAndContext()
	if err != nil {
		return nil, err
	}
	return es.GetRewardInfoOf(ctx, id.Int64())
}

func (s *chainScore) Ex_getRewardInfosOf(ids []interface{}) (map[string]interface{}, error) {
	if err := s.tryChargeCall(); err != nil {
		return nil, err
	}
	es, ctx, err := s.getExtensionStateAndContext()
	if err != nil {
		return nil, err
	}
	planetIds := make([]int64, len(ids))
	for i := 0; i < len(ids); i++ {
		planetIds[i] = (ids[i].(*common.HexInt)).Int64()
	}
	return es.GetRewardInfosOf(ctx, planetIds)
}

func (s *chainScore) Ex_getRewardInfo() (map[string]interface{}, error) {
	if err := s.tryChargeCall(); err != nil {
		return nil, err
	}
	if err := s.checkIntercallOnQueryMode(); err != nil {
		return nil, err
	}

	es, ctx, err := s.getExtensionStateAndContext()
	if err != nil {
		return nil, err
	}
	return es.GetRewardInfo(ctx)
}

func (s *chainScore) Ex_fallback() error {
	if err := s.tryChargeCall(); err != nil {
		return err
	}
	es, ctx, err := s.getExtensionStateAndContext()
	if err != nil {
		return err
	}
	return es.BurnCoin(ctx, s.value)
}

func (s *chainScore) Ex_setPrivateClaimableRate(
	numerator *common.HexInt, denominator *common.HexInt) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	es, _, err := s.getExtensionStateAndContext()
	if err != nil {
		return err
	}
	return es.SetPrivateClaimableRate(numerator.Int64(), denominator.Int64())
}

func (s *chainScore) Ex_getPrivateClaimableRate() (map[string]interface{}, error) {
	if err := s.tryChargeCall(); err != nil {
		return nil, err
	}
	es, _, err := s.getExtensionStateAndContext()
	if err != nil {
		return nil, err
	}
	return es.GetPrivateClaimableRate()
}

func (s *chainScore) Ex_withdrawLostTo(to module.Address) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	es, ctx, err := s.getExtensionStateAndContext()
	if err != nil {
		return err
	}
	return es.WithdrawLostTo(ctx, to)
}

func (s *chainScore) Ex_getLost() (*big.Int, error) {
	if err := s.tryChargeCall(); err != nil {
		return nil, err
	}
	es, _, err := s.getExtensionStateAndContext()
	if err != nil {
		return nil, err
	}
	return es.GetLost()
}

func (s *chainScore) Ex_setBlockVoteCheckParameters(period, allowance *common.HexInt) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	es, _, err := s.getExtensionStateAndContext()
	if err != nil {
		return err
	}
	return es.SetBlockVoteCheckParameters(period.Int64(), allowance.Int64())
}

func (s *chainScore) Ex_getBlockVoteCheckParameters() (map[string]interface{}, error) {
	if err := s.tryChargeCall(); err != nil {
		return nil, err
	}
	es, ctx, err := s.getExtensionStateAndContext()
	if err != nil {
		return nil, err
	}
	return es.GetBlockVoteCheckParameters(ctx)
}

func (s *chainScore) Ex_registerValidator(
	owner module.Address, nodePublicKey []byte, grade, name string) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	es, _, err := s.getExtensionStateAndContext()
	if err != nil {
		return err
	}
	return es.RegisterValidator(owner, nodePublicKey, grade, name)
}

func (s *chainScore) Ex_unregisterValidator(owner module.Address) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	es, _, err := s.getExtensionStateAndContext()
	if err != nil {
		return err
	}
	return es.UnregisterValidator(owner)
}

func (s *chainScore) Ex_getNetworkStatus() (map[string]interface{}, error) {
	if err := s.tryChargeCall(); err != nil {
		return nil, err
	}
	es, ctx, err := s.getExtensionStateAndContext()
	if err != nil {
		return nil, err
	}
	return es.GetNetworkStatus(ctx)
}

func (s *chainScore) Ex_setValidatorInfo(name, url string) error {
	if err := s.tryChargeCall(); err != nil {
		return err
	}
	es, ctx, err := s.getExtensionStateAndContext()
	if err != nil {
		return err
	}
	return es.SetValidatorInfo(ctx, name, url)
}

func (s *chainScore) Ex_enableValidator(owner module.Address) error {
	if err := s.tryChargeCall(); err != nil {
		return err
	}
	es, ctx, err := s.getExtensionStateAndContext()
	if err != nil {
		return err
	}
	return es.EnableValidator(ctx, owner)
}

func (s *chainScore) Ex_getValidatorInfo(owner module.Address) (map[string]interface{}, error) {
	if err := s.tryChargeCall(); err != nil {
		return nil, err
	}
	es, ctx, err := s.getExtensionStateAndContext()
	if err != nil {
		return nil, err
	}
	return es.GetValidatorInfo(ctx, owner)
}

func (s *chainScore) Ex_getValidatorStatus(owner module.Address) (map[string]interface{}, error) {
	if err := s.tryChargeCall(); err != nil {
		return nil, err
	}
	es, ctx, err := s.getExtensionStateAndContext()
	if err != nil {
		return nil, err
	}
	return es.GetValidatorStatus(ctx, owner)
}

func (s *chainScore) Ex_setNodePublicKey(publicKey []byte) error {
	if err := s.tryChargeCall(); err != nil {
		return err
	}
	es, ctx, err := s.getExtensionStateAndContext()
	if err != nil {
		return err
	}
	return es.SetNodePublicKey(ctx, publicKey)
}

func (s *chainScore) Ex_setValidatorCount(count int) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	es, _, err := s.getExtensionStateAndContext()
	if err != nil {
		return err
	}
	return es.SetValidatorCount(count)
}

func (s *chainScore) Ex_getValidatorCount() (int, error) {
	if err := s.tryChargeCall(); err != nil {
		return 0, err
	}
	es, _, err := s.getExtensionStateAndContext()
	if err != nil {
		return 0, err
	}
	return es.GetValidatorCount()
}

func (s *chainScore) Ex_getValidatorsOf(grade string) (map[string]interface{}, error) {
	if err := s.tryChargeCall(); err != nil {
		return nil, err
	}
	es, ctx, err := s.getExtensionStateAndContext()
	if err != nil {
		return nil, err
	}
	return es.GetValidatorsOf(ctx, grade)
}
