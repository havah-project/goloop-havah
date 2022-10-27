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
		panic("CalcIssueAmount must be called after issue has started")
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
			hvhmodule.StatusCriticalError, "Invalid EcoSystem reward: %d", reward)
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
	rewardRemain := new(big.Int).Add(s.getBigInt(hvhmodule.VarRewardRemain), issueAmount)
	if err := s.setBigInt(hvhmodule.VarRewardRemain, rewardRemain); err != nil {
		return err
	}
	if err := s.setBigInt(hvhmodule.VarRewardTotal, rewardRemain); err != nil {
		return err
	}

	s.logger.Debugf(
		"OnTermStart() end: allPlanet=%d activeUSDT=%d rwdRemain=%d rwdTotal=%d",
		allPlanet, usdtPrice, rewardRemain, rewardRemain,
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
		"height":    height,
		"total":     pr.Total(),
		"remain":    pr.Current(),
		"claimable": claimable,
	}, nil
}

func (s *State) GetRewardPerActivePlanet() *big.Int {
	if s.getBigInt(hvhmodule.VarActivePlanet).Sign() == 0 {
		return hvhmodule.BigIntZero
	}
	return new(big.Int).Div(
		s.getBigInt(hvhmodule.VarRewardTotal),
		s.getBigInt(hvhmodule.VarActivePlanet),
	)
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
		readonly,
		store,
		logger,
	}
}
