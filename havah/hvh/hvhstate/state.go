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
)

type State struct {
	readonly bool
	store    trie.Mutable
	logger   log.Logger

	cachedDictDBs map[string]*containerdb.DictDB
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
		return scoreresult.InvalidParameterError.Errorf("Invalid BlockVoteCheckPeriod: %d", period)
	}
	if allowance < 0 {
		return scoreresult.InvalidParameterError.Errorf("Invalid NonVoteAllowance: %d", allowance)
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

func (s *State) RegisterValidator(owner module.Address, nodePublicKey []byte, grade int, name string) error {
	s.logger.Debugf(
		"RegisterValidator() start: owner=%s grade=%d name=%s nodePublicKey=%x",
		owner, grade, name, nodePublicKey)

	vi, err := s.registerValidatorInfo(owner, nodePublicKey, grade, name)
	if err != nil {
		return err
	}
	if err = s.registerValidatorStatus(owner); err != nil {
		return err
	}
	if err = s.registerNodeAddress(vi.Address(), owner); err != nil {
		return err
	}
	if err = s.addValidatorList(owner); err != nil {
		return err
	}

	s.logger.Debugf("RegisterValidator() end: owner=%s", owner)
	return nil
}

func (s *State) registerValidatorInfo(
	owner module.Address, nodePublicKey []byte, grade int, name string) (*ValidatorInfo, error) {
	if owner == nil {
		return nil, scoreresult.Errorf(hvhmodule.StatusIllegalArgument, "Invalid owner: %v", owner)
	}
	if !isGradeValid(grade) {
		return nil, scoreresult.Errorf(hvhmodule.StatusIllegalArgument, "Invalid grade: %d", grade)
	}
	if len(name) > hvhmodule.MaxValidatorNameLen {
		return nil, scoreresult.Errorf(hvhmodule.StatusIllegalArgument, "Too long name: %s", name)
	}

	db := s.getDictDBFromCache(hvhmodule.DictValidatorInfo)
	if v := db.Get(ToKey(owner)); v != nil {
		return nil, scoreresult.Errorf(
			hvhmodule.StatusDuplicate, "ValidatorInfo already exists: %s", owner)
	}

	vi, err := NewValidatorInfo(owner, nodePublicKey, grade, name)
	if err != nil {
		return nil, err
	}
	return vi, db.Set(ToKey(owner), vi.Bytes())
}

func (s *State) registerValidatorStatus(owner module.Address) error {
	db := s.getDictDBFromCache(hvhmodule.DictValidatorStatus)
	if v := db.Get(ToKey(owner)); v != nil {
		return scoreresult.Errorf(
			hvhmodule.StatusDuplicate, "ValidatorStatus already exists: %s", owner)
	}
	vs := NewValidatorStatus()
	return db.Set(ToKey(owner), vs.Bytes())
}

func (s *State) registerNodeAddress(node, owner module.Address) error {
	key := ToKey(node)
	db := s.getDictDBFromCache(hvhmodule.DictNodeToOwner)

	if v := db.Get(key); v != nil {
		return scoreresult.Errorf(
			hvhmodule.StatusDuplicate,
			"NodeAddress already exists: owner=%s node=%s", owner, node)
	}
	return db.Set(key, owner)
}

func (s *State) getOwnerByNode(db *containerdb.DictDB, node module.Address) (module.Address, error) {
	key := ToKey(node)
	v := db.Get(key)
	if v == nil {
		return nil, scoreresult.Errorf(
			hvhmodule.StatusNotFound, "NodeAddress not found: %s", node)
	}
	return v.Address(), nil
}

func (s *State) addValidatorList(owner module.Address) error {
	db := s.getArrayDB(hvhmodule.ArrayValidators)
	return db.Put(owner)
}

func (s *State) UnregisterValidator(owner module.Address) error {
	s.logger.Debugf("UnregisterValidator() start: owner=%s", owner)
	defer s.logger.Debugf("UnregisterValidator() end: owner=%s", owner)

	if owner == nil {
		return scoreresult.Errorf(hvhmodule.StatusIllegalArgument, "Invalid owner: %v", owner)
	}

	key := ToKey(owner)
	db := s.getDictDBFromCache(hvhmodule.DictValidatorStatus)
	v := db.Get(key)
	if v == nil {
		return scoreresult.Errorf(
			hvhmodule.StatusNotFound, "ValidatorStatus not found: %s", owner)
	}
	bs := v.Bytes()
	if bs == nil {
		return scoreresult.Errorf(
			hvhmodule.StatusNotFound, "ValidatorStatus not found: %s", owner)
	}

	vs, err := NewValidatorStatusFromBytes(bs)
	if err != nil {
		return errors.InvalidStateError.Wrapf(
			err, "Failed to create a ValidatorStatus from bytes: %s", owner)
	}

	vs.SetDisqualified()
	return db.Set(key, vs.Bytes())
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
	db := s.getVarDB(hvhmodule.VarNetworkStatus)
	return db.Set(ns.Bytes())
}

func (s *State) RenewNetworkStatusOnTermStart() error {
	db := s.getVarDB(hvhmodule.VarNetworkStatus)
	ns, err := s.getNetworkStatus(db)
	if err != nil {
		return err
	}

	dirty := false
	period := s.GetBlockVoteCheckPeriod()
	allowance := s.GetNonVoteAllowance()

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

	if dirty {
		err = db.Set(ns.Bytes())
	}
	return err
}

func (s *State) SetValidatorInfo(owner module.Address, name, url string) error {
	db := s.getDictDBFromCache(hvhmodule.DictValidatorInfo)
	vi, err := s.getValidatorInfo(db, owner)
	if err != nil {
		return err
	}
	if err = vi.SetName(name); err != nil {
		return err
	}
	if err = vi.SetUrl(url); err != nil {
		return err
	}
	return db.Set(ToKey(owner), vi.Bytes())
}

func (s *State) GetValidatorInfo(owner module.Address) (*ValidatorInfo, error) {
	db := s.getDictDBFromCache(hvhmodule.DictValidatorInfo)
	return s.getValidatorInfo(db, owner)
}

func (s *State) GetValidatorInfos(owners []module.Address) ([]*ValidatorInfo, error) {
	vis := make([]*ValidatorInfo, len(owners))
	db := s.getDictDBFromCache(hvhmodule.DictValidatorInfo)
	for i, owner := range owners {
		if vi, err := s.getValidatorInfo(db, owner); err == nil {
			vis[i] = vi
		} else {
			return nil, err
		}
	}
	return vis, nil
}

func (s *State) getValidatorInfo(db *containerdb.DictDB, owner module.Address) (*ValidatorInfo, error) {
	key := ToKey(owner)
	v := db.Get(key)
	if v == nil {
		return nil, scoreresult.Errorf(
			hvhmodule.StatusNotFound, "ValidatorInfo not found: %s", owner)
	}
	vi, err := NewValidatorInfoFromBytes(v.Bytes())
	return vi, err
}

func (s *State) GetValidatorStatus(owner module.Address) (*ValidatorStatus, error) {
	db := s.getDictDBFromCache(hvhmodule.DictValidatorStatus)
	return s.getValidatorStatus(db, owner)
}

func (s *State) getValidatorStatus(db *containerdb.DictDB, owner module.Address) (*ValidatorStatus, error) {
	v := db.Get(ToKey(owner))
	if v == nil {
		return nil, scoreresult.Errorf(
			hvhmodule.StatusNotFound, "ValidatorStatus not found: %s", owner)
	}
	return NewValidatorStatusFromBytes(v.Bytes())
}

func (s *State) EnableValidator(owner module.Address, calledByGov bool) error {
	key := ToKey(owner)
	db := s.getDictDBFromCache(hvhmodule.DictValidatorInfo)
	v := db.Get(key)
	if v == nil {
		return scoreresult.Errorf(
			hvhmodule.StatusNotFound, "ValidatorStatus not found: owner=%s", owner)
	}
	bs := v.Bytes()
	if bs == nil || len(bs) == 0 {
		return errors.InvalidStateError.Errorf("ValidatorStatus is broken: owner=%s", owner)
	}
	vs, err := NewValidatorStatusFromBytes(bs)
	if err != nil {
		return errors.InvalidStateError.Wrapf(err, "ValidatorStatus is broken: owner=%s", owner)
	}
	if err = vs.Enable(calledByGov); err != nil {
		return err
	}
	return db.Set(key, vs.Bytes())
}

// DisableValidator is called on imposing nonVotePenalty
func (s *State) DisableValidator(owner module.Address) error {
	db := s.getDictDBFromCache(hvhmodule.DictValidatorStatus)
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

func (s *State) SetNodePublicKey(owner module.Address, publicKey []byte) error {
	if owner == nil {
		return scoreresult.Errorf(hvhmodule.StatusIllegalArgument, "Invalid owner")
	}
	if publicKey == nil || len(publicKey) == 0 {
		return scoreresult.Errorf(hvhmodule.StatusIllegalArgument, "Invalid publicKey")
	}
	node, err := s.setNodePublicKey(owner, publicKey)
	if err == nil {
		err = s.registerNodeAddress(node, owner)
	}
	return err
}

func (s *State) setNodePublicKey(owner module.Address, publicKey []byte) (module.Address, error) {
	db := s.getDictDBFromCache(hvhmodule.DictValidatorInfo)
	vi, err := s.getValidatorInfo(db, owner)
	if err != nil {
		return nil, err
	}
	if err = vi.SetPublicKey(publicKey); err != nil {
		return nil, scoreresult.Wrapf(
			err, hvhmodule.StatusIllegalArgument, "Invalid publicKey")
	}
	if err = db.Set(ToKey(owner), vi.Bytes()); err != nil {
		return nil, err
	}
	return vi.Address(), nil
}

func (s *State) GetAvailableValidators() ([]module.Address, error) {
	vlDB := s.getArrayDB(hvhmodule.ArrayValidators)
	vsDB := s.getDictDBFromCache(hvhmodule.DictValidatorStatus)
	size := vlDB.Size()
	validators := make([]module.Address, 0, size)

	for i := 0; i < size; i++ {
		owner := vlDB.Get(i).Address()
		vs, err := s.getValidatorStatus(vsDB, owner)
		if err != nil {
			return nil, errors.InvalidStateError.Errorf("Mismatch between validatorList and validatorStatus")
		}
		if vs.Enabled() {
			validators = append(validators, owner)
		}
	}

	return validators, nil
}

func (s *State) SetValidatorCount(count int) error {
	if count < 1 {
		return scoreresult.Errorf(hvhmodule.StatusIllegalArgument, "Invalid validator count: %d", count)
	}
	db := s.getVarDB(hvhmodule.VarValidatorCount)
	return db.Set(count)
}

// GetValidatorCount returns the number of validators involved in validating blocks
// Initial value is 0
func (s *State) GetValidatorCount() int {
	db := s.getVarDB(hvhmodule.VarValidatorCount)
	return int(db.Int64())
}

// SetStandbyValidatorOwners initialize standby validator owner set at the beginning of each term
// This set will remain unchanged until the current term is over.
func (s *State) SetStandbyValidatorOwners(owners *AddressSet) error {
	var err error
	db := s.getVarDB(hvhmodule.VarStandbyValidators)
	if owners == nil || owners.Len() == 0 {
		_, err = db.Delete()
		return err
	}
	return db.Set(owners.Bytes())
}

func (s *State) OnBlockVote(node module.Address, vote bool) (bool, error) {
	ntoDB := s.getDictDBFromCache(hvhmodule.DictNodeToOwner)
	viDB := s.getDictDBFromCache(hvhmodule.DictValidatorInfo)
	vsDB := s.getDictDBFromCache(hvhmodule.DictValidatorStatus)
	ns, err := s.GetNetworkStatus()
	if err != nil {
		return false, err
	}
	nonVoteAllowance := ns.NonVoteAllowance()

	owner, err := s.getOwnerByNode(ntoDB, node)
	if err != nil {
		return false, err
	}
	vi, err := s.getValidatorInfo(viDB, owner)
	if err != nil {
		return false, err
	}
	vs, err := s.getValidatorStatus(vsDB, owner)
	if err != nil {
		return false, err
	}

	if vote {
		vs.ResetNonVotes()
	} else {
		vs.IncrementNonVotes()
	}

	penalized := false
	if vi.Grade() != GradeMain {
		penalized = vs.NonVotes() > nonVoteAllowance
		if penalized {
			vs.ResetNonVotes()
			vs.SetDisabled()
		}
	}
	err = vsDB.Set(ToKey(owner), vs.Bytes())
	return penalized, err
}

func (s *State) GetNextActiveValidatorsAndChangeIndex(count int) ([]module.Address, error) {
	if count == 0 {
		return nil, nil
	}

	sviDB := s.getVarDB(hvhmodule.VarStandbyValidatorsIndex)
	index := int(sviDB.Int64())
	if index < 0 {
		// No available standby validatorSet
		return nil, nil
	}
	orgIndex := index

	standbyValidatorSet, err := s.getStandbyValidatorSet()
	if err != nil {
		if status, ok := scoreresult.StatusOf(err); ok {
			if status == hvhmodule.StatusNotFound {
				return nil, nil
			}
		}
		return nil, err
	}

	newActiveValidators := make([]module.Address, 0, count)
	viDB := s.getDictDBFromCache(hvhmodule.DictValidatorInfo)
	vsDB := s.getDictDBFromCache(hvhmodule.DictValidatorStatus)
	size := standbyValidatorSet.Len()

	for ; index < size; index++ {
		owner := standbyValidatorSet.Get(index)
		if vs, err := s.getValidatorStatus(vsDB, owner); err == nil {
			if vs.Enabled() {
				if vi, err := s.getValidatorInfo(viDB, owner); err == nil {
					newActiveValidators = append(newActiveValidators, vi.Address())
					if len(newActiveValidators) == count {
						// index points out the next standby validator
						index++
						break
					}
				}
			}
		}
	}

	if index != orgIndex {
		if index == size {
			index = -1
		}
		if err = sviDB.Set(index); err != nil {
			return nil, err
		}
	}
	return newActiveValidators, err
}

func (s *State) getStandbyValidatorSet() (*AddressSet, error) {
	db := s.getVarDB(hvhmodule.VarStandbyValidators)
	bs := db.Bytes()
	if bs == nil {
		return nil, scoreresult.Errorf(
			hvhmodule.StatusNotFound, "standbyValidatorSet not found")
	}
	return NewAddressSetFromBytes(bs)
}

func (s *State) IsDecentralizationPossible(rev int) bool {
	if rev < hvhmodule.RevisionDecentralization {
		return false
	}
	validatorCount := s.GetValidatorCount()
	if validatorCount <= 0 {
		return false
	}
	validators, err := s.GetAvailableValidators()
	if err != nil {
		return false
	}
	if len(validators) < validatorCount {
		return false
	}
	return true
}

func (s *State) getDictDBFromCache(key string) *containerdb.DictDB {
	db, ok := s.cachedDictDBs[key]
	if !ok {
		db = s.getDictDB(key, 1)
		s.cachedDictDBs[key] = db
	}
	return db
}

func (s *State) IsItTimeToCheckBlockVote(blockIndexInTerm int64) bool {
	if ns, err := s.GetNetworkStatus(); err == nil {
		if ns.IsDecentralized() {
			return IsItTimeToCheckBlockVote(blockIndexInTerm, ns.BlockVoteCheckPeriod())
		}
	}
	return false
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

func NewStateFromSnapshot(ss *Snapshot, readonly bool, logger log.Logger) *State {
	store := trie_manager.NewMutableFromImmutable(ss.store)
	return &State{
		readonly:      readonly,
		store:         store,
		logger:        logger,
		cachedDictDBs: make(map[string]*containerdb.DictDB),
	}
}
