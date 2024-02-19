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
	"github.com/icon-project/goloop/common/errors"
	"github.com/icon-project/goloop/havah/hvh"
	"github.com/icon-project/goloop/havah/hvhmodule"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/scoredb"
	"github.com/icon-project/goloop/service/scoreresult"
	"github.com/icon-project/goloop/service/state"
)

const (
	CIDForVegaNet = 0x630a4
)

func (s *chainScore) tryChargeCall() error {
	if s.gov {
		return nil
	}
	if err := s.cc.ApplyCallSteps(); err != nil {
		return err
	}
	return nil
}

// Ex_setRevision sets the system revision to the given number.
// This can only be called by the governance SCORE.
func (s *chainScore) Ex_setRevision(code *common.HexInt) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	if hvhmodule.MaxRevision < code.Int64() {
		return scoreresult.Errorf(StatusIllegalArgument,
			"IllegalArgument(max=%#x,new=%s)", hvhmodule.MaxRevision, code)
	}

	as := s.cc.GetAccountState(state.SystemID)
	r := scoredb.NewVarDB(as, state.VarRevision).Int64()
	if code.Int64() < r {
		return scoreresult.Errorf(StatusIllegalArgument,
			"IllegalArgument(current=%#x,new=%s)", r, code)
	}

	if err := scoredb.NewVarDB(as, state.VarRevision).Set(code); err != nil {
		return err
	}

	cid := s.cc.ChainID()
	if cid == CIDForVegaNet && int(code.Int64()) == hvhmodule.Revision6 {
		// Replay a bug in handleRevisionChange() on VegaNet
		return nil
	}

	if err := s.handleRevisionChange(int(r), int(code.Int64())); err != nil {
		return err
	}

	if err := as.MigrateForRevision(s.cc.ToRevision(int(code.Int64()))); err != nil {
		return err
	}
	as.SetAPIInfo(s.GetAPI())
	return nil
}

func (s *chainScore) getScoreAddress(txHash []byte) module.Address {
	sysAs := s.cc.GetAccountState(state.SystemID)
	h2a := scoredb.NewDictDB(sysAs, state.VarTxHashToAddress, 1)
	value := h2a.Get(txHash)
	if value != nil {
		return value.Address()
	}
	return nil
}

func (s *chainScore) Ex_setStepPrice(price *common.HexInt) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	as := s.cc.GetAccountState(state.SystemID)
	return scoredb.NewVarDB(as, state.VarStepPrice).Set(price)
}

func (s *chainScore) Ex_setStepCost(costType string, cost *common.HexInt) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	if !state.IsValidStepType(costType) {
		return scoreresult.IllegalFormatError.Errorf("InvalidStepType(%s)", costType)
	}
	costZero := cost.Sign() == 0
	as := s.cc.GetAccountState(state.SystemID)
	stepCostDB := scoredb.NewDictDB(as, state.VarStepCosts, 1)
	stepTypes := scoredb.NewArrayDB(as, state.VarStepTypes)
	if stepCostDB.Get(costType) == nil && !costZero {
		if err := stepTypes.Put(costType); err != nil {
			return err
		}
	}
	if costZero {
		// remove the step type and cost
		for i := 0; i < stepTypes.Size(); i++ {
			if stepTypes.Get(i).String() == costType {
				last := stepTypes.Pop().String()
				if i < stepTypes.Size() {
					if err := stepTypes.Set(i, last); err != nil {
						return err
					}
				}
				return stepCostDB.Delete(costType)
			}
		}
		return nil
	} else {
		return stepCostDB.Set(costType, cost)
	}
}

func (s *chainScore) Ex_setMaxStepLimit(contextType string, cost *common.HexInt) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	as := s.cc.GetAccountState(state.SystemID)
	stepLimitDB := scoredb.NewDictDB(as, state.VarStepLimit, 1)
	if stepLimitDB.Get(contextType) == nil {
		stepLimitTypes := scoredb.NewArrayDB(as, state.VarStepLimitTypes)
		if err := stepLimitTypes.Put(contextType); err != nil {
			return err
		}
	}
	return stepLimitDB.Set(contextType, cost)
}

func (s *chainScore) Ex_getRevision() (int64, error) {
	if err := s.tryChargeCall(); err != nil {
		return 0, err
	}
	as := s.cc.GetAccountState(state.SystemID)
	return scoredb.NewVarDB(as, state.VarRevision).Int64(), nil
}

func (s *chainScore) Ex_getStepPrice() (int64, error) {
	if err := s.tryChargeCall(); err != nil {
		return 0, err
	}
	as := s.cc.GetAccountState(state.SystemID)
	return scoredb.NewVarDB(as, state.VarStepPrice).Int64(), nil
}

func (s *chainScore) Ex_getStepCost(t string) (int64, error) {
	if err := s.tryChargeCall(); err != nil {
		return 0, err
	}
	as := s.cc.GetAccountState(state.SystemID)
	stepCostDB := scoredb.NewDictDB(as, state.VarStepCosts, 1)
	if v := stepCostDB.Get(t); v != nil {
		return v.Int64(), nil
	}
	return 0, nil
}

func (s *chainScore) Ex_getStepCosts() (map[string]interface{}, error) {
	if err := s.tryChargeCall(); err != nil {
		return nil, err
	}
	as := s.cc.GetAccountState(state.SystemID)

	stepCosts := make(map[string]interface{})
	stepTypes := scoredb.NewArrayDB(as, state.VarStepTypes)
	stepCostDB := scoredb.NewDictDB(as, state.VarStepCosts, 1)
	tcount := stepTypes.Size()
	for i := 0; i < tcount; i++ {
		tname := stepTypes.Get(i).String()
		stepCosts[tname] = stepCostDB.Get(tname).Int64()
	}
	return stepCosts, nil
}

func (s *chainScore) Ex_getMaxStepLimit(contextType string) (int64, error) {
	if err := s.tryChargeCall(); err != nil {
		return 0, err
	}
	as := s.cc.GetAccountState(state.SystemID)
	stepLimitDB := scoredb.NewDictDB(as, state.VarStepLimit, 1)
	if v := stepLimitDB.Get(contextType); v != nil {
		return v.Int64(), nil
	}
	return 0, nil
}

func (s *chainScore) Ex_getServiceConfig() (int64, error) {
	if err := s.tryChargeCall(); err != nil {
		return 0, err
	}
	as := s.cc.GetAccountState(state.SystemID)
	return scoredb.NewVarDB(as, state.VarServiceConfig).Int64(), nil
}

func (s *chainScore) Ex_getScoreOwner(score module.Address) (module.Address, error) {
	if err := s.tryChargeCall(); err != nil {
		return nil, err
	}
	return hvh.NewCallContext(s.cc, s.from).GetScoreOwner(score)
}

func (s *chainScore) Ex_setScoreOwner(score module.Address, owner module.Address) error {
	if err := s.tryChargeCall(); err != nil {
		return err
	}
	return hvh.NewCallContext(s.cc, s.from).SetScoreOwner(s.from, score, owner)
}

func (s *chainScore) Ex_getRoundLimitFactor() (int64, error) {
	if err := s.tryChargeCall(); err != nil {
		return 0, err
	}
	as := s.cc.GetAccountState(state.SystemID)
	return scoredb.NewVarDB(as, state.VarRoundLimitFactor).Int64(), nil
}

func (s *chainScore) Ex_setRoundLimitFactor(f *common.HexInt) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	if f.Sign() < 0 {
		return scoreresult.New(StatusIllegalArgument, "IllegalArgument")
	}
	as := s.cc.GetAccountState(state.SystemID)
	factor := scoredb.NewVarDB(as, state.VarRoundLimitFactor)
	return factor.Set(f)
}

func (s *chainScore) Ex_addDeployer(address module.Address) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	as := s.cc.GetAccountState(state.SystemID)
	db := scoredb.NewArrayDB(as, state.VarDeployers)
	for i := 0; i < db.Size(); i++ {
		if db.Get(i).Address().Equal(address) == true {
			return nil
		}
	}
	return db.Put(address)
}

func (s *chainScore) Ex_removeDeployer(address module.Address) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	as := s.cc.GetAccountState(state.SystemID)
	db := scoredb.NewArrayDB(as, state.VarDeployers)
	for i := 0; i < db.Size(); i++ {
		if db.Get(i).Address().Equal(address) == true {
			rAddr := db.Pop().Address()
			if i < db.Size() { // addr is not rAddr
				if err := db.Set(i, rAddr); err != nil {
					return err
				}
			}
			break
		}
	}
	return nil
}

func (s *chainScore) Ex_isDeployer(address module.Address) (int, error) {
	if err := s.tryChargeCall(); err != nil {
		return 0, err
	}
	as := s.cc.GetAccountState(state.SystemID)
	db := scoredb.NewArrayDB(as, state.VarDeployers)
	for i := 0; i < db.Size(); i++ {
		if db.Get(i).Address().Equal(address) == true {
			return 1, nil
		}
	}
	return 0, nil
}

func (s *chainScore) Ex_getDeployers() ([]interface{}, error) {
	if err := s.tryChargeCall(); err != nil {
		return nil, err
	}
	as := s.cc.GetAccountState(state.SystemID)
	db := scoredb.NewArrayDB(as, state.VarDeployers)
	deployers := make([]interface{}, db.Size())
	for i := 0; i < db.Size(); i++ {
		deployers[i] = db.Get(i).Address()
	}
	return deployers, nil
}

func (s *chainScore) Ex_setTimestampThreshold(threshold *common.HexInt) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	as := s.cc.GetAccountState(state.SystemID)
	db := scoredb.NewVarDB(as, state.VarTimestampThreshold)
	return db.Set(threshold)
}

func (s *chainScore) Ex_getTimestampThreshold() (int64, error) {
	if err := s.tryChargeCall(); err != nil {
		return 0, err
	}
	as := s.cc.GetAccountState(state.SystemID)
	db := scoredb.NewVarDB(as, state.VarTimestampThreshold)
	return db.Int64(), nil
}

func (s *chainScore) Ex_grantValidator(address module.Address) error {
	if err := s.tryChargeCall(); err != nil {
		return err
	}
	if address == nil {
		return scoreresult.ErrInvalidParameter
	}
	if err := s.checkGovernance(false); err != nil {
		return err
	}
	if address.IsContract() {
		return scoreresult.New(StatusIllegalArgument, "address should be EOA")
	}

	if s.cc.MembershipEnabled() {
		found := false
		as := s.cc.GetAccountState(state.SystemID)
		db := scoredb.NewArrayDB(as, state.VarMembers)
		for i := 0; i < db.Size(); i++ {
			if db.Get(i).Address().Equal(address) {
				found = true
				break
			}
		}
		if !found {
			return scoreresult.New(StatusIllegalArgument, "NotInMembers")
		}
	}

	if v, err := state.ValidatorFromAddress(address); err == nil {
		return s.cc.GetValidatorState().Add(v)
	} else {
		return err
	}
}

func (s *chainScore) Ex_revokeValidator(address module.Address) error {
	if err := s.tryChargeCall(); err != nil {
		return err
	}
	if address == nil {
		return scoreresult.ErrInvalidParameter
	}
	if err := s.checkGovernance(false); err != nil {
		return err
	}
	if address.IsContract() {
		return scoreresult.New(StatusIllegalArgument, "AddressIsContract")
	}
	if v, err := state.ValidatorFromAddress(address); err == nil {
		vl := s.cc.GetValidatorState()
		if ok := vl.Remove(v); !ok {
			return scoreresult.New(StatusNotFound, "NotFound")
		}
		if vl.Len() == 0 {
			return scoreresult.New(StatusIllegalArgument, "OnlyValidator")
		}
		return nil
	} else {
		return err
	}
}

func (s *chainScore) Ex_getValidators() ([]interface{}, error) {
	if err := s.tryChargeCall(); err != nil {
		return nil, err
	}
	vs := s.cc.GetValidatorState()
	validators := make([]interface{}, vs.Len())
	for i := 0; i < vs.Len(); i++ {
		if v, ok := vs.Get(i); ok {
			validators[i] = v.Address()
		} else {
			return nil, errors.CriticalUnknownError.New("Unexpected access failure")
		}
	}
	return validators, nil
}
