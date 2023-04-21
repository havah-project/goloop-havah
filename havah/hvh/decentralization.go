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
	oc := vs.Len() // old active validator count

	m := make(map[string]int)
	addrs := make([]module.Address, 0, vs.Len()*3/2)
	for i := 0; i < vs.Len(); i++ {
		v, _ := vs.Get(i)
		node := v.Address()
		m[hvhstate.ToKey(node)] = -1
		addrs = append(addrs, node)
	}

	// Next active validators
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
			// Reused active validator
			continue
		}

		owner, err = es.state.GetOwnerByNode(node)
		if err != nil {
			return err
		}
		if v < 0 {
			onActiveValidatorRemoved(cc, owner, node, "termchange")
		} else {
			onActiveValidatorAdded(cc, owner, node)
		}
	}

	// Actual number of active validators has been changed
	nc := len(validators)
	if oc != nc {
		onActiveValidatorCountChanged(cc, int64(oc), int64(nc))
	}

	return vs.Set(validators)
}

type penalizedInfo struct {
	owner module.Address
	validator module.Validator
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
	penalizedInfos := make([]*penalizedInfo, 0)

	// Check block vote
	for i, vote := range voted {
		voter, _ := voters.Get(i)
		node := voter.Address()
		if penalized, owner, err := es.state.OnBlockVote(node, vote); err == nil {
			if penalized {
				penalizedInfos = append(
					penalizedInfos, &penalizedInfo{owner: owner, validator: voter})
				onActiveValidatorPenalized(cc, owner, node)
			}
		}
	}

	// Penalized validators not found
	if len(penalizedInfos) == 0 {
		// No penalized validators
		return nil
	}
	return es.replacePenalizedActiveValidators(cc, penalizedInfos)
}

func (es *ExtensionStateImpl) IsItTimeToCheckBlockVote(blockIndexInTerm int64) bool {
	return es.state.IsItTimeToCheckBlockVote(blockIndexInTerm)
}

func (es *ExtensionStateImpl) replacePenalizedActiveValidators(
	cc hvhmodule.CallContext, penalizedInfos []*penalizedInfo) error {
	es.logger.Debugf("replacePenalizedActiveValidators() start: penalizedValidators=%d", len(penalizedInfos))

	m := make(map[int]*penalizedInfo)
	validatorState := cc.GetValidatorState()

	for _, info := range penalizedInfos {
		node := info.validator.Address()
		if idx := validatorState.IndexOf(node); idx >= 0 {
			m[idx] = info
		}
	}
	if len(m) == 0 {
		// No validator to remove
		es.logger.Debug("replacePenalizedActiveValidators() end")
		return nil
	}

	// Get the new standby validators from sub validators
	validatorsToAdd, err := es.state.GetNextActiveValidatorsAndChangeIndex(validatorState, len(m))
	if err != nil {
		return err
	}

	i := 0
	oc := validatorState.Len()
	var owner module.Address
	var validator module.Validator
	for idx := range m {
		info, _ := m[idx]
		onActiveValidatorRemoved(cc, info.owner, info.validator.Address(), "penalized")

		if i < len(validatorsToAdd) {
			validator, err = state.ValidatorFromAddress(validatorsToAdd[i])
			if err != nil {
				return err
			}
			node := validator.Address()
			if owner, err = es.state.GetOwnerByNode(node); err != nil {
				return err
			}
			if err = validatorState.SetAt(idx, validator); err != nil {
				return err
			}
			onActiveValidatorAdded(cc, owner, node)
			i++
		} else {
			validatorState.Remove(info.validator)
		}
	}

	if nc := validatorState.Len(); oc != nc {
		onActiveValidatorCountChanged(cc, int64(oc), int64(nc))
	}
	es.logger.Debugf("replaceActiveValidators() end: validatorsToAdd=%v", validatorsToAdd)
	return nil
}
