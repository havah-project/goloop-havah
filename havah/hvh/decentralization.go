package hvh

import (
	"bytes"
	"sort"

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
