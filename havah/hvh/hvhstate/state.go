package hvhstate

import (
	"math/big"

	"github.com/icon-project/goloop/common/containerdb"
	"github.com/icon-project/goloop/common/log"
	"github.com/icon-project/goloop/common/trie"
	"github.com/icon-project/goloop/common/trie/trie_manager"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/service/scoreresult"
)

type State struct {
	readonly bool
	store    trie.Mutable
	logger   log.Logger
}

func (s *State) GetSnapshot() *Snapshot {
	s.store.GetSnapshot()
	return newSnapshotFromImmutable(s.store.GetSnapshot())
}

func (s *State) Reset(ss *Snapshot) error {
	return s.store.Reset(ss.store)
}

func (s *State) getVarDB(key string) *containerdb.VarDB {
	keyBuilder := containerdb.ToKey(containerdb.HashBuilder, key)
	return containerdb.NewVarDB(s.store, keyBuilder)
}

func (s *State) getDictDB(key string, depth int) *containerdb.DictDB {
	keyBuilder := containerdb.ToKey(containerdb.HashBuilder, key)
	return containerdb.NewDictDB(s.store, depth, keyBuilder)
}

func (s *State) getArrayDB(key string) *containerdb.ArrayDB {
	keyBuilder := containerdb.ToKey(containerdb.HashBuilder, key)
	return containerdb.NewArrayDB(s.store, keyBuilder)
}

func (s *State) GetUSDTPrice() *big.Int {
	varDB := s.getVarDB(hvhmodule.VarUSDTPrice)
	return varDB.BigInt()
}

func (s *State) SetUSDTPrice(price *big.Int) error {
	if price.Sign() < 0 {
		return scoreresult.RevertedError.New("Invalid USDTPrice")
	}
	varDB := s.getVarDB(hvhmodule.VarUSDTPrice)
	return varDB.Set(price)
}

func NewStateFromSnapshot(ss *Snapshot, readonly bool, logger log.Logger) *State {
	store := trie_manager.NewMutableFromImmutable(ss.store)
	return &State{
		readonly,
		store,
		logger,
	}
}
