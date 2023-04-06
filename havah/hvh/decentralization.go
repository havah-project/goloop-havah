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
	addrs, err := es.state.GetMainValidators(validatorCount)
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
		if cc.ReadOnlyMode() {
			// Skip to run handleBlockVote() because ConsensusInfo is nil on query call
			return nil
		}
		return errors.InvalidStateError.Errorf("Invalid ConsensusInfo")
	}

	// Assume that voters and validatorSet can be different
	voters := ci.Voters()
	voted := ci.Voted()
	validatorState := cc.GetValidatorState()
	validatorsToRemove := make([]module.Validator, 0)

	// Check block vote
	for i, vote := range voted {
		voter, _ := voters.Get(i)
		nodeAddr := voter.Address()
		if penalized, err := es.state.OnBlockVote(nodeAddr, vote); err == nil {
			if penalized {
				if index := validatorState.IndexOf(nodeAddr); index >= 0 {
					validatorsToRemove = append(validatorsToRemove, voter)
				}
			}
		}
	}

	// Penalized validators not found
	if len(validatorsToRemove) == 0 {
		// No penalized validators
		return nil
	}

	return es.replaceActiveValidators(cc, validatorsToRemove)
}

func (es *ExtensionStateImpl) IsItTimeToCheckBlockVote(blockIndexInTerm int64) bool {
	return es.state.IsItTimeToCheckBlockVote(blockIndexInTerm)
}

func (es *ExtensionStateImpl) replaceActiveValidators(
	cc hvhmodule.CallContext, validatorsToRemove []module.Validator) error {
	es.logger.Debugf("replaceActiveValidators(): start: validatorsToRemove=%v", validatorsToRemove)

	if len(validatorsToRemove) == 0 {
		// Nothing to remove
		return nil
	}

	count := 0  // Number of removed validators
	validatorState := cc.GetValidatorState()
	// Remove old validators
	for _, validator := range validatorsToRemove {
		if validatorState.Remove(validator) {
			count++
		}
	}

	if count == 0 {
		// No validator is removed
		es.logger.Debugf("replaceActiveValidators(): end")
		return nil
	}

	// Get the new standby validators from sub validators
	validatorsToAdd, err := es.state.GetNextActiveValidatorsAndChangeIndex(validatorState, count)
	if err != nil {
		return err
	}

	// Add new validators
	var validator module.Validator
	for _, addr := range validatorsToAdd {
		validator, err = state.ValidatorFromAddress(addr)
		if err != nil {
			return err
		}
		if err = validatorState.Add(validator); err != nil {
			return err
		}
	}

	es.logger.Debugf("replaceActiveValidators(): end: validatorsToAdd=%v", validatorsToAdd)
	return nil
}
