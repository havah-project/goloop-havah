package hvhstate

import (
	"math/big"

	"github.com/icon-project/goloop/common/containerdb"
	"github.com/icon-project/goloop/common/log"
	"github.com/icon-project/goloop/common/trie"
	"github.com/icon-project/goloop/common/trie/trie_manager"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/scoreresult"
)

type State struct {
	readonly bool
	store    trie.Mutable
	logger   log.Logger
}

func (s *State) GetValue(key []byte) ([]byte, error) {
	return s.store.Get(key)
}

func (s *State) SetValue(key []byte, value []byte) ([]byte, error) {
	return s.store.Set(key, value)
}

func (s *State) DeleteValue(key []byte) ([]byte, error) {
	return s.store.Delete(key)
}

func (s *State) GetSnapshot() *Snapshot {
	s.store.GetSnapshot()
	return newSnapshotFromImmutable(s.store.GetSnapshot())
}

func (s *State) Reset(ss *Snapshot) error {
	return s.store.Reset(ss.store)
}

func (s *State) getVarDB(key string) *containerdb.VarDB {
	keyBuilder := s.getKeyBuilder(key)
	return containerdb.NewVarDB(s, keyBuilder)
}

func (s *State) getDictDB(key string, depth int) *containerdb.DictDB {
	keyBuilder := s.getKeyBuilder(key)
	return containerdb.NewDictDB(s, depth, keyBuilder)
}

func (s *State) getArrayDB(key string) *containerdb.ArrayDB {
	keyBuilder := s.getKeyBuilder(key)
	return containerdb.NewArrayDB(s, keyBuilder)
}

func (s *State) getKeyBuilder(key string) containerdb.KeyBuilder {
	return containerdb.ToKey(containerdb.HashBuilder, key)
}

func (s *State) GetUSDTPrice() *big.Int {
	varDB := s.getVarDB(hvhmodule.VarUSDTPrice)
	price := varDB.BigInt()
	if price == nil {
		price = hvhmodule.BigIntZero
	}
	return price
}

func (s *State) SetUSDTPrice(price *big.Int) error {
	if price == nil || price.Sign() < 0 {
		return scoreresult.RevertedError.New("Invalid USDTPrice")
	}
	varDB := s.getVarDB(hvhmodule.VarUSDTPrice)
	return varDB.Set(price)
}

func (s *State) GetIssueStart() int64 {
	varDB := s.getVarDB(hvhmodule.VarIssueStart)
	return varDB.Int64()
}

// SetIssueStart writes the issue start height to varDB in ExtensionState
func (s *State) SetIssueStart(curBH, startBH int64) error {
	if startBH < 1 || startBH <= curBH {
		return scoreresult.RevertedError.New("Invalid height")
	}
	varDB := s.getVarDB(hvhmodule.VarIssueStart)
	return varDB.Set(startBH)
}

func (s *State) GetTermPeriod() int64 {
	varDB := s.getVarDB(hvhmodule.VarTermPeriod)
	return varDB.Int64()
}

func (s *State) AddPlanetManager(address module.Address) error {
	if ok, err := s.IsPlanetManager(address); err != nil {
		return err
	} else if ok {
		return scoreresult.RevertedError.New("Duplicate address")
	}
	arrayDB := s.getArrayDB(hvhmodule.ArrayPlanetManager)
	return arrayDB.Put(address)
}

func (s *State) RemovePlanetManager(address module.Address) error {
	if address == nil {
		return scoreresult.RevertedError.New("Invalid address")
	}
	arrayDB := s.getArrayDB(hvhmodule.ArrayPlanetManager)
	size := arrayDB.Size()
	for i := 0; i < size; i++ {
		if address.Equal(arrayDB.Get(i).Address()) {
			pmAddress := arrayDB.Pop().Address()
			if i < size-1 {
				if err := arrayDB.Set(i, pmAddress); err != nil {
					return err
				}
			}
			return nil
		}
	}
	return scoreresult.RevertedError.New("Address not found")
}

func (s *State) IsPlanetManager(address module.Address) (bool, error) {
	if address == nil {
		return false, scoreresult.RevertedError.New("Invalid address")
	}
	arrayDB := s.getArrayDB(hvhmodule.ArrayPlanetManager)
	size := arrayDB.Size()
	for i := 0; i < size; i++ {
		if address.Equal(arrayDB.Get(i).Address()) {
			return true, nil
		}
	}
	return false, nil
}

func (s *State) RegisterPlanet(
	id int64, isPrivate, isCompany bool, owner module.Address, usdt, price *big.Int, height int64,
) error {
	if id < 0 {
		return scoreresult.Errorf(hvhmodule.StatusIllegalArgument, "Invalid id: %d", id)
	}

	// Check if id is available
	planetDictDB := s.getDictDB(hvhmodule.DictPlanet, 1)
	if planetDictDB.Get(id) != nil {
		return scoreresult.Errorf(hvhmodule.StatusIllegalArgument, "Already in use: id=%d", id)
	}

	// Check if too many planets have been registered
	allPlanetVarDB := s.getVarDB(hvhmodule.VarAllPlanet)
	planetCount := allPlanetVarDB.Int64()
	if planetCount >= hvhmodule.MaxPlanetCount {
		return scoreresult.Errorf(hvhmodule.StatusIllegalArgument, "Too many planets: %d", planetCount)
	}

	p := newPlanet(isPrivate, isCompany, owner, usdt, price, height)

	if err := planetDictDB.Set(id, p.Bytes()); err != nil {
		return scoreresult.UnknownFailureError.Wrap(err, "Failed to write to planetDictDB")
	}
	if err := allPlanetVarDB.Set(planetCount + 1); err != nil {
		return scoreresult.UnknownFailureError.Wrap(err, "Failed to write to allPlanetVarDB")
	}
	return nil
}

func (s *State) UnregisterPlanet(id int64) error {
	if id < 0 {
		return scoreresult.Errorf(hvhmodule.StatusIllegalArgument, "Invalid id: %d", id)
	}

	// Check if id exists
	planetDictDB := s.getDictDB(hvhmodule.DictPlanet, 1)
	if planetDictDB.Get(id) == nil {
		return scoreresult.Errorf(hvhmodule.StatusIllegalArgument, "Planet not found: id=%d", id)
	}

	allPlanetVarDB := s.getVarDB(hvhmodule.VarAllPlanet)
	planetCount := allPlanetVarDB.Int64()
	if planetCount < 1 {
		return scoreresult.Errorf(hvhmodule.StatusCriticalError, "Planet state mismatch")
	}

	if err := planetDictDB.Delete(id); err != nil {
		return scoreresult.UnknownFailureError.Wrap(err, "Failed to delete data from planetVarDB")
	}
	if err := allPlanetVarDB.Set(planetCount - 1); err != nil {
		return scoreresult.UnknownFailureError.Wrap(err, "Failed to write to allPlanetVarDB")
	}

	// TODO: Remaining reward and active planet state handling
	return nil
}

func (s *State) SetPlanetOwner(id int64, owner module.Address) error {
	planetDictDB := s.getDictDB(hvhmodule.DictPlanet, 1)
	p, err := s.getPlanet(planetDictDB, id)
	if err != nil {
		return err
	}
	err = p.setOwner(owner)
	if err != nil {
		return err
	}
	if p.isDirty() {
		return planetDictDB.Set(id, p.Bytes())
	}
	return nil
}

func (s *State) GetPlanet(id int64) (*planet, error) {
	dictDB := s.getDictDB(hvhmodule.DictPlanet, 1)
	return s.getPlanet(dictDB, id)
}

func (s *State) getPlanet(dictDB *containerdb.DictDB, id int64) (*planet, error) {
	if err := validatePlanetId(id); err != nil {
		return nil, err
	}
	value := dictDB.Get(id)
	if value == nil {
		return nil, scoreresult.Errorf(hvhmodule.StatusIllegalArgument, "Planet not found: id=%d", id)
	}
	p, err := newPlanetFromBytes(value.Bytes())
	if err != nil {
		return nil, err
	}
	return p, nil
}

func validatePlanetId(id int64) error {
	if id < 0 {
		return scoreresult.Errorf(hvhmodule.StatusIllegalArgument, "Invalid id: %d", id)
	}
	return nil
}

func NewStateFromSnapshot(ss *Snapshot, readonly bool, logger log.Logger) *State {
	store := trie_manager.NewMutableFromImmutable(ss.store)
	return &State{
		readonly,
		store,
		logger,
	}
}
