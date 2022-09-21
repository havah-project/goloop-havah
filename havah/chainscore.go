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
	"github.com/icon-project/goloop/common/log"
	"github.com/icon-project/goloop/havah/hvh"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/havah/hvhutils"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/contract"
	"github.com/icon-project/goloop/service/scoreapi"
	"github.com/icon-project/goloop/service/scoredb"
	"github.com/icon-project/goloop/service/scoreresult"
	"github.com/icon-project/goloop/service/state"
	"github.com/icon-project/goloop/service/transaction"
)

const (
	StatusIllegalArgument = hvhmodule.StatusIllegalArgument
	StatusNotFound        = hvhmodule.StatusNotFound
)

type chainMethod struct {
	scoreapi.Method
	minVer, maxVer int
}

type chainScore struct {
	cc    contract.CallContext
	log   log.Logger
	from  module.Address
	value *big.Int
	gov   bool
}

var chainMethods = []*chainMethod{
	{scoreapi.Method{
		scoreapi.Function, "setRevision",
		scoreapi.FlagExternal, 1,
		[]scoreapi.Parameter{
			{"code", scoreapi.Integer, nil, nil},
		},
		nil,
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "setStepPrice",
		scoreapi.FlagExternal, 1,
		[]scoreapi.Parameter{
			{"price", scoreapi.Integer, nil, nil},
		},
		nil,
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "setStepCost",
		scoreapi.FlagExternal, 2,
		[]scoreapi.Parameter{
			{"type", scoreapi.String, nil, nil},
			{"cost", scoreapi.Integer, nil, nil},
		},
		nil,
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "setMaxStepLimit",
		scoreapi.FlagExternal, 2,
		[]scoreapi.Parameter{
			{"contextType", scoreapi.String, nil, nil},
			{"limit", scoreapi.Integer, nil, nil},
		},
		nil,
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "getRevision",
		scoreapi.FlagReadOnly, 0,
		nil,
		[]scoreapi.DataType{
			scoreapi.Integer,
		},
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "getStepPrice",
		scoreapi.FlagReadOnly, 0,
		nil,
		[]scoreapi.DataType{
			scoreapi.Integer,
		},
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "getStepCost",
		scoreapi.FlagReadOnly, 1,
		[]scoreapi.Parameter{
			{"type", scoreapi.String, nil, nil},
		},
		[]scoreapi.DataType{
			scoreapi.Integer,
		},
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "getStepCosts",
		scoreapi.FlagReadOnly, 0,
		nil,
		[]scoreapi.DataType{
			scoreapi.Dict,
		},
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "getMaxStepLimit",
		scoreapi.FlagReadOnly, 1,
		[]scoreapi.Parameter{
			{"contextType", scoreapi.String, nil, nil},
		},
		[]scoreapi.DataType{
			scoreapi.Integer,
		},
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "getServiceConfig",
		scoreapi.FlagReadOnly, 0,
		nil,
		[]scoreapi.DataType{
			scoreapi.Integer,
		},
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "getScoreOwner",
		scoreapi.FlagReadOnly | scoreapi.FlagExternal, 1,
		[]scoreapi.Parameter{
			{"score", scoreapi.Address, nil, nil},
		},
		[]scoreapi.DataType{
			scoreapi.Address,
		},
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "setScoreOwner",
		scoreapi.FlagExternal, 2,
		[]scoreapi.Parameter{
			{"score", scoreapi.Address, nil, nil},
			{"owner", scoreapi.Address, nil, nil},
		},
		nil,
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "setRoundLimitFactor",
		scoreapi.FlagExternal, 1,
		[]scoreapi.Parameter{
			{"factor", scoreapi.Integer, nil, nil},
		},
		nil,
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "getRoundLimitFactor",
		scoreapi.FlagReadOnly | scoreapi.FlagExternal, 0,
		nil,
		[]scoreapi.DataType{
			scoreapi.Integer,
		},
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "setUSDTPrice",
		scoreapi.FlagExternal, 1,
		[]scoreapi.Parameter{
			{"price", scoreapi.Integer, nil, nil},
		},
		nil,
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "getUSDTPrice",
		scoreapi.FlagReadOnly | scoreapi.FlagExternal, 0,
		nil,
		[]scoreapi.DataType{
			scoreapi.Integer,
		},
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "getIssueInfo",
		scoreapi.FlagReadOnly | scoreapi.FlagExternal, 0,
		nil,
		[]scoreapi.DataType{
			scoreapi.Dict,
		},
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "startRewardIssue",
		scoreapi.FlagExternal, 1,
		[]scoreapi.Parameter{
			{"height", scoreapi.Integer, nil, nil},
		},
		nil,
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "addPlanetManager",
		scoreapi.FlagExternal, 1,
		[]scoreapi.Parameter{
			{"address", scoreapi.Address, nil, nil},
		},
		nil,
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "removePlanetManager",
		scoreapi.FlagExternal, 1,
		[]scoreapi.Parameter{
			{"address", scoreapi.Address, nil, nil},
		},
		nil,
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "isPlanetManager",
		scoreapi.FlagReadOnly | scoreapi.FlagExternal, 1,
		[]scoreapi.Parameter{
			{"address", scoreapi.Address, nil, nil},
		},
		[]scoreapi.DataType{
			scoreapi.Bool,
		},
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "registerPlanet",
		scoreapi.FlagExternal, 6,
		[]scoreapi.Parameter{
			{"id", scoreapi.Integer, nil, nil},
			{"isPrivate", scoreapi.Bool, nil, nil},
			{"isCompany", scoreapi.Bool, nil, nil},
			{"owner", scoreapi.Address, nil, nil},
			{"usdt", scoreapi.Integer, nil, nil},
			{"price", scoreapi.Integer, nil, nil},
		},
		nil,
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "unregisterPlanet",
		scoreapi.FlagExternal, 1,
		[]scoreapi.Parameter{
			{"id", scoreapi.Integer, nil, nil},
		},
		nil,
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "setPlanetOwner",
		scoreapi.FlagExternal, 2,
		[]scoreapi.Parameter{
			{"id", scoreapi.Integer, nil, nil},
			{"owner", scoreapi.Address, nil, nil},
		},
		nil,
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "getPlanetInfo",
		scoreapi.FlagReadOnly | scoreapi.FlagExternal, 1,
		[]scoreapi.Parameter{
			{"id", scoreapi.Integer, nil, nil},
		},
		[]scoreapi.DataType{
			scoreapi.Dict,
		},
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "reportPlanetWork",
		scoreapi.FlagExternal, 1,
		[]scoreapi.Parameter{
			{"id", scoreapi.Integer, nil, nil},
		},
		nil,
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "claimPlanetReward",
		scoreapi.FlagExternal, 1,
		[]scoreapi.Parameter{
			{"ids", scoreapi.ListTypeOf(1, scoreapi.Integer), nil, nil},
		},
		nil,
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "getRewardInfoOf",
		scoreapi.FlagReadOnly | scoreapi.FlagExternal, 1,
		[]scoreapi.Parameter{
			{"id", scoreapi.Integer, nil, nil},
		},
		[]scoreapi.DataType{
			scoreapi.Dict,
		},
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "getRewardInfo",
		scoreapi.FlagReadOnly | scoreapi.FlagExternal, 0,
		nil,
		[]scoreapi.DataType{
			scoreapi.Dict,
		},
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "setPrivateClaimableRate",
		scoreapi.FlagExternal, 2,
		[]scoreapi.Parameter{
			{"numerator", scoreapi.Integer, nil, nil},
			{"denominator", scoreapi.Integer, nil, nil},
		},
		nil,
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "getPrivateClaimableRate",
		scoreapi.FlagReadOnly | scoreapi.FlagExternal, 0,
		nil,
		[]scoreapi.DataType{
			scoreapi.Dict,
		},
	}, 0, 0},
	{scoreapi.Method{scoreapi.Function, "addDeployer",
		scoreapi.FlagExternal, 1,
		[]scoreapi.Parameter{
			{"address", scoreapi.Address, nil, nil},
		},
		nil,
	}, 0, 0},
	{scoreapi.Method{scoreapi.Function, "removeDeployer",
		scoreapi.FlagExternal, 1,
		[]scoreapi.Parameter{
			{"address", scoreapi.Address, nil, nil},
		},
		nil,
	}, 0, 0},
	{scoreapi.Method{scoreapi.Function, "isDeployer",
		scoreapi.FlagReadOnly | scoreapi.FlagExternal, 1,
		[]scoreapi.Parameter{
			{"address", scoreapi.Address, nil, nil},
		},
		[]scoreapi.DataType{
			scoreapi.Integer,
		},
	}, 0, 0},
	{scoreapi.Method{scoreapi.Function, "getDeployers",
		scoreapi.FlagReadOnly | scoreapi.FlagExternal, 0,
		nil,
		[]scoreapi.DataType{
			scoreapi.List,
		},
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "setTimestampThreshold",
		scoreapi.FlagExternal, 1,
		[]scoreapi.Parameter{
			{"threshold", scoreapi.Integer, nil, nil},
		},
		nil,
	}, 0, 0},
	{scoreapi.Method{
		scoreapi.Function, "getTimestampThreshold",
		scoreapi.FlagReadOnly | scoreapi.FlagExternal, 0,
		nil,
		[]scoreapi.DataType{
			scoreapi.Integer,
		},
	}, 0, 0},
	{scoreapi.Method{scoreapi.Function, "grantValidator",
		scoreapi.FlagExternal, 1,
		[]scoreapi.Parameter{
			{"address", scoreapi.Address, nil, nil},
		},
		nil,
	}, 0, 0},
	{scoreapi.Method{scoreapi.Function, "revokeValidator",
		scoreapi.FlagExternal, 1,
		[]scoreapi.Parameter{
			{"address", scoreapi.Address, nil, nil},
		},
		nil,
	}, 0, 0},
	{scoreapi.Method{scoreapi.Function, "getValidators",
		scoreapi.FlagReadOnly, 0,
		nil,
		[]scoreapi.DataType{
			scoreapi.List,
		},
	}, 0, 0},
	{scoreapi.Method{scoreapi.Fallback, "fallback",
		scoreapi.FlagPayable, 0,
		nil,
		nil,
	}, 0, 0},
}

func initFeeConfig(cfg *FeeConfig, as state.AccountState) error {
	if cfg != nil {
		if err := applyStepLimits(cfg, as); err != nil {
			return err
		}
		if err := applyStepCosts(cfg, as); err != nil {
			return err
		}
		if err := applyStepPrice(as, &cfg.StepPrice.Int); err != nil {
			return err
		}
	}
	return nil
}

func applyStepLimits(fee *FeeConfig, as state.AccountState) error {
	stepLimitTypes := scoredb.NewArrayDB(as, state.VarStepLimitTypes)
	stepLimitDB := scoredb.NewDictDB(as, state.VarStepLimit, 1)
	if fee.StepLimit != nil {
		for _, k := range state.AllStepLimitTypes {
			if err := stepLimitTypes.Put(k); err != nil {
				return err
			}
			icost := fee.StepLimit[k]
			if err := stepLimitDB.Set(k, icost.Value); err != nil {
				return err
			}
		}
	} else {
		for _, k := range state.AllStepLimitTypes {
			if err := stepLimitTypes.Put(k); err != nil {
				return err
			}
			if err := stepLimitDB.Set(k, 0); err != nil {
				return err
			}
		}
	}
	return nil
}

func applyStepCosts(fee *FeeConfig, as state.AccountState) error {
	stepTypes := scoredb.NewArrayDB(as, state.VarStepTypes)
	stepCostDB := scoredb.NewDictDB(as, state.VarStepCosts, 1)
	if fee.StepCosts != nil {
		for k := range fee.StepCosts {
			if !state.IsValidStepType(k) {
				return scoreresult.IllegalFormatError.Errorf("InvalidStepType(%s)", k)
			}
		}
		for _, k := range state.AllStepTypes {
			cost, ok := fee.StepCosts[k]
			if !ok {
				continue
			}
			if err := stepTypes.Put(k); err != nil {
				return err
			}
			if err := stepCostDB.Set(k, cost.Value); err != nil {
				return err
			}
		}
	} else {
		for _, k := range state.InitialStepTypes {
			if err := stepTypes.Put(k); err != nil {
				return err
			}
			if err := stepCostDB.Set(k, 0); err != nil {
				return err
			}
		}
	}
	return nil
}

func applyStepPrice(as state.AccountState, price *big.Int) error {
	return scoredb.NewVarDB(as, state.VarStepPrice).Set(price)
}

func (s *chainScore) initPlatformConfig(cfg *hvh.PlatformConfig) error {
	if cfg == nil {
		return nil
	}

	// Initialize ExtensionState
	s.cc.GetExtensionState().Reset(hvh.NewExtensionSnapshot(s.cc.Database(), nil))

	es, err := s.getExtensionState()
	if err != nil {
		return err
	}
	if err = es.InitPlatformConfig(cfg); err != nil {
		return err
	}
	return nil
}

func (s *chainScore) Install(param []byte) error {
	s.log.Debugf("chainScore start")
	if s.from != nil {
		return scoreresult.AccessDeniedError.New("AccessDeniedToInstallChainSCORE")
	}

	cfg := newChainConfig()

	var systemConfig int
	var revision int

	if param != nil {
		if err := json.Unmarshal(param, cfg); err != nil {
			return scoreresult.Errorf(
				module.StatusIllegalFormat,
				"Failed to parse parameter for chainScore. err(%+v)\n",
				err,
			)
		}
	}

	as := s.cc.GetAccountState(state.SystemID)

	if cfg.Revision.Value != 0 {
		revision = int(cfg.Revision.Value)
		if revision > hvhmodule.MaxRevision {
			return scoreresult.IllegalFormatError.Errorf(
				"RevisionIsHigherMax(%d > %d)", revision, hvhmodule.MaxRevision)
		} else if revision > hvhmodule.LatestRevision {
			s.log.Warnf("Revision in genesis is higher than latest(%d > %d)",
				revision, hvhmodule.LatestRevision)
		}
	}
	if err := scoredb.NewVarDB(as, state.VarRevision).Set(revision); err != nil {
		return err
	}

	if cfg.RoundLimitFactor != nil {
		factor := cfg.RoundLimitFactor.Value
		if err := scoredb.NewVarDB(as, state.VarRoundLimitFactor).Set(factor); err != nil {
			return err
		}
	}

	if cfg.BlockInterval != nil {
		blockInterval := cfg.BlockInterval.Value
		if err := scoredb.NewVarDB(as, state.VarBlockInterval).Set(blockInterval); err != nil {
			return err
		}
	}

	validators := make([]module.Validator, len(cfg.ValidatorList))
	for i, validator := range cfg.ValidatorList {
		validators[i], _ = state.ValidatorFromAddress(validator)
		s.log.Debugf("add validator %d: %v", i, validator)
	}
	if err := s.cc.GetValidatorState().Set(validators); err != nil {
		return errors.CriticalUnknownError.Wrap(err, "FailToSetValidators")
	}

	if err := scoredb.NewVarDB(as, state.VarChainID).Set(s.cc.ChainID()); err != nil {
		return err
	}

	if len(validators) > 0 {
		if err := s.cc.GetValidatorState().Set(validators); err != nil {
			return errors.CriticalUnknownError.Wrap(err, "FailToSetValidators")
		}
	}

	feeConfig := cfg.Fee
	systemConfig |= state.SysConfigFee
	if feeConfig == nil {
		feeConfig = new(FeeConfig)
		feeConfig.StepLimit = map[string]common.HexInt64{
			state.StepLimitTypeInvoke: {0x9502f900},
			state.StepLimitTypeQuery:  {0x2faf080},
		}
		feeConfig.StepCosts = map[string]common.HexInt64{
			state.StepTypeSchema:         {1},
			state.StepTypeDefault:        {100_000},
			state.StepTypeContractCall:   {25_000},
			state.StepTypeContractCreate: {1_000_000_000},
			state.StepTypeContractUpdate: {1_000_000_000},
			state.StepTypeContractSet:    {15_000},
			state.StepTypeGetBase:        {3_000},
			state.StepTypeGet:            {25},
			state.StepTypeSetBase:        {10_000},
			state.StepTypeSet:            {320},
			state.StepTypeDeleteBase:     {200},
			state.StepTypeDelete:         {-240},
			state.StepTypeInput:          {200},
			state.StepTypeLogBase:        {5_000},
			state.StepTypeLog:            {100},
			state.StepTypeApiCall:        {10_000},
		}
	}
	if err := initFeeConfig(feeConfig, as); err != nil {
		return err
	}

	if cfg.TSThreshold != nil {
		ts := int(cfg.TSThreshold.Value)
		if err := scoredb.NewVarDB(as, state.VarTimestampThreshold).Set(ts); err != nil {
			return err
		}
	}

	platformConfig := cfg.Havah
	if platformConfig != nil {
		if err := s.initPlatformConfig(platformConfig); err != nil {
			return transaction.InvalidGenesisError.Wrap(err, "Failed to initialize platformConfig")
		}
	} else {
		return transaction.InvalidGenesisError.New("NoPlatformConfiguration(field=\"havah\")")
	}

	if err := scoredb.NewVarDB(as, state.VarServiceConfig).Set(systemConfig); err != nil {
		return err
	}

	err := s.handleRevisionChange(as, hvhmodule.Revision0, revision)
	s.log.Debugf("chainScore end")
	return err
}

type handleRevFunc func(*chainScore, state.AccountState, int, int) error

var handleRevFuncs = map[int]handleRevFunc{
	hvhmodule.Revision0: handleRev1,
}

func handleRev1(s *chainScore, as state.AccountState, oldRev, newRev int) error {
	return nil
}

func (s *chainScore) handleRevisionChange(as state.AccountState, oldRev, newRev int) error {
	if oldRev >= newRev {
		return nil
	}
	for rev := oldRev; rev < newRev; rev++ {
		if fn, ok := handleRevFuncs[rev]; ok {
			if err := fn(s, as, oldRev, newRev); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *chainScore) Update(param []byte) error {
	return nil
}

func (s *chainScore) GetAPI() *scoreapi.Info {
	ass := s.cc.GetAccountSnapshot(state.SystemID)
	store := scoredb.NewStateStoreWith(ass)
	revision := int(scoredb.NewVarDB(store, state.VarRevision).Int64())
	methods := make([]*scoreapi.Method, 0, len(chainMethods))

	for _, m := range chainMethods {
		if m.minVer <= revision && (m.maxVer == 0 || revision <= m.maxVer) {
			methods = append(methods, &m.Method)
		}
	}
	return scoreapi.NewInfo(methods)
}

func (s *chainScore) checkGovernance(charge bool) error {
	if !s.gov {
		if charge {
			if err := s.cc.ApplyCallSteps(); err != nil {
				return err
			}
		}
		return scoreresult.New(module.StatusAccessDenied, "NoPermission")
	}
	return nil
}

func newChainScore(cc contract.CallContext, from module.Address, value *big.Int) (contract.SystemScore, error) {
	fromGov := cc.Governance().Equal(from)
	return &chainScore{
		cc:    cc,
		from:  from,
		value: value,
		log:   hvhutils.NewLogger(cc.Logger()),
		gov:   fromGov,
	}, nil
}
