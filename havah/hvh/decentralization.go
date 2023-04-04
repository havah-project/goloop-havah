package hvh

import (
	"github.com/icon-project/goloop/common/errors"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/state"
)

func (es *ExtensionStateImpl) isDecentralizationPossible(cc hvhmodule.CallContext) bool {
	return es.state.IsDecentralizationPossible(cc.Revision().Value())
}

func (es *ExtensionStateImpl) initValidatorSet(cc hvhmodule.CallContext) error {
	validatorCount := int(es.state.GetActiveValidatorCount())

	// addrs contains validator node addresses
	addrs, err := es.state.GetMainValidators()
	if err != nil {
		return err
	}

	if count := validatorCount - len(addrs); count > 0 {
		subValidators, err := es.state.GetNextActiveValidatorsAndChangeIndex(nil, count)
		if err != nil {
			return err
		}
		if subValidators != nil {
			addrs = append(addrs, subValidators...)
		}
	}

	validators := make([]module.Validator, len(addrs))
	for i, addr := range addrs {
		validators[i], _ = state.ValidatorFromAddress(addr)
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
	newActiveValidators, err := es.state.GetNextActiveValidatorsAndChangeIndex(validatorState, penalizedCount)
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
