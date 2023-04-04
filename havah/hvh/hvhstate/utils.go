package hvhstate

import (
	"github.com/icon-project/goloop/common/errors"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/scoreresult"
)

func ToKey(o interface{}) string {
	switch o := o.(type) {
	case module.Address:
		return string(o.Bytes())
	case []byte:
		return string(o)
	case string:
		return o
	default:
		panic(errors.Errorf("Unsupported type: %v", o))
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func IsItTimeToCheckBlockVote(blockIndexInTerm, blockVoteCheckPeriod int64) bool {
	return blockVoteCheckPeriod > 0 && blockIndexInTerm%blockVoteCheckPeriod == 0
}

func validatePrivateClaimableRate(num, denom int64) bool {
	if denom <= 0 || denom > 10000 {
		return false
	}
	if num < 0 {
		return false
	}
	if num > denom {
		return false
	}
	return true
}

func validatePlanetId(id int64) error {
	if id < 0 {
		return scoreresult.Errorf(hvhmodule.StatusIllegalArgument, "Invalid id: %d", id)
	}
	return nil
}

func GetTermStartAndIndex(height, issueStart, termPeriod int64) (int64, int64, error){
	if height < issueStart {
		return -1, -1, scoreresult.Errorf(
			hvhmodule.StatusIllegalArgument,
			"height(%d) < issueStart(%d)", height, issueStart)
	}
	termIndex := (height - issueStart) % termPeriod
	termStart := height - termIndex
	return termStart, termIndex, nil
}
