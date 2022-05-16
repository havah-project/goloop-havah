package hvhstate

import (
	"github.com/icon-project/goloop/common/db"
	"github.com/icon-project/goloop/common/merkle"
	"github.com/icon-project/goloop/common/trie"
)

type Snapshot struct {
}

func (ss *Snapshot) Bytes() []byte {
	return nil
}

func (ss *Snapshot) Flush() error {
	return nil
}

func (ss *Snapshot) GetValue(key []byte) ([]byte, error) {
	return nil, nil
}

func (ss *Snapshot) NewState(readonly bool) *State {
	return nil
}

func NewSnapshot(dbase db.Database, h []byte) *Snapshot {
	return nil
}

func NewSnapshotWithBuilder(builder merkle.Builder, h []byte) *Snapshot {
	return nil
}

func newSnapshotFromImmutableForObject(t trie.ImmutableForObject) *Snapshot {
	return nil
}
