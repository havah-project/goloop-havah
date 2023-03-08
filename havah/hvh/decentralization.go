package hvh

import (
	"bytes"
	"sort"

	"github.com/icon-project/goloop/common/errors"
	"github.com/icon-project/goloop/havah/hvh/hvhstate"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/state"
)

func (es *ExtensionStateImpl) isDecentralizationPossible(cc hvhmodule.CallContext) bool {
	return es.state.IsDecentralizationPossible(cc.Revision().Value())
}

func (es *ExtensionStateImpl) initValidatorSet(cc hvhmodule.CallContext) error {
	// owners are owner list
	owners, err := es.state.GetAvailableValidators()
	if err != nil {
		return err
	}

	vis, err := es.state.GetValidatorInfos(owners)
	if err != nil {
		return err
	}

	sortValidatorInfos(vis)
	count := es.state.GetValidatorCount()
	activeSet, standbySet := splitValidatorInfos(vis, count)
	if err = setActiveValidators(cc, activeSet); err != nil {
		return err
	}
	if err = es.state.InitStandbyValidators(standbySet); err != nil {
		return err
	}
	return nil
}

func sortValidatorInfos(vis []*hvhstate.ValidatorInfo) {
	sort.Slice(vis, func(i, j int) bool {
		// GradeMain first
		if vis[i].Grade() > vis[j].Grade() {
			return true
		}
		return bytes.Compare(vis[i].Owner().Bytes(), vis[i].Owner().Bytes()) < 0
	})
}

func splitValidatorInfos(
	vis []*hvhstate.ValidatorInfo, validatorCount int) ([]*hvhstate.ValidatorInfo, []*hvhstate.ValidatorInfo) {
	if len(vis) <= validatorCount {
		return vis, nil
	}
	return vis[:validatorCount], vis[validatorCount:]
}

func setActiveValidators(cc hvhmodule.CallContext, activeSet []*hvhstate.ValidatorInfo) error {
	validators := make([]module.Validator, len(activeSet))
	for i, vi := range activeSet {
		validators[i], _ = state.ValidatorFromAddress(vi.Address())
	}
	return cc.SetValidators(validators)
}

func (es *ExtensionStateImpl) handleBlockVote(cc hvhmodule.CallContext) error {
	ci := cc.ConsensusInfo()
	if ci == nil {
		return errors.InvalidStateError.Errorf("Invalid ConsensusInfo")
	}

	// Assume that voters and validatorSet can be different
	voters := ci.Voters()
	voted := ci.Voted()
	validatorState := cc.GetValidatorState()
	penalizedValidatorIndexes := make(map[int]struct{})

	// Check block vote
	for i, vote := range voted {
		voter, _ := voters.Get(i)
		nodeAddr := voter.Address()
		if penalized, err := es.state.OnBlockVote(nodeAddr, vote); err == nil {
			if penalized {
				if index := validatorState.IndexOf(nodeAddr); index >= 0 {
					penalizedValidatorIndexes[index] = struct{}{}
				}
			}
		}
	}

	// Penalized validators not found
	penalizedCount := len(penalizedValidatorIndexes)
	if penalizedCount == 0 {
		// No penalized validators
		return nil
	}

	// Get standby validators to replace penalized active validators
	newActiveValidators, err := es.state.GetNextActiveValidators(penalizedCount)
	if err != nil {
		return err
	}

	// Build a new active validatorSet
	size := validatorState.Len()
	validators := make([]module.Validator, 0, size)

	// Append existing active validators except for the penalized
	for i := 0; i < size; i++ {
		if _, ok := penalizedValidatorIndexes[i]; !ok {
			validator, _ := validatorState.Get(i)
			validators = append(validators, validator)
		}
	}
	// Append new validators to replace the penalized
	for _, newAV := range newActiveValidators {
		if validator, err := state.ValidatorFromAddress(newAV); err == nil {
			validators = append(validators, validator)
		} else {
			return err
		}
	}
	// Update active validatorSet with new one
	return cc.SetValidators(validators)
}

func (es *ExtensionStateImpl) IsItTimeToCheckBlockVote(blockIndexInTerm int64) bool {
	return es.state.IsItTimeToCheckBlockVote(blockIndexInTerm)
}
