package hvhstate

import (
	"github.com/icon-project/goloop/common/db"
	"github.com/icon-project/goloop/common/merkle"
	"github.com/icon-project/goloop/common/trie"
	"github.com/icon-project/goloop/common/trie/trie_manager"
)

type Snapshot struct {
	store trie.Immutable
}

func (ss *Snapshot) Bytes() []byte {
	return ss.store.Hash()
}

func (ss *Snapshot) Flush() error {
	if sshot, ok := ss.store.(trie.Snapshot); ok {
		return sshot.Flush()
	}
	return nil
}

func (ss *Snapshot) GetValue(key []byte) ([]byte, error) {
	return ss.store.Get(key)
}

func NewSnapshot(dbase db.Database, h []byte) *Snapshot {
	store := trie_manager.NewImmutable(dbase, h)
	return &Snapshot{
		store,
	}
}

func NewSnapshotWithBuilder(builder merkle.Builder, h []byte) *Snapshot {
	dbase := builder.Database()
	store := trie_manager.NewImmutable(dbase, h)
	store.Resolve(builder)
	return newSnapshotFromImmutable(store)
}

func newSnapshotFromImmutable(store trie.Immutable) *Snapshot {
	return &Snapshot{
		store,
	}
}
