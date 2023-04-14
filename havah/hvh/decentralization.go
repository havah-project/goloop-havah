package hvh

import (
	"github.com/icon-project/goloop/common/errors"
	"github.com/icon-project/goloop/havah/hvh/hvhstate"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/state"
)

// Initialize active validator set due to term change
func (es *ExtensionStateImpl) initActiveValidatorSet(cc hvhmodule.CallContext) error {
	if cc.ReadOnlyMode() {
		return nil
	}

	validatorCount := int(es.state.GetActiveValidatorCount())
	// nodes contains validator node addresses
	nodes, err := es.state.GetMainValidators(validatorCount)
	if err != nil {
		return err
	}

	// Gets the next active validators from sub validators
	if count := validatorCount - len(nodes); count > 0 {
		subValidators, err := es.state.GetNextActiveValidatorsAndChangeIndex(nil, count)
		if err != nil {
			return err
		}
		if subValidators != nil {
			nodes = append(nodes, subValidators...)
		}
	}

	vs := cc.GetValidatorState()
	m := make(map[string]int)
	addrs := make([]module.Address, 0, vs.Len()*3/2)
	for i := 0; i < vs.Len(); i++ {
		v, _ := vs.Get(i)
		node := v.Address()
		m[hvhstate.ToKey(node)] = -1
		addrs = append(addrs, node)
	}

	validators := make([]module.Validator, len(nodes))
	for i, node := range nodes {
		validators[i], err = state.ValidatorFromAddress(node)
		if err != nil {
			return err
		}

		k := hvhstate.ToKey(node)
		num, ok := m[k]
		m[k] = num + 1
		if !ok {
			addrs = append(addrs, node)
		}
	}

	// Record eventLogs
	var owner module.Address
	for _, node := range addrs {
		k := hvhstate.ToKey(node)
		v := m[k]
		if v == 0 {
			continue
		}

		owner, err = es.state.GetOwnerByNode(node)
		if err != nil {
			return err
		}
		if v < 0 {
			onValidatorRemoved(cc, owner, node, "termchange")
		} else {
			onValidatorAdded(cc, owner, node)
		}
	}

	return vs.Set(validators)
}

func (es *ExtensionStateImpl) handleBlockVote(cc hvhmodule.CallContext) error {
	ci := cc.ConsensusInfo()
	if ci == nil {
		if cc.ReadOnlyMode() {
			// Skip to run handleBlockVote() because ConsensusInfo is nil on query call
			return nil
		}
		return errors.InvalidStateError.New("InvalidConsensusInfo")
	}

	// Assume that voters and validatorSet can be different
	voters := ci.Voters()
	voted := ci.Voted()
	validatorState := cc.GetValidatorState()
	validatorsToRemove := make([]module.Validator, 0)

	// Check block vote
	for i, vote := range voted {
		voter, _ := voters.Get(i)
		node := voter.Address()
		if penalized, owner, err := es.state.OnBlockVote(node, vote); err == nil {
			if penalized {
				onValidatorPenalized(cc, owner, node)
				if index := validatorState.IndexOf(node); index >= 0 {
					validatorsToRemove = append(validatorsToRemove, voter)
					onValidatorRemoved(cc, owner, node, "penalized")
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

	m := make(map[int]struct{})
	validatorState := cc.GetValidatorState()
	// Remove old validators
	for _, v := range validatorsToRemove {
		if idx := validatorState.IndexOf(v.Address()); idx >= 0 {
			m[idx] = struct{}{}
		}
	}

	if len(m) == 0 {
		// No validator to remove
		es.logger.Debugf("replaceActiveValidators(): end")
		return nil
	}

	// Get the new standby validators from sub validators
	validatorsToAdd, err := es.state.GetNextActiveValidatorsAndChangeIndex(validatorState, len(m))
	if err != nil {
		return err
	}

	size := validatorState.Len() - len(m) + len(validatorsToAdd)
	validators := make([]module.Validator, 0, size)

	j := 0
	var validator module.Validator
	for i := 0; i < validatorState.Len(); i++ {
		validator, _ = validatorState.Get(i)

		// If this validator should be removed
		if _, removed := m[i]; removed {
			if j < len(validatorsToAdd) {
				// If a new validator exists
				if validator, err = state.ValidatorFromAddress(validatorsToAdd[j]); err == nil {
					j++
					node := validator.Address()
					owner, err := es.state.GetOwnerByNode(node)
					if err != nil {
						return err
					}
					onValidatorAdded(cc, owner, node)
				} else {
					return err
				}
			} else {
				validator = nil
			}
		}

		if validator != nil {
			validators = append(validators, validator)
		}
	}

	es.logger.Debugf("replaceActiveValidators(): end: validatorsToAdd=%v", validatorsToAdd)
	return validatorState.Set(validators)
}
