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
	value := varDB.Int64()
	if value <= 0 {
		value = hvhmodule.TermPeriod
	}
	return value
}

func (s *State) GetIssueReductionCycle() int64 {
	varDB := s.getVarDB(hvhmodule.VarIssueReductionCycle)
	value := varDB.Int64()
	if value <= 0 {
		value = hvhmodule.ReductionCycle
	}
	return value
}

func (s *State) GetIssueReductionRate() *big.Rat {
	return hvhmodule.BigRatIssueReductionRate
}

// GetIssueAmount returns the amount of coins which are issued during one term
func (s *State) GetIssueAmount() *big.Int {
	value := s.GetBigInt(hvhmodule.VarIssueAmount)
	if value.Sign() <= 0 {
		value.Set(hvhmodule.BigIntInitIssueAmount)
	}
	return value
}

func (s *State) GetTermSequence(height int64) int64 {
	issueStart := s.GetIssueStart()
	if issueStart == 0 {
		return -1
	}
	termPeriod := s.GetTermPeriod()
	return (height - issueStart) / termPeriod
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

	if err := s.setPlanet(planetDictDB, id, p); err != nil {
		return err
	}
	if err := allPlanetVarDB.Set(planetCount + 1); err != nil {
		return err
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
		return err
	}
	if err := allPlanetVarDB.Set(planetCount - 1); err != nil {
		return err
	}

	// TODO: Remaining reward and active Planet state handling
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
	return s.setPlanet(planetDictDB, id, p)
}

func (s *State) GetPlanet(id int64) (*Planet, error) {
	dictDB := s.getDictDB(hvhmodule.DictPlanet, 1)
	return s.getPlanet(dictDB, id)
}

func (s *State) getPlanet(dictDB *containerdb.DictDB, id int64) (*Planet, error) {
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

func (s *State) setPlanet(dictDB *containerdb.DictDB, id int64, p *Planet) error {
	return dictDB.Set(id, p.Bytes())
}

/*
func (s *State) ReportPlanetWork(id, height int64) (*big.Int, *big.Int, error) {
	p, err := s.GetPlanet(id)
	if err != nil {
		return nil, nil, err
	}

	issueStart := s.GetIssueStart()
	termPeriod := s.GetTermPeriod()
	termSequence := (height - issueStart) / termPeriod
	termStart := termSequence*termPeriod + issueStart

	if p.height >= termStart {
		// If a Planet is registered in this term, ignore its work report
		return nil, nil, nil
	}

	reward := new(big.Int).Div(
		s.GetBigInt(hvhmodule.VarRewardTotal),
		s.GetBigInt(hvhmodule.VarActivePlanet))
	rewardWithHoover := reward

	if err = s.decreaseRewardRemain(reward); err != nil {
		// Not enough rewardRemain
		return nil, nil, err
	}

	pr, err := s.GetPlanetReward(id)
	if err != nil {
		return nil, nil, err
	}

	// hooverLimit = planetReward.total + reward - Planet.price
	hooverLimit := pr.Total()
	hooverLimit.Add(hooverLimit, reward)
	hooverLimit.Sub(hooverLimit, p.Price())
	if hooverLimit.Sign() > 0 {
		hooverGuide := s.calcHooverGuide(p)

		// if reward < hooverGuide
		if reward.Cmp(hooverGuide) < 0 {
			hooverRequest := new(big.Int).Sub(hooverGuide, reward)
			// if hooverRequest > hooverLimit
			if hooverRequest.Cmp(hooverLimit) > 0 {
				hooverRequest = hooverLimit
			}
			rewardWithHoover = new(big.Int).Add(reward, hooverRequest)
		}
	}
	return s.offerReward(termSequence+1, id, rewardWithHoover)
}
*/

func (s *State) GetBigInt(key string) *big.Int {
	return s.getVarDB(key).BigInt()
}

func (s *State) SetBigInt(key string, value *big.Int) error {
	if value == nil {
		return scoreresult.New(hvhmodule.StatusIllegalArgument, "Invalid value")
	}
	return s.getVarDB(key).Set(value)
}

func (s *State) GetInt64(key string) int64 {
	return s.getVarDB(key).Int64()
}

func (s *State) SetInt64(key string, value int64) error {
	return s.getVarDB(key).Set(value)
}

func (s *State) DecreaseRewardRemain(amount *big.Int) error {
	if amount == nil || amount.Sign() < 0 {
		return scoreresult.Errorf(hvhmodule.StatusIllegalArgument, "Invalid amount: %v", amount)
	}
	if amount.Sign() == 0 {
		// Nothing to do
		return nil
	}
	varDB := s.getVarDB(hvhmodule.VarRewardRemain)
	rewardRemain := varDB.BigInt()
	rewardRemain.Sub(rewardRemain, amount)
	if rewardRemain.Sign() < 0 {
		return scoreresult.Errorf(
			hvhmodule.StatusRewardError,
			"Not enough rewardRemain: rewardRemain=%v reward=%v",
			rewardRemain, amount)
	}
	return varDB.Set(rewardRemain)
}

func (s *State) OfferReward(tn, id int64, pr *planetReward, amount *big.Int) error {
	if pr == nil {
		return scoreresult.New(
			hvhmodule.StatusIllegalArgument, "Invalid planetReward")
	}
	if err := pr.increment(tn, amount); err != nil {
		return err
	}
	return s.setPlanetReward(id, pr)
}

func (s *State) IncreaseEcoSystemReward(amount *big.Int) error {
	varDB := s.getVarDB(hvhmodule.VarEcoReward)
	reward := varDB.BigInt()
	return varDB.Set(reward.Add(reward, amount))
}

func (s *State) GetPlanetReward(id int64) (*planetReward, error) {
	planetRewardDictDB := s.getDictDB(hvhmodule.DictPlanetReward, 1)
	b := planetRewardDictDB.Get(id).Bytes()
	return newPlanetRewardFromBytes(b)
}

func (s *State) setPlanetReward(id int64, pr *planetReward) error {
	dictDB := s.getDictDB(hvhmodule.DictPlanetReward, 1)
	return dictDB.Set(id, pr.Bytes())
}

func (s *State) ClaimEcoSystemReward() (*big.Int, error) {
	reward := s.GetBigInt(hvhmodule.VarEcoReward)
	if reward == nil || reward.Sign() < 0 {
		return nil, scoreresult.Errorf(
			hvhmodule.StatusCriticalError, "Invalid EcoSystem reward: %v", reward)
	}

	if reward.Sign() > 0 {
		if err := s.SetBigInt(hvhmodule.VarEcoReward, new(big.Int)); err != nil {
			return nil, err
		}
	}

	return reward, nil
}

func (s *State) ClaimPlanetReward(id, height int64, owner module.Address) (*big.Int, error) {
	p, err := s.GetPlanet(id)
	if err != nil {
		return nil, err
	}

	if !owner.Equal(p.Owner()) {
		return nil, scoreresult.AccessDeniedError.Errorf(
			"Invalid owner: id=%d owner=%s from=%s", id, p.Owner(), owner)
	}
	if p.IsCompany() {
		return nil, scoreresult.Errorf(
			hvhmodule.StatusRewardError,
			"Claim is not allowed for company Planet")
	}

	pr, err := s.GetPlanetReward(id)
	if err != nil {
		return nil, err
	}

	claimableReward, err := s.calcClaimableReward(height, p, pr)
	if err != nil {
		return nil, err
	}

	if claimableReward.Sign() > 0 {
		if err = pr.claim(claimableReward); err != nil {
			return nil, err
		}
		if err = s.setPlanetReward(id, pr); err != nil {
			return nil, err
		}
	}
	return claimableReward, nil
}

func (s *State) calcClaimableReward(height int64, p *Planet, pr *planetReward) (*big.Int, error) {
	claimableReward := pr.Current()
	if !p.IsPrivate() || claimableReward.Sign() == 0 {
		return claimableReward, nil
	}

	termPeriod := s.GetTermPeriod()
	privateLockupTerm := s.GetInt64(hvhmodule.VarPrivateLockup)

	// All rewards have been locked
	lockupTerm := (height - p.Height() - 1) / termPeriod
	if lockupTerm < privateLockupTerm {
		return new(big.Int), nil
	}

	releaseCycle := (lockupTerm-privateLockupTerm)/hvhmodule.DayPerMonth + 1

	if releaseCycle < hvhmodule.MaxPrivateReleaseCycle {
		lockedReward := big.NewInt(hvhmodule.MaxPrivateReleaseCycle - releaseCycle)
		lockedReward.Mul(lockedReward, pr.Total())
		lockedReward.Div(lockedReward, big.NewInt(hvhmodule.MaxPrivateReleaseCycle))

		claimableReward.Sub(claimableReward, lockedReward)
		if claimableReward.Sign() < 0 {
			claimableReward.SetInt64(0)
		}
	}

	return claimableReward, nil
}

func (s *State) OnTermStart(termSeq int64, issueAmount *big.Int) error {
	var err error
	// unit: term
	// 0 means that there is no issueLimit
	issueLimit := s.getIssueLimit()

	if termSeq > 0 && termSeq <= issueLimit {
		if err = s.SetInt64(hvhmodule.VarWorkingPlanet, 0); err != nil {
			return err
		}
		if err = s.SetInt64(hvhmodule.VarActivePlanet, 0); err != nil {
			return err
		}
		if err = s.SetBigInt(hvhmodule.VarActiveUSDTPrice, new(big.Int)); err != nil {
			return err
		}
	}

	if termSeq < issueLimit {
		if err = s.SetInt64(hvhmodule.VarActivePlanet, s.GetInt64(hvhmodule.VarAllPlanet)); err != nil {
			return err
		}
		if err = s.SetBigInt(hvhmodule.VarActiveUSDTPrice, s.GetUSDTPrice()); err != nil {
			return err
		}
		rewardRemain := new(big.Int).Add(s.GetBigInt(hvhmodule.VarRewardRemain), issueAmount)
		if err = s.SetBigInt(hvhmodule.VarRewardRemain, rewardRemain); err != nil {
			return err
		}
		if err = s.SetBigInt(hvhmodule.VarRewardTotal, rewardRemain); err != nil {
			return err
		}
	}
	return nil
}

func (s *State) getIssueLimit() int64 {
	issueLimit := s.GetInt64(hvhmodule.VarIssueLimit)
	if issueLimit <= 0 {
		issueLimit = hvhmodule.IssueLimit
	}
	return issueLimit
}

func (s *State) GetRewardInfo(height, id int64) (map[string]interface{}, error) {
	pr, err := s.GetPlanetReward(id)
	if err != nil {
		return nil, err
	}
	p, err := s.GetPlanet(id)
	if err != nil {
		return nil, err
	}

	claimable, err := s.calcClaimableReward(height, p, pr)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"height":    height,
		"total":     pr.Total(),
		"remain":    pr.Current(),
		"claimable": claimable,
	}, nil
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