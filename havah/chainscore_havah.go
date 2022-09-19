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

func (s *chainScore) newCallContext() hvhmodule.CallContext {
	return hvh.NewCallContext(s.cc, s.from)
}

func (s *chainScore) Ex_getUSDTPrice() (*big.Int, error) {
	es, err := s.getExtensionState()
	if err != nil {
		return nil, err
	}
	return es.GetUSDTPrice()
}

func (s *chainScore) Ex_setUSDTPrice(price *common.HexInt) error {
	// TODO: caller restriction
	es, err := s.getExtensionState()
	if err != nil {
		return err
	}
	return es.SetUSDTPrice(price.Value())
}

func (s *chainScore) Ex_getIssueInfo() (map[string]interface{}, error) {
	es, err := s.getExtensionState()
	if err != nil {
		return nil, err
	}
	return es.GetIssueInfo(s.newCallContext())
}

func (s *chainScore) Ex_startRewardIssue(height *common.HexInt) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	es, err := s.getExtensionState()
	if err != nil {
		return err
	}
	startBH := height.Int64()
	if startBH <= s.cc.BlockHeight() {
		return scoreresult.RevertedError.New("Invalid height")
	}
	return es.StartRewardIssue(s.newCallContext(), startBH)
}

func (s *chainScore) Ex_addPlanetManager(address module.Address) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	es, err := s.getExtensionState()
	if err != nil {
		return err
	}
	return es.AddPlanetManager(address)
}

func (s *chainScore) Ex_removePlanetManager(address module.Address) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	es, err := s.getExtensionState()
	if err != nil {
		return err
	}
	return es.RemovePlanetManager(address)
}

func (s *chainScore) Ex_isPlanetManager(address module.Address) (bool, error) {
	es, err := s.getExtensionState()
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
	es, err := s.getExtensionState()
	if err != nil {
		return err
	}
	return es.RegisterPlanet(
		s.newCallContext(), id.Int64(),
		isPrivate, isCompany, owner, usdt.Value(), price.Value())
}

func (s *chainScore) Ex_unregisterPlanet(id *common.HexInt) error {
	if err := s.checkNFT(true); err != nil {
		return err
	}
	es, err := s.getExtensionState()
	if err != nil {
		return err
	}
	return es.UnregisterPlanet(s.newCallContext(), id.Int64())
}

func (s *chainScore) Ex_setPlanetOwner(id *common.HexInt, owner module.Address) error {
	if err := s.checkNFT(true); err != nil {
		return err
	}
	es, err := s.getExtensionState()
	if err != nil {
		return err
	}
	return es.SetPlanetOwner(s.newCallContext(), id.Int64(), owner)
}

func (s *chainScore) Ex_getPlanetInfo(id *common.HexInt) (map[string]interface{}, error) {
	if err := s.tryChargeCall(); err != nil {
		return nil, err
	}
	es, err := s.getExtensionState()
	if err != nil {
		return nil, err
	}
	return es.GetPlanetInfo(s.newCallContext(), id.Int64())
}

func (s *chainScore) Ex_reportPlanetWork(id *common.HexInt) error {
	if err := s.tryChargeCall(); err != nil {
		return err
	}
	es, err := s.getExtensionState()
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
	return es.ReportPlanetWork(s.newCallContext(), id.Int64())
}

func (s *chainScore) Ex_claimPlanetReward(ids []interface{}) error {
	if err := s.tryChargeCall(); err != nil {
		return err
	}
	es, err := s.getExtensionState()
	if err != nil {
		return err
	}

	planetIds := make([]int64, len(ids))
	for i := 0; i < len(ids); i++ {
		planetIds[i] = (ids[i].(*common.HexInt)).Int64()
	}
	// PlanetOwner is checked in ExtensionStateImpl.ClaimPlanetReward()
	return es.ClaimPlanetReward(s.newCallContext(), planetIds)
}

func (s *chainScore) Ex_getRewardInfoOf(id *common.HexInt) (map[string]interface{}, error) {
	if err := s.tryChargeCall(); err != nil {
		return nil, err
	}
	es, err := s.getExtensionState()
	if err != nil {
		return nil, err
	}
	return es.GetRewardInfoOf(s.newCallContext(), id.Int64())
}

func (s *chainScore) Ex_getRewardInfo() (map[string]interface{}, error) {
	if err := s.tryChargeCall(); err != nil {
		return nil, err
	}
	es, err := s.getExtensionState()
	if err != nil {
		return nil, err
	}
	return es.GetRewardInfo(s.newCallContext())
}

func (s *chainScore) Ex_fallback() error {
	if err := s.tryChargeCall(); err != nil {
		return err
	}
	es, err := s.getExtensionState()
	if err != nil {
		return err
	}
	return es.BurnCoin(s.newCallContext(), s.value)
}

func (s *chainScore) Ex_setPrivateClaimableRate(
	numerator *common.HexInt, denominator *common.HexInt) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	es, err := s.getExtensionState()
	if err != nil {
		return err
	}
	return es.SetPrivateClaimableRate(numerator.Int64(), denominator.Int64())
}

func (s *chainScore) Ex_getPrivateClaimableRate() (map[string]interface{}, error) {
	if err := s.tryChargeCall(); err != nil {
		return nil, err
	}
	es, err := s.getExtensionState()
	if err != nil {
		return nil, err
	}
	return es.GetPrivateClaimableRate()
}
