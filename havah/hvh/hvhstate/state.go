package hvhstate

import (
	"math/big"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/containerdb"
	"github.com/icon-project/goloop/common/errors"
	"github.com/icon-project/goloop/common/log"
	"github.com/icon-project/goloop/common/trie"
	"github.com/icon-project/goloop/common/trie/trie_manager"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/scoreresult"
	"github.com/icon-project/goloop/service/state"
)

type State struct {
	readonly bool
	store    trie.Mutable
	logger   log.Logger

	cachedContainerDBs map[string]interface{}
}

func (s *State) SetLogger(logger log.Logger) {
	s.logger = logger
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
	s.logger.Debugf("Reset() start")
	s.ClearCache()
	err := s.store.Reset(ss.store)
	s.logger.Debugf("Reset() end: err=%v", err)
	return err
}

func (s *State) ClearCache() {
	if len(s.cachedContainerDBs) > 0 {
		s.cachedContainerDBs = make(map[string]interface{})
	}
}

func (s *State) getVarDB(key string) *containerdb.VarDB {
	if db := s.getCachedContainerDB(key); db != nil {
		return db.(*containerdb.VarDB)
	}
	keyBuilder := s.getKeyBuilder(key)
	vdb := containerdb.NewVarDB(s, keyBuilder)
	s.cachedContainerDBs[key] = vdb
	return vdb
}

func (s *State) getDictDB(key string, depth int) *containerdb.DictDB {
	if db := s.getCachedContainerDB(key); db != nil {
		return db.(*containerdb.DictDB)
	}
	keyBuilder := s.getKeyBuilder(key)
	ddb := containerdb.NewDictDB(s, depth, keyBuilder)
	s.cachedContainerDBs[key] = ddb
	return ddb
}

func (s *State) getArrayDB(key string) *containerdb.ArrayDB {
	if db := s.getCachedContainerDB(key); db != nil {
		return db.(*containerdb.ArrayDB)
	}
	keyBuilder := s.getKeyBuilder(key)
	adb := containerdb.NewArrayDB(s, keyBuilder)
	s.cachedContainerDBs[key] = adb
	return adb
}

func (s *State) getCachedContainerDB(key string) interface{} {
	if db, ok := s.cachedContainerDBs[key]; ok {
		return db
	}
	return nil
}

func (s *State) getKeyBuilder(key string) containerdb.KeyBuilder {
	return containerdb.ToKey(containerdb.HashBuilder, key)
}

func (s *State) GetUSDTPrice() *big.Int {
	return s.getBigInt(hvhmodule.VarUSDTPrice)
}

func (s *State) SetUSDTPrice(price *big.Int) error {
	if price == nil || price.Sign() < 0 {
		return scoreresult.RevertedError.New("Invalid USDTPrice")
	}
	return s.setBigInt(hvhmodule.VarUSDTPrice, price)
}

func (s *State) GetActiveUSDTPrice() *big.Int {
	return s.getBigInt(hvhmodule.VarActiveUSDTPrice)
}

func IsIssueStarted(height, issueStart int64) bool {
	return issueStart > 0 && height >= issueStart
}

func GetTermSequenceAndBlockIndex(height, issueStart, termPeriod int64) (int64, int64) {
	blocks := height - issueStart
	if issueStart == 0 || blocks < 0 {
		// Issuing has not been started yet.
		return -1, -1
	}
	termSeq := blocks / termPeriod
	blocks %= termPeriod
	return termSeq, blocks
}

func (s *State) GetIssueStart() int64 {
	return s.getInt64(hvhmodule.VarIssueStart)
}

// SetIssueStart writes the issue start height to varDB in ExtensionState
func (s *State) SetIssueStart(height, newIssueStart int64) error {
	if newIssueStart < 1 || newIssueStart <= height {
		return scoreresult.Errorf(
			hvhmodule.StatusIllegalArgument,
			"Invalid height: cur=%v start=%v", height, newIssueStart)
	}

	varDB := s.getVarDB(hvhmodule.VarIssueStart)
	issueStart := varDB.Int64()
	if IsIssueStarted(height, issueStart) {
		return scoreresult.Errorf(
			hvhmodule.StatusIllegalArgument, "RewardIssue has already begun")
	}

	return varDB.Set(newIssueStart)
}

func (s *State) GetTermPeriod() int64 {
	return s.getInt64OrDefault(hvhmodule.VarTermPeriod, hvhmodule.TermPeriod)
}

func (s *State) GetIssueReductionCycle() int64 {
	return s.getInt64OrDefault(hvhmodule.VarIssueReductionCycle, hvhmodule.IssueReductionCycle)
}

func (s *State) GetIssueReductionRate() *big.Rat {
	return hvhmodule.BigRatIssueReductionRate
}

func (s *State) GetIssueAmount(height, is int64) *big.Int {
	if !IsIssueStarted(height, is) {
		return hvhmodule.BigIntZero
	}
	baseCount := height - is
	termPeriod := s.GetTermPeriod()
	if baseCount%termPeriod != 0 {
		// Coin is issued at the beginning of every term
		return hvhmodule.BigIntZero
	}
	issue, _ := s.GetIssueAmountByTS(baseCount / termPeriod)
	return issue
}

func (s *State) GetIssueAmountByTS(termSeq int64) (*big.Int, bool) {
	issue := s.getBigIntOrDefault(hvhmodule.VarIssueAmount, hvhmodule.BigIntInitIssueAmount)
	reductionCycle := s.GetIssueReductionCycle()
	if termSeq > 0 && termSeq%reductionCycle == 0 {
		reductionRate := s.GetIssueReductionRate()
		reduction := new(big.Int).Mul(issue, reductionRate.Num())
		reduction = reduction.Div(reduction, reductionRate.Denom())
		if reduction.Sign() > 0 {
			issue = new(big.Int).Sub(issue, reduction)
			return issue, true
		}
	}
	return issue, false
}

func (s *State) SetIssueAmount(value *big.Int) error {
	return s.setBigInt(hvhmodule.VarIssueAmount, value)
}

func (s *State) GetHooverBudget() *big.Int {
	return s.getBigIntOrDefault(
		hvhmodule.VarHooverBudget, hvhmodule.BigIntHooverBudget)
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
	rev int,
	id int64, isPrivate, isCompany bool, owner module.Address, usdt, price *big.Int, height int64,
) error {
	s.logger.Debugf(
		"RegisterPlanet() start: height=%d id=%d private=%t company=%t owner=%s usdt=%d price=%d",
		height, id, isPrivate, isCompany, owner, usdt, price)

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

	s.logger.Debugf(
		"RegisterPlanet() end: height=%d planetCount=%d", height, planetCount+1)
	return nil
}

func (s *State) UnregisterPlanet(rev int, id int64) (*big.Int, error) {
	var err error
	if id < 0 {
		return nil, scoreresult.Errorf(
			hvhmodule.StatusIllegalArgument, "Invalid id: %d", id)
	}

	// Check if id exists
	planetDictDB := s.getDictDB(hvhmodule.DictPlanet, 1)
	if planetDictDB.Get(id) == nil {
		return nil, scoreresult.Errorf(
			hvhmodule.StatusIllegalArgument, "Planet not found: id=%d", id)
	}

	allPlanetVarDB := s.getVarDB(hvhmodule.VarAllPlanet)
	planetCount := allPlanetVarDB.Int64()
	if planetCount < 1 {
		return nil, errors.InvalidStateError.Errorf(
			"Planet state mismatch: planetCount=%d", planetCount)
	}

	if err = planetDictDB.Delete(id); err != nil {
		return nil, err
	}
	if err = allPlanetVarDB.Set(planetCount - 1); err != nil {
		return nil, err
	}

	amount := hvhmodule.BigIntZero
	if rev >= hvhmodule.RevisionPlanetIDReuse {
		if pr, err := s.GetPlanetReward(id); err != nil {
			return nil, errors.InvalidStateError.Errorf("PlanetReward not found: id=%d", id)
		} else {
			amount = pr.Current()
			if err = s.addLost(amount); err != nil {
				return nil, err
			}
		}
		planetRewardDictDB := s.getDictDB(hvhmodule.DictPlanetReward, 1)
		if err = planetRewardDictDB.Delete(id); err != nil {
			return nil, err
		}
	}

	return amount, nil
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

func (s *State) getBigInt(key string) *big.Int {
	value := s.getVarDB(key).BigInt()
	if value == nil {
		return hvhmodule.BigIntZero
	}
	return value
}

func (s *State) getBigIntOrDefault(key string, defValue *big.Int) *big.Int {
	value := s.getVarDB(key).BigInt()
	if value == nil {
		return defValue
	}
	return value
}

func (s *State) setBigInt(key string, value *big.Int) error {
	if value == nil {
		return scoreresult.New(hvhmodule.StatusIllegalArgument, "Invalid value")
	}
	return s.getVarDB(key).Set(value)
}

func (s *State) getInt64(key string) int64 {
	return s.getVarDB(key).Int64()
}

func (s *State) getInt64OrDefault(key string, defValue int64) int64 {
	value := s.getVarDB(key).Int64()
	if value <= 0 {
		return defValue
	}
	return value
}

func (s *State) setInt64(key string, value int64) error {
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
	if rewardRemain == nil {
		rewardRemain = new(big.Int).Neg(amount)
	} else {
		rewardRemain = new(big.Int).Sub(rewardRemain, amount)
	}
	if rewardRemain.Sign() < 0 {
		return scoreresult.Errorf(
			hvhmodule.StatusRewardError,
			"Not enough rewardRemain: rewardRemain=%v reward=%v",
			rewardRemain, amount)
	}
	return varDB.Set(rewardRemain)
}

func (s *State) OfferReward(tn, id int64, pr *planetReward, amount, total *big.Int) error {
	s.logger.Debugf(
		"OfferReward() start: tn=%d id=%d amount=%d total=%d",
		tn, id, amount, total)

	if pr == nil {
		return scoreresult.New(
			hvhmodule.StatusIllegalArgument, "Invalid planetReward")
	}
	if err := pr.increment(tn, amount, total); err != nil {
		return err
	}
	err := s.setPlanetReward(id, pr)

	s.logger.Debugf("OfferReward() end: tn=%d id=%d", tn, id)
	return err
}

func (s *State) IncrementWorkingPlanet() error {
	s.logger.Debugf("IncrementWorkingPlanet() start")
	varDB := s.getVarDB(hvhmodule.VarWorkingPlanet)
	planets := varDB.Int64() + 1
	err := varDB.Set(planets)
	s.logger.Debugf("IncrementWorkingPlanet() end: workingPlanet=%d err=%v", planets, err)
	return err
}

func (s *State) IncreaseEcoSystemReward(amount *big.Int) error {
	s.logger.Debugf("IncreaseEcoSystemReward() start: amount=%d", amount)

	varDB := s.getVarDB(hvhmodule.VarEcoReward)
	reward := varDB.BigInt()
	if reward == nil {
		reward = amount
	} else {
		reward = new(big.Int).Add(reward, amount)
	}
	err := varDB.Set(reward)

	s.logger.Debugf("IncreaseEcoSystemReward() end: amount=%d reward=%d", amount, reward)
	return err
}

func (s *State) GetPlanetReward(id int64) (*planetReward, error) {
	var b []byte
	planetRewardDictDB := s.getDictDB(hvhmodule.DictPlanetReward, 1)
	value := planetRewardDictDB.Get(id)
	if value != nil {
		b = value.Bytes()
	}
	return newPlanetRewardFromBytes(b)
}

func (s *State) setPlanetReward(id int64, pr *planetReward) error {
	dictDB := s.getDictDB(hvhmodule.DictPlanetReward, 1)
	return dictDB.Set(id, pr.Bytes())
}

func (s *State) ClaimEcoSystemReward() (*big.Int, error) {
	s.logger.Debugf("ClaimEcoSystemReward() start")

	reward := s.getBigInt(hvhmodule.VarEcoReward)
	if reward == nil || reward.Sign() < 0 {
		return nil, scoreresult.Errorf(
			hvhmodule.StatusInvalidState, "Invalid EcoSystem reward: %d", reward)
	}

	if reward.Sign() > 0 {
		if err := s.setBigInt(hvhmodule.VarEcoReward, hvhmodule.BigIntZero); err != nil {
			return nil, err
		}
	}

	s.logger.Debugf("ClaimEcoSystemReward() end: reward=%d", reward)
	return reward, nil
}

func (s *State) ClaimPlanetReward(id, height int64, owner module.Address) (*big.Int, error) {
	s.logger.Debugf("ClaimPlanetReward() start: id=%d height=%d owner=%s", id, height, owner)

	p, err := s.GetPlanet(id)
	if err != nil {
		return nil, err
	}

	if !owner.Equal(p.Owner()) {
		return nil, scoreresult.AccessDeniedError.Errorf(
			"NoPermission: id=%d owner=%s from=%s", id, p.Owner(), owner)
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

	s.logger.Debugf("ClaimPlanetReward() end: claimableReward=%d", claimableReward)
	return claimableReward, nil
}

func (s *State) calcClaimableReward(height int64, p *Planet, pr *planetReward) (*big.Int, error) {
	claimableReward := pr.Current()
	if !p.IsPrivate() || claimableReward.Sign() == 0 {
		return claimableReward, nil
	}

	num, denom := s.GetPrivateClaimableRate()
	if !validatePrivateClaimableRate(num, denom) {
		return nil, scoreresult.InvalidParameterError.Errorf(
			"InvalidPrivateClaimableRate: num=%d denom=%d", num, denom)
	}

	if num < denom {
		lockedReward := big.NewInt(denom - num)
		lockedReward.Mul(lockedReward, pr.Total())
		lockedReward.Div(lockedReward, big.NewInt(denom))

		claimableReward = new(big.Int).Sub(claimableReward, lockedReward)
		if claimableReward.Sign() < 0 {
			claimableReward.SetInt64(0)
		}
	}

	return claimableReward, nil
}

func (s *State) ClaimMissedReward() (*big.Int, error) {
	activePlanet := s.getBigInt(hvhmodule.VarActivePlanet)
	workingPlanet := s.getBigInt(hvhmodule.VarWorkingPlanet)
	missedPlanet := new(big.Int).Sub(activePlanet, workingPlanet)

	rewardTotal := s.getBigInt(hvhmodule.VarRewardTotal)
	missedReward := rewardTotal

	if activePlanet.Sign() > 0 {
		missedReward = new(big.Int).Div(rewardTotal, activePlanet)
		missedReward = missedReward.Mul(missedReward, missedPlanet)
	}

	rewardRemain := s.getBigInt(hvhmodule.VarRewardRemain)
	rewardRemain = new(big.Int).Sub(rewardRemain, missedReward)
	if rewardRemain.Sign() < 0 {
		return nil, errors.InvalidStateError.Errorf(
			"RewardRemainRemain(remain=%d,missedReward=%d)",
			rewardRemain, missedReward)
	}

	if err := s.setBigInt(hvhmodule.VarRewardRemain, rewardRemain); err != nil {
		return nil, err
	}
	return missedReward, nil
}

func (s *State) OnTermEnd() error {
	s.logger.Debugf("OnTermEnd() start")
	if err := s.setInt64(hvhmodule.VarWorkingPlanet, 0); err != nil {
		return err
	}
	if err := s.setInt64(hvhmodule.VarActivePlanet, 0); err != nil {
		return err
	}
	if err := s.setBigInt(hvhmodule.VarActiveUSDTPrice, hvhmodule.BigIntZero); err != nil {
		return err
	}
	s.logger.Debugf("OnTermEnd() end")
	return nil
}

func (s *State) OnTermStart(issueAmount *big.Int) error {
	s.logger.Debugf("OnTermStart() start: issue=%s", issueAmount)
	allPlanet := s.getInt64(hvhmodule.VarAllPlanet)
	usdtPrice := s.GetUSDTPrice()

	if err := s.setInt64(hvhmodule.VarActivePlanet, allPlanet); err != nil {
		return err
	}
	if err := s.setBigInt(hvhmodule.VarActiveUSDTPrice, usdtPrice); err != nil {
		return err
	}
	oldRewardRemain := s.getBigInt(hvhmodule.VarRewardRemain)
	rewardTotal := new(big.Int).Add(oldRewardRemain, issueAmount)
	if err := s.setBigInt(hvhmodule.VarRewardRemain, rewardTotal); err != nil {
		return err
	}
	if err := s.setBigInt(hvhmodule.VarRewardTotal, rewardTotal); err != nil {
		return err
	}

	s.logger.Debugf(
		"OnTermStart() end: allPlanet=%d activeUSDT=%d issued=%d oldRwdRemain=%d rwdTotal=%d",
		allPlanet, usdtPrice, issueAmount, oldRewardRemain, rewardTotal,
	)
	return nil
}

func (s *State) GetIssueLimit() int64 {
	return s.getInt64OrDefault(hvhmodule.VarIssueLimit, hvhmodule.IssueLimit)
}

func (s *State) GetRewardInfoOf(height, id int64) (map[string]interface{}, error) {
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
		"id":        id,
		"total":     pr.Total(),
		"remain":    pr.Current(),
		"claimable": claimable,
	}, nil
}

func (s *State) GetActivePlanetCountAndReward() (*big.Int, *big.Int) {
	activePlanets := s.getBigInt(hvhmodule.VarActivePlanet)
	rewardPerActivePlanet := hvhmodule.BigIntZero

	if activePlanets.Sign() > 0 {
		rewardPerActivePlanet = new(big.Int).Div(
			s.getBigInt(hvhmodule.VarRewardTotal), activePlanets)
	}

	return activePlanets, rewardPerActivePlanet
}

type StateConfig struct {
	TermPeriod          *common.HexInt64 `json:"termPeriod,omitempty"`          // 43200 in block
	IssueReductionCycle *common.HexInt64 `json:"issueReductionCycle,omitempty"` // 360 in term
	IssueLimit          *common.HexInt64 `json:"issueLimit,omitempty"`

	IssueAmount  *common.HexInt `json:"issueAmount,omitempty"`  // 5M in HVH
	HooverBudget *common.HexInt `json:"hooverBudget,omitempty"` // unit: HVH
	USDTPrice    *common.HexInt `json:"usdtPrice"`              // unit: HVH
}

func (s *State) InitState(cfg *StateConfig) error {
	var err error

	if cfg.TermPeriod != nil {
		if err = s.setInt64(hvhmodule.VarTermPeriod, cfg.TermPeriod.Value); err != nil {
			return err
		}
	}
	if cfg.IssueReductionCycle != nil {
		if err = s.setInt64(hvhmodule.VarIssueReductionCycle, cfg.IssueReductionCycle.Value); err != nil {
			return err
		}
	}
	if cfg.IssueLimit != nil {
		if err = s.setInt64(hvhmodule.VarIssueLimit, cfg.IssueLimit.Value); err != nil {
			return err
		}
	}

	if cfg.IssueAmount != nil {
		if err = s.setBigInt(hvhmodule.VarIssueAmount, cfg.IssueAmount.Value()); err != nil {
			return err
		}
	}
	if cfg.HooverBudget != nil {
		if err = s.setBigInt(hvhmodule.VarHooverBudget, cfg.HooverBudget.Value()); err != nil {
			return err
		}
	}
	if cfg.USDTPrice != nil {
		if err = s.setBigInt(hvhmodule.VarUSDTPrice, cfg.USDTPrice.Value()); err != nil {
			return err
		}
	} else {
		return scoreresult.InvalidParameterError.New("USDTPrice not found")
	}

	s.printInitState()
	return nil
}

func (s *State) printInitState() {
	if s.logger == nil {
		return
	}
	s.logger.Infof("Initial platform configuration\n"+
		"TermPeriod: %d\n"+
		"IssueReductionCycle: %d\n"+
		"IssueAmount: %d\n"+
		"HooverBudget: %d\n"+
		"USDTPrice: %d\n",
		s.GetTermPeriod(),
		s.GetIssueReductionCycle(),
		s.getBigIntOrDefault(hvhmodule.VarIssueAmount, hvhmodule.BigIntInitIssueAmount),
		s.GetHooverBudget(),
		s.GetUSDTPrice())
}

func (s *State) SetPrivateClaimableRate(num, denom int64) error {
	if !validatePrivateClaimableRate(num, denom) {
		return scoreresult.InvalidParameterError.Errorf(
			"InvalidPrivateClaimableRate: num=%d denom=%d", num, denom)
	}
	return s.setInt64(hvhmodule.VarPrivateClaimableRate, num<<16|denom)
}

func (s *State) GetPrivateClaimableRate() (int64, int64) {
	value := s.getInt64OrDefault(
		hvhmodule.VarPrivateClaimableRate, hvhmodule.PrivateClaimableRate)
	num := value >> 16
	denom := value & 0xffff
	return num, denom
}

func (s *State) GetLost() (*big.Int, error) {
	return s.getBigInt(hvhmodule.VarLost), nil
}

func (s *State) DeleteLost() (*big.Int, error) {
	db := s.getVarDB(hvhmodule.VarLost)
	value, err := db.Delete()
	if err != nil {
		return nil, err
	}

	amount := value.BigInt()
	if amount == nil {
		amount = hvhmodule.BigIntZero
	}
	return amount, nil
}

func (s *State) addLost(amount *big.Int) error {
	if amount == nil || amount.Sign() < 0 {
		return scoreresult.Errorf(
			hvhmodule.StatusIllegalArgument, "Invalid amount: %v", amount)
	}
	if amount.Sign() == 0 {
		return nil
	}

	var lost *big.Int
	db := s.getVarDB(hvhmodule.VarLost)
	if lost = db.BigInt(); lost == nil {
		lost = hvhmodule.BigIntZero
	}
	return db.Set(new(big.Int).Add(lost, amount))
}

func (s *State) SetBlockVoteCheckParameters(period, allowance int64) error {
	if period < 0 {
		return scoreresult.InvalidParameterError.Errorf("InvalidArgument(period=%d)", period)
	}
	if allowance < 0 {
		return scoreresult.InvalidParameterError.Errorf("InvalidArgument(allowance=%d)", allowance)
	}

	db := s.getVarDB(hvhmodule.VarBlockVoteCheckPeriod)
	if err := db.Set(period); err != nil {
		return err
	}
	db = s.getVarDB(hvhmodule.VarNonVoteAllowance)
	return db.Set(allowance)
}

func (s *State) GetBlockVoteCheckPeriod() int64 {
	db := s.getVarDB(hvhmodule.VarBlockVoteCheckPeriod)
	if period := db.BigInt(); period != nil {
		return period.Int64()
	}
	return hvhmodule.BlockVoteCheckPeriod
}

func (s *State) GetNonVoteAllowance() int64 {
	db := s.getVarDB(hvhmodule.VarNonVoteAllowance)
	if allowance := db.BigInt(); allowance != nil {
		return allowance.Int64()
	}
	return hvhmodule.NonVoteAllowance
}

func (s *State) RegisterValidator(
	owner module.Address, nodePublicKey []byte, grade Grade, name string, urlPtr *string) error {
	s.logger.Debugf(
		"RegisterValidator() start: owner=%s pubKey=%x grade=%s name=%s urlPtr=%v",
		owner, nodePublicKey, grade, name, urlPtr)

	vi, err := s.registerValidatorInfo(owner, nodePublicKey, grade, name, urlPtr)
	if err != nil {
		return err
	}
	if err = s.registerValidatorStatus(owner); err != nil {
		return err
	}
	if err = s.addNodeToOwnerMap(vi.Address(), owner); err != nil {
		return err
	}
	if err = s.addToValidatorList(grade, owner); err != nil {
		return err
	}

	s.logger.Debugf("RegisterValidator() end: owner=%s", owner)
	return nil
}

func (s *State) registerValidatorInfo(
	owner module.Address, nodePublicKey []byte, grade Grade, name string, urlPtr *string) (*ValidatorInfo, error) {
	if owner == nil {
		return nil, scoreresult.InvalidParameterError.Errorf("InvalidArgument(owner=%s)", owner)
	}
	if grade == GradeNone {
		return nil, scoreresult.InvalidParameterError.Errorf("InvalidArgument(grade=%s)", grade)
	}

	db := s.getDictDB(hvhmodule.DictValidatorInfo, 1)
	if v := db.Get(ToKey(owner)); v != nil {
		return nil, scoreresult.Errorf(
			hvhmodule.StatusDuplicate, "ValidatorInfoAlreadyExists(%s)", owner)
	}

	vi, err := NewValidatorInfo(owner, nodePublicKey, grade, name, urlPtr)
	if err != nil {
		return nil, err
	}
	return vi, db.Set(ToKey(owner), vi.Bytes())
}

func (s *State) registerValidatorStatus(owner module.Address) error {
	db := s.getDictDB(hvhmodule.DictValidatorStatus, 1)
	if v := db.Get(ToKey(owner)); v != nil {
		return scoreresult.Errorf(
			hvhmodule.StatusDuplicate, "ValidatorStatusAlreadyExists(%s)", owner)
	}
	vs := NewValidatorStatus()
	return db.Set(ToKey(owner), vs.Bytes())
}

func (s *State) addNodeToOwnerMap(node, owner module.Address) error {
	key := ToKey(node)
	db := s.getDictDB(hvhmodule.DictNodeToOwner, 1)

	if v := db.Get(key); v != nil {
		return scoreresult.Errorf(
			hvhmodule.StatusDuplicate,
			"NodeAddressAlreadyExists(owner=%s,node=%s)", owner, node)
	}
	return db.Set(key, owner)
}

func (s *State) GetOwnerByNode(node module.Address) (module.Address, error) {
	db := s.getDictDB(hvhmodule.DictNodeToOwner, 1)
	return s.getOwnerByNode(db, node)
}

func (s *State) getOwnerByNode(db *containerdb.DictDB, node module.Address) (module.Address, error) {
	key := ToKey(node)
	v := db.Get(key)
	if v == nil {
		return nil, scoreresult.Errorf(
			hvhmodule.StatusNotFound, "NodeAddressNotFound(%s)", node)
	}
	return v.Address(), nil
}

func (s *State) addToValidatorList(grade Grade, owner module.Address) error {
	if grade != GradeSub && grade != GradeMain {
		return scoreresult.InvalidParameterError.Errorf("InvalidArgument(grade=%s)", grade)
	}
	db := s.getValidatorArrayDB(grade)
	if db == nil {
		return scoreresult.Errorf(
			hvhmodule.StatusNotFound, "ValidatorListNotFound(%s)", grade)
	}
	return db.Put(owner)
}

func (s *State) removeFromValidatorList(grade Grade, owner module.Address) error {
	if grade != GradeSub && grade != GradeMain {
		return scoreresult.InvalidParameterError.Errorf("InvalidArgument(grade=%s)", grade)
	}
	db := s.getValidatorArrayDB(grade)
	if db == nil {
		return scoreresult.Errorf(
			hvhmodule.StatusNotFound, "ValidatorListNotFound(%s)", grade)
	}

	size := db.Size()
	if size > 0 {
		addr := db.Pop().Address()
		if owner.Equal(addr) {
			return nil
		}

		size--
		for i := 0; i < size; i++ {
			if value := db.Get(i); value != nil {
				if owner.Equal(value.Address()) {
					return db.Set(i, addr)
				}
			}
		}
	}

	return scoreresult.Errorf(
		hvhmodule.StatusNotFound,
		"FailedToRemoveFromValidatorList(grade=%s,owner=%s,validators=%d)", grade, owner, size)
}

func (s *State) addToDisqualifiedValidatorList(owner module.Address) error {
	db := s.getArrayDB(hvhmodule.ArrayDisqualifiedValidators)
	return db.Put(owner)
}

func (s *State) getValidatorArrayDB(grade Grade) *containerdb.ArrayDB {
	switch grade {
	case GradeMain:
		return s.getArrayDB(hvhmodule.ArrayMainValidators)
	case GradeSub:
		return s.getArrayDB(hvhmodule.ArraySubValidators)
	default:
		return nil
	}
}

// UnregisterValidator returns nodeAddress of the unregistered validator
func (s *State) UnregisterValidator(owner module.Address) (module.Address, error) {
	s.logger.Debugf("UnregisterValidator() start: owner=%s", owner)
	defer s.logger.Debugf("UnregisterValidator() end: owner=%s", owner)

	if owner == nil || owner.IsContract() {
		return nil, scoreresult.InvalidParameterError.Errorf("InvalidArgument(%s)", owner)
	}

	vsDB := s.getDictDB(hvhmodule.DictValidatorStatus, 1)
	vs, err := s.getValidatorStatus(vsDB, owner)
	if err != nil {
		return nil, err
	}

	if vs.Disqualified() {
		// Already unregistered, so nothing to do
		return nil, nil
	}

	// Turn disqualified flag on
	vs.SetDisqualified()
	key := ToKey(owner)
	if err = vsDB.Set(key, vs.Bytes()); err != nil {
		return nil, err
	}

	// Remove it from validatorList based on its grade
	var vi *ValidatorInfo
	viDB := s.getDictDB(hvhmodule.DictValidatorInfo, 1)
	vi, err = s.getValidatorInfo(viDB, owner)
	if err != nil {
		return nil, err
	}
	err = s.removeFromValidatorList(vi.Grade(), owner)
	if err != nil {
		return nil, err
	}
	if err = s.addToDisqualifiedValidatorList(owner); err != nil {
		return nil, err
	}
	return vi.Address(), nil
}

func (s *State) GetNetworkStatus() (*NetworkStatus, error) {
	db := s.getVarDB(hvhmodule.VarNetworkStatus)
	return s.getNetworkStatus(db)
}

func (s *State) getNetworkStatus(db *containerdb.VarDB) (*NetworkStatus, error) {
	var err error
	var ns *NetworkStatus
	if bs := db.Bytes(); bs != nil {
		ns, err = NewNetworkStatusFromBytes(bs)
	} else {
		ns = NewNetworkStatus()
	}
	return ns, err
}

func (s *State) SetNetworkStatus(ns *NetworkStatus) error {
	if ns == nil {
		return scoreresult.InvalidParameterError.New("InvalidArgument")
	}
	db := s.getVarDB(hvhmodule.VarNetworkStatus)
	return db.Set(ns.Bytes())
}

func (s *State) SetDecentralized() error {
	db := s.getVarDB(hvhmodule.VarNetworkStatus)
	if ns, err := s.getNetworkStatus(db); err == nil {
		ns.SetDecentralized()
		return s.SetNetworkStatus(ns)
	} else {
		return err
	}
}

func (s *State) RenewNetworkStatusOnTermStart() error {
	db := s.getVarDB(hvhmodule.VarNetworkStatus)
	ns, err := s.getNetworkStatus(db)
	if err != nil {
		return err
	}
	if !ns.IsDecentralized() {
		s.logger.Debugf(
			"RenewNetworkStatusOnTermStart() should not be called before decentralization")
		return nil
	}

	dirty := false
	period := s.GetBlockVoteCheckPeriod()
	allowance := s.GetNonVoteAllowance()
	count := s.GetActiveValidatorCount()

	if ns.BlockVoteCheckPeriod() != period {
		if err = ns.SetBlockVoteCheckPeriod(period); err != nil {
			return err
		}
		dirty = true
	}
	if ns.NonVoteAllowance() != allowance {
		if err = ns.SetNonVoteAllowance(allowance); err != nil {
			return err
		}
		dirty = true
	}
	if ns.ActiveValidatorCount() != count {
		if err = ns.SetActiveValidatorCount(count); err != nil {
			return err
		}
		dirty = true
	}

	if dirty {
		return db.Set(ns.Bytes())
	}
	return nil
}

func (s *State) SetValidatorInfo(owner module.Address, values map[string]string) error {
	s.logger.Debugf("SetValidatorInfo() start: owner=%s values=%v", owner, values)

	db := s.getDictDB(hvhmodule.DictValidatorInfo, 1)
	vi, err := s.getValidatorInfo(db, owner)
	if err != nil {
		return err
	}

	for key, value := range values {
		switch key {
		case "name":
			if err = vi.SetName(value); err != nil {
				return err
			}
		case "url":
			if err = vi.SetUrl(&value); err != nil {
				return err
			}
		default:
			// "nodePublicKey" is handled outside of this method
			return scoreresult.InvalidParameterError.Errorf("InvalidArgument(%s)", key)
		}
	}

	s.logger.Debugf("SetValidatorInfo() end: owner=%s values=%v", owner, values)
	return db.Set(ToKey(owner), vi.Bytes())
}

// SetNodePublicKey returns old node address, new node address and error
func (s *State) SetNodePublicKey(owner module.Address, pubKey []byte) (module.Address, module.Address, error) {
	db := s.getDictDB(hvhmodule.DictValidatorInfo, 1)
	vi, err := s.getValidatorInfo(db, owner)
	if err != nil {
		return nil, nil, err
	}
	vs, err := s.GetValidatorStatus(owner)
	if err != nil {
		return nil, nil, err
	}
	if vs.Disqualified() {
		return nil, nil, scoreresult.New(hvhmodule.StatusIllegalArgument, "DisqualifiedValidator")
	}
	oldNode := vi.Address()
	err = vi.SetPublicKey(pubKey)
	if err != nil {
		return nil, nil, scoreresult.InvalidParameterError.Errorf("InvalidArgument(pubKey=%x)", pubKey)
	}

	newNode := vi.Address()
	if oldNode.Equal(newNode) {
		return oldNode, newNode, nil
	}

	if err = s.addNodeToOwnerMap(newNode, owner); err != nil {
		return nil, nil, err
	}
	if err = db.Set(ToKey(owner), vi.Bytes()); err != nil {
		return nil, nil, err
	}
	return oldNode, newNode, nil
}

func (s *State) GetValidatorInfo(owner module.Address) (*ValidatorInfo, error) {
	db := s.getDictDB(hvhmodule.DictValidatorInfo, 1)
	return s.getValidatorInfo(db, owner)
}

func (s *State) getValidatorInfo(db *containerdb.DictDB, owner module.Address) (*ValidatorInfo, error) {
	v := db.Get(ToKey(owner))
	if v == nil {
		return nil, scoreresult.Errorf(
			hvhmodule.StatusNotFound, "ValidatorInfoNotFound(%s)", owner)
	}
	return NewValidatorInfoFromBytes(v.Bytes())
}

func (s *State) GetValidatorStatus(owner module.Address) (*ValidatorStatus, error) {
	db := s.getDictDB(hvhmodule.DictValidatorStatus, 1)
	return s.getValidatorStatus(db, owner)
}

func (s *State) getValidatorStatus(db *containerdb.DictDB, owner module.Address) (*ValidatorStatus, error) {
	v := db.Get(ToKey(owner))
	if v == nil {
		return nil, scoreresult.Errorf(
			hvhmodule.StatusNotFound, "ValidatorStatusNotFound(%s)", owner)
	}
	return NewValidatorStatusFromBytes(v.Bytes())
}

func (s *State) EnableValidator(owner module.Address, calledByGov bool) error {
	db := s.getDictDB(hvhmodule.DictValidatorStatus, 1)
	vs, err := s.getValidatorStatus(db, owner)
	if err != nil {
		return err
	}
	if err = vs.Enable(calledByGov); err != nil {
		return err
	}
	vs.ResetNonVotes()
	return db.Set(ToKey(owner), vs.Bytes())
}

// DisableValidator is used for test
// Do not use this method anywhere else
func (s *State) DisableValidator(owner module.Address) error {
	db := s.getDictDB(hvhmodule.DictValidatorStatus, 1)
	vs, err := s.getValidatorStatus(db, owner)
	if err != nil {
		return err
	}
	if vs.Disabled() {
		return nil
	}
	vs.SetDisabled()
	return db.Set(ToKey(owner), vs.Bytes())
}

func (s *State) GetMainValidators(cc hvhmodule.CallContext, count int) ([]module.Address, error) {
	if count < 0 {
		return nil, scoreresult.InvalidParameterError.Errorf("InvalidArgument(count=%d)", count)
	}
	if count == 0 {
		return nil, nil
	}

	mvDB := s.getArrayDB(hvhmodule.ArrayMainValidators)
	viDB := s.getDictDB(hvhmodule.DictValidatorInfo, 1)

	bs := cc.GetBTPState()
	btx := cc.GetBTPContext()
	size := mvDB.Size()
	validators := make([]module.Address, 0, size)

	for i := 0; i < size; i++ {
		owner := mvDB.Get(i).Address()
		vi, err := s.getValidatorInfo(viDB, owner)
		if err != nil {
			return nil, errors.InvalidStateError.Errorf(
				"MismatchBetweenMainValidatorsAndValidatorInfo(owner=%s)", owner)
		}
		node := vi.Address()

		// Check if validator has a public key for BTP
		if err = bs.CheckPublicKey(btx, node); err != nil {
			continue
		}

		validators = append(validators, node)
		if len(validators) == count {
			break
		}
	}

	return validators, nil
}

// GetValidatorsOf returns a list of validator owner filtered by grade
func (s *State) GetValidatorsOf(gradeFilter GradeFilter) ([]module.Address, error) {
	validatorOwners := make([]module.Address, 0, 20)

	if gradeFilter == GradeFilterMain || gradeFilter == GradeFilterAll {
		db := s.getArrayDB(hvhmodule.ArrayMainValidators)
		size := db.Size()
		for i := 0; i < size; i++ {
			validatorOwners = append(validatorOwners, db.Get(i).Address())
		}
	}
	if gradeFilter == GradeFilterSub || gradeFilter == GradeFilterAll {
		db := s.getArrayDB(hvhmodule.ArraySubValidators)
		size := db.Size()
		for i := 0; i < size; i++ {
			validatorOwners = append(validatorOwners, db.Get(i).Address())
		}
	}

	return validatorOwners, nil
}

func (s *State) GetDisqualifiedValidators() ([]module.Address, error) {
	db := s.getArrayDB(hvhmodule.ArrayDisqualifiedValidators)
	size := db.Size()
	if size == 0 {
		return nil, nil
	}

	owners := make([]module.Address, size)
	for i := 0; i < size; i++ {
		owners[i] = db.Get(i).Address()
	}
	return owners, nil
}

func (s *State) SetActiveValidatorCount(count int64) error {
	if count < 1 {
		return scoreresult.InvalidParameterError.Errorf("InvalidArgument(%d)", count)
	}
	db := s.getVarDB(hvhmodule.VarActiveValidatorCount)
	return db.Set(count)
}

// GetActiveValidatorCount returns the number of validators involved in validating blocks
// Initial value is 0
func (s *State) GetActiveValidatorCount() int64 {
	db := s.getVarDB(hvhmodule.VarActiveValidatorCount)
	return db.Int64()
}

// OnBlockVote returns penalized, owner address and error
func (s *State) OnBlockVote(node module.Address, vote bool) (bool, module.Address, error) {
	ntoDB := s.getDictDB(hvhmodule.DictNodeToOwner, 1)
	viDB := s.getDictDB(hvhmodule.DictValidatorInfo, 1)
	vsDB := s.getDictDB(hvhmodule.DictValidatorStatus, 1)
	ns, err := s.GetNetworkStatus()
	if err != nil {
		return false, nil, err
	}
	nonVoteAllowance := ns.NonVoteAllowance()

	owner, err := s.getOwnerByNode(ntoDB, node)
	if err != nil {
		return false, nil, err
	}
	vi, err := s.getValidatorInfo(viDB, owner)
	if err != nil {
		return false, nil, err
	}
	vs, err := s.getValidatorStatus(vsDB, owner)
	if err != nil {
		return false, nil, err
	}

	if vs.Enabled() {
		if vote {
			vs.ResetNonVotes()
		} else {
			vs.IncrementNonVotes()
		}

		penalized := false
		if vi.Grade() != GradeMain {
			penalized = vs.NonVotes() > nonVoteAllowance
			if penalized {
				vs.SetDisabled()
			}
		}
		err = vsDB.Set(ToKey(owner), vs.Bytes())
		return penalized, owner, err
	}
	return false, owner, nil
}

func (s *State) GetNextActiveValidatorsAndChangeIndex(
	cc hvhmodule.CallContext,
	activeValidators state.ValidatorState, count int) ([]module.Address, error) {
	if count < 0 {
		return nil, scoreresult.InvalidParameterError.Errorf("InvalidArgument(%d)", count)
	}
	if count == 0 {
		return nil, nil
	}

	svDB := s.getArrayDB(hvhmodule.ArraySubValidators)
	size := svDB.Size()
	if size == 0 {
		return nil, nil
	}
	count = min(count, size)

	var err error
	var vi *ValidatorInfo
	var vs *ValidatorStatus
	sviDB := s.getVarDB(hvhmodule.VarSubValidatorsIndex)
	svIndex := int(sviDB.Int64())
	oldSVIndex := svIndex
	if svIndex >= size {
		// If svIndex is invalid, reset the index to 0
		svIndex = 0
	}

	bs := cc.GetBTPState()
	btx := cc.GetBTPContext()
	nextActiveValidators := make([]module.Address, 0, count)
	viDB := s.getDictDB(hvhmodule.DictValidatorInfo, 1)
	vsDB := s.getDictDB(hvhmodule.DictValidatorStatus, 1)

	for i := 0; i < size && len(nextActiveValidators) < count; i++ {
		owner := svDB.Get(svIndex).Address()
		if vi, err = s.getValidatorInfo(viDB, owner); err != nil {
			return nil, err
		}
		node := vi.Address()

		if activeValidators == nil || activeValidators.IndexOf(node) < 0 {
			if vs, err = s.getValidatorStatus(vsDB, owner); err == nil {
				if vs.Enabled() {
					// Check if validator has a public key for BTP
					if err = bs.CheckPublicKey(btx, node); err == nil {
						nextActiveValidators = append(nextActiveValidators, node)
					}
				}
			}
		}

		svIndex = (svIndex + 1) % size
	}

	if oldSVIndex != svIndex {
		if err = sviDB.Set(svIndex); err != nil {
			return nil, err
		}
	}
	return nextActiveValidators, err
}

func (s *State) GetSubValidatorsIndex() int64 {
	sviDB := s.getVarDB(hvhmodule.VarSubValidatorsIndex)
	return sviDB.Int64()
}

func (s *State) IsDecentralizationPossible(rev int) bool {
	if rev < hvhmodule.RevisionDecentralization {
		return false
	}
	validatorCount := int(s.GetActiveValidatorCount())
	if validatorCount < 1 {
		return false
	}

	mvDB := s.getArrayDB(hvhmodule.ArrayMainValidators)
	svDB := s.getArrayDB(hvhmodule.ArraySubValidators)
	return (mvDB.Size() + svDB.Size()) >= validatorCount
}

func (s *State) IsItTimeToCheckBlockVote(blockIndexInTerm int64) bool {
	if ns, err := s.GetNetworkStatus(); err == nil {
		if ns.IsDecentralized() {
			return IsItTimeToCheckBlockVote(blockIndexInTerm, ns.BlockVoteCheckPeriod())
		}
	}
	return false
}

func NewStateFromSnapshot(ss *Snapshot, readonly bool, logger log.Logger) *State {
	store := trie_manager.NewMutableFromImmutable(ss.store)
	return &State{
		readonly:           readonly,
		store:              store,
		logger:             logger,
		cachedContainerDBs: make(map[string]interface{}),
	}
}
