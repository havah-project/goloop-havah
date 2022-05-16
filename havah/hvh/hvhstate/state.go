package hvhstate

import "github.com/icon-project/goloop/common/log"

type State struct {
	readonly bool
}

func (s *State) GetSnapshot() *Snapshot {
	return nil
}

func NewStateFromSnapshot(ss *Snapshot, readonly bool, logger log.Logger) *State {
	return nil
}
