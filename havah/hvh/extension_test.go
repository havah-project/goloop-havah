package hvh

import (
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/db"
	"github.com/icon-project/goloop/havah/hvh/hvhstate"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/contract"
	"github.com/icon-project/goloop/service/state"
	"github.com/icon-project/goloop/service/trace"
)

var (
	SystemTreasury = common.MustNewAddressFromString("hx1000000000000000000000000000000000000000")
)

type mockAccount struct {
	state.AccountState
	contract bool
	balance  *big.Int
	store    map[string][]byte
}

func (as *mockAccount) GetBalance() *big.Int {
	return as.balance
}

func (as *mockAccount) SetBalance(balance *big.Int) {
	as.balance = balance
}

func (as *mockAccount) GetValue(k []byte) ([]byte, error) {
	return as.store[string(k)], nil
}

func (as *mockAccount) SetValue(k, v []byte) ([]byte, error) {
	ks := string(k)
	old := as.store[ks]
	as.store[ks] = v
	return old, nil
}

func (as *mockAccount) DeleteValue(k []byte) ([]byte, error) {
	ks := string(k)
	old := as.store[ks]
	delete(as.store, ks)
	return old, nil
}

var (
	EcoSystemIDStr       = string(hvhmodule.EcoSystem.ID())
	HooverFundIDStr      = string(hvhmodule.HooverFund.ID())
	SustainableFundIDStr = string(hvhmodule.SustainableFund.ID())
	CompanyTreasuryIDStr = string(hvhmodule.CompanyTreasury.ID())
	ServiceTreasuryIDStr = string(hvhmodule.ServiceTreasury.ID())
)

func newMockAccount(id []byte) *mockAccount {
	ids := string(id)
	switch ids {
	case EcoSystemIDStr, HooverFundIDStr, SustainableFundIDStr, CompanyTreasuryIDStr, ServiceTreasuryIDStr, state.SystemIDStr:
		return &mockAccount{
			contract: true,
			balance:  hvhmodule.BigIntZero,
			store:    make(map[string][]byte),
		}
	default:
		return &mockAccount{
			contract: false,
			balance:  hvhmodule.BigIntZero,
			store:    nil,
		}
	}
}

type mockCallContext struct {
	contract.CallContext
	height   int64
	accounts map[string]*mockAccount
}

func (cc *mockCallContext) BlockHeight() int64 {
	return cc.height
}

func (cc *mockCallContext) GetAccountState(id []byte) state.AccountState {
	ids := string(id)
	if as, ok := cc.accounts[ids]; ok {
		return as
	} else {
		as = newMockAccount(id)
		cc.accounts[ids] = as
		return as
	}
}

func (cc *mockCallContext) OnEvent(addr module.Address, indexed, data [][]byte) {
	// TODO print something?
	return
}

func (cc *mockCallContext) Treasury() module.Address {
	return SystemTreasury
}

func (cc *mockCallContext) GetBalance(address module.Address) *big.Int {
	as := cc.GetAccountState(address.ID())
	return as.GetBalance()
}

func (cc *mockCallContext) SetBalance(address module.Address, amount *big.Int) {
	as := cc.GetAccountState(address.ID())
	as.SetBalance(amount)
}

func (cc *mockCallContext) FrameLogger() *trace.Logger {
	return nil
}

func newMockCallContext() *mockCallContext {
	return &mockCallContext{
		accounts: make(map[string]*mockAccount),
	}
}

func newMockExtensionState(t *testing.T, cfg *PlatformConfig) *ExtensionStateImpl {
	dbase := db.NewMapDB()
	ess := NewExtensionSnapshot(dbase, nil)
	es := ess.NewState(false)
	esi := es.(*ExtensionStateImpl)

	err := esi.InitPlatformConfig(cfg)
	assert.NoError(t, err)

	return esi
}

func toUSDT(value int64) *big.Int {
	usdt := big.NewInt(value)
	return usdt.Mul(usdt, hvhmodule.BigIntUSDTDecimal)
}

func toHVH(value int64) *big.Int {
	hvh := big.NewInt(value)
	return hvh.Mul(hvh, hvhmodule.BigIntCoinDecimal)
}

func newSimplePlatformConfig(termPeriod, usdtPrice int64) *PlatformConfig {
	return &PlatformConfig{
		StateConfig: hvhstate.StateConfig{
			TermPeriod: &common.HexInt64{Value: termPeriod},
			USDTPrice:  new(common.HexInt).SetValue(toHVH(usdtPrice)),
		},
	}
}

func newMockContextAndExtensionState(t *testing.T, cfg *PlatformConfig) (*mockCallContext, *ExtensionStateImpl) {
	cc := newMockCallContext()
	es := newMockExtensionState(t, cfg)
	assert.Equal(t, cfg.TermPeriod.Value, es.GetTermPeriod())
	return cc, es
}

func goByHeight(
	t *testing.T, targetHeight int64,
	es *ExtensionStateImpl, mcc *mockCallContext, from module.Address) {
	count := targetHeight - mcc.height
	goByCount(t, count, es, mcc, from)
}

func goByCount(t *testing.T, count int64,
	es *ExtensionStateImpl, mcc *mockCallContext, from module.Address) {
	cc := NewCallContext(mcc, from)
	is := es.GetIssueStart()
	termPeriod := es.GetTermPeriod()

	for i := int64(0); i < count; i++ {
		mcc.height++
		if hvhstate.IsIssueStarted(mcc.height, is) {
			_, blockIndex := hvhstate.GetTermSequenceAndBlockIndex(mcc.height, is, termPeriod)
			if blockIndex == 0 {
				// A baseTx is created only at the beginning of each term
				data := es.NewBaseTransactionData(mcc.height, is)
				bs, err := json.Marshal(data)
				assert.NoError(t, err)

				err = es.OnBaseTx(cc, bs)
				assert.NoError(t, err)
			}
		}
	}
}

func checkRewardInfo(t *testing.T,
	jso map[string]interface{},
	height int64, expTotal, expRemain, expClaimable *big.Int) {
	assert.Equal(t, height, jso["height"].(int64))
	total, remain, claimable := getRewardInfoOf(jso)
	assert.Zero(t, expTotal.Cmp(total))
	assert.Zero(t, expRemain.Cmp(remain))
	assert.Zero(t, expClaimable.Cmp(claimable))
}

// getRewardInfoOf() returns total, remain and claimable values
func getRewardInfoOf(jso map[string]interface{}) (*big.Int, *big.Int, *big.Int) {
	return jso["total"].(*big.Int), jso["remain"].(*big.Int), jso["claimable"].(*big.Int)
}

func getPrivateClaimableRate(jso map[string]interface{}) (int64, int64) {
	return jso["numerator"].(int64), jso["denominator"].(int64)
}

func calcNewIssueAmount(issueAmount *big.Int, rate *big.Rat) *big.Int {
	reduction := new(big.Int).Set(issueAmount)
	reduction.Mul(reduction, rate.Num())
	reduction.Div(reduction, rate.Denom())
	return new(big.Int).Sub(issueAmount, reduction)
}

func TestExtensionStateImpl_StartRewardIssueInvalid(t *testing.T) {
	termPeriod := int64(hvhmodule.TermPeriod)
	mcc, es := newMockContextAndExtensionState(t, newSimplePlatformConfig(termPeriod, 1))
	mcc.height = 100
	cc := NewCallContext(mcc, nil)
	err := es.StartRewardIssue(cc, 100)
	assert.Error(t, err)
}

func TestExtensionStateImpl_StartRewardIssueValid(t *testing.T) {
	var err error
	termPeriod := int64(hvhmodule.TermPeriod)
	mcc, es := newMockContextAndExtensionState(t, newSimplePlatformConfig(termPeriod, 1))
	mcc.height = 10
	cc := NewCallContext(mcc, nil)

	err = es.StartRewardIssue(cc, 100)
	assert.NoError(t, err)

	err = es.StartRewardIssue(cc, 110)
	assert.NoError(t, err)

	goByHeight(t, 110, es, mcc, nil)

	err = es.StartRewardIssue(cc, 150)
	assert.Error(t, err)
}

// Case 0
// - No planet
// - The first term has just been ended
func TestExtensionStateImpl_OnBaseTx(t *testing.T) {
	var balance *big.Int
	issueStart := int64(10)
	termPeriod := int64(100)
	from := common.MustNewAddressFromString("hx1234")

	mcc, es := newMockContextAndExtensionState(t, newSimplePlatformConfig(termPeriod, 1))
	mcc.height = 1
	cc := NewCallContext(mcc, from)

	balance = mcc.GetBalance(hvhmodule.HooverFund)
	assert.Zero(t, balance.Sign())

	err := es.StartRewardIssue(cc, issueStart)
	assert.NoError(t, err)

	goByHeight(t, issueStart, es, mcc, from)
	goByCount(t, termPeriod, es, mcc, from)

	balance = mcc.GetBalance(hvhmodule.HooverFund)
	assert.Zero(t, balance.Cmp(hvhmodule.BigIntHooverBudget))
}

func TestExtensionStateImpl_Reward0(t *testing.T) {
	id := int64(1)
	issueStart := int64(10)
	termPeriod := int64(100)
	owner := common.MustNewAddressFromString("hx1111")
	pm := common.MustNewAddressFromString("hx2222")

	mcc, es := newMockContextAndExtensionState(t, newSimplePlatformConfig(termPeriod, 1))
	mcc.height = 1
	cc := NewCallContext(mcc, owner)

	err := es.StartRewardIssue(cc, issueStart)
	assert.NoError(t, err)

	assert.NoError(t, es.AddPlanetManager(pm))
	ok, err := es.IsPlanetManager(pm)
	assert.True(t, ok)
	assert.NoError(t, err)

	priceInUSDT := toUSDT(5_000)
	priceInHVH := toHVH(50_000)
	err = es.RegisterPlanet(cc, id, false, false, owner, priceInUSDT, priceInHVH)
	assert.NoError(t, err)

	pi, err := es.GetPlanetInfo(cc, id)
	assert.NoError(t, err)
	assert.False(t, pi["isPrivate"].(bool))
	assert.False(t, pi["isCompany"].(bool))
	assert.True(t, pi["owner"].(module.Address).Equal(owner))
	assert.Equal(t, mcc.height, pi["height"].(int64))
	assert.Zero(t, pi["usdtPrice"].(*big.Int).Cmp(priceInUSDT))
	assert.Zero(t, pi["havahPrice"].(*big.Int).Cmp(priceInHVH))

	// termSeq 0 has just started
	goByHeight(t, issueStart, es, mcc, owner)

	// ReportPlanetWork
	cc = NewCallContext(mcc, pm)
	assert.NoError(t, es.ReportPlanetWork(cc, id))

	ri0, err := es.GetRewardInfoOf(cc, id)
	assert.NoError(t, err)
	checkRewardInfo(t, ri0, mcc.height,
		hvhmodule.BigIntInitIssueAmount,
		hvhmodule.BigIntInitIssueAmount,
		hvhmodule.BigIntInitIssueAmount)

	goByCount(t, 1, es, mcc, owner)

	cc = NewCallContext(mcc, owner)
	err = es.ClaimPlanetReward(cc, []int64{id})
	assert.NoError(t, err)

	balance := cc.GetBalance(owner)
	assert.Zero(t, ri0["claimable"].(*big.Int).Cmp(balance))

	ri1, err := es.GetRewardInfoOf(cc, id)
	assert.NoError(t, err)
	checkRewardInfo(t, ri1, mcc.height,
		ri0["total"].(*big.Int), hvhmodule.BigIntZero, hvhmodule.BigIntZero)

	goByCount(t, termPeriod, es, mcc, owner)

	balance = cc.GetBalance(hvhmodule.SustainableFund)
	assert.Zero(t, balance.Sign())
	balance = cc.GetBalance(hvhmodule.HooverFund)
	assert.Zero(t, balance.Sign())
}

// Case 1
// Not enough hoover budget
// pubic planet reward
func TestExtensionStateImpl_Reward1(t *testing.T) {
	id := int64(1)
	issueStart := int64(10)
	termPeriod := int64(100)
	issueAmount := toHVH(2)
	usdtPrice := toHVH(1) // 1 USDT == 1 HVH

	owner := common.MustNewAddressFromString("hx1111")
	pm := common.MustNewAddressFromString("hx2222")

	stateCfg := hvhstate.StateConfig{
		TermPeriod:  &common.HexInt64{Value: termPeriod},
		USDTPrice:   new(common.HexInt).SetValue(usdtPrice),
		IssueAmount: new(common.HexInt).SetValue(issueAmount),
	}
	mcc, es := newMockContextAndExtensionState(t, &PlatformConfig{StateConfig: stateCfg})
	mcc.height = 1
	cc := NewCallContext(mcc, owner)

	err := es.StartRewardIssue(cc, issueStart)
	assert.NoError(t, err)

	// Register a Planet
	priceInUSDT := toUSDT(36_000)
	priceInHVH := toHVH(36_000)
	err = es.RegisterPlanet(
		cc, id, false, false, owner, priceInUSDT, priceInHVH)
	assert.NoError(t, err)

	// termSeq 0 has just started
	goByHeight(t, issueStart, es, mcc, owner)
	assert.Zero(t, mcc.GetBalance(hvhmodule.PublicTreasury).Cmp(issueAmount))

	// height = issueStart + 10
	goByCount(t, 10, es, mcc, owner)
	assert.Equal(t, issueStart+10, mcc.BlockHeight())

	// Make the case where hooverBudget is not enough to support planet rewards
	balance := toHVH(3)
	mcc.SetBalance(hvhmodule.HooverFund, balance)
	assert.Zero(t, mcc.GetBalance(hvhmodule.HooverFund).Cmp(balance))

	balance = mcc.GetBalance(hvhmodule.PublicTreasury)
	assert.Zero(t, balance.Cmp(issueAmount))

	// ReportPlanetWork
	cc = NewCallContext(mcc, pm)
	assert.NoError(t, es.ReportPlanetWork(cc, id))

	// Check if hooverFund is transferred to public treasury
	assert.Zero(t, mcc.GetBalance(hvhmodule.HooverFund).Sign())
	assert.Zero(t, mcc.GetBalance(hvhmodule.PublicTreasury).Cmp(toHVH(5)))

	ri0, err := es.GetRewardInfoOf(cc, id)
	assert.NoError(t, err)
	checkRewardInfo(t, ri0, mcc.height, toHVH(5), toHVH(5), toHVH(5))

	goByCount(t, 1, es, mcc, owner)

	// Before claiming rewards
	assert.Zero(t, cc.GetBalance(owner).Sign())

	// Claim rewards
	cc = NewCallContext(mcc, owner)
	err = es.ClaimPlanetReward(cc, []int64{id})
	assert.NoError(t, err)

	// After claiming rewards
	balance = cc.GetBalance(owner)
	assert.Zero(t, ri0["claimable"].(*big.Int).Cmp(balance))

	ri1, err := es.GetRewardInfoOf(cc, id)
	assert.NoError(t, err)
	checkRewardInfo(t, ri1, mcc.height,
		ri0["total"].(*big.Int), hvhmodule.BigIntZero, hvhmodule.BigIntZero)
}

// Case 2
// private planet reward
// No claimable reward if private claimable rate is zero
func TestExtensionStateImpl_Reward2(t *testing.T) {
	id := int64(1)
	issueStart := int64(10)
	termPeriod := int64(10)
	issueAmount := toHVH(1)
	usdtPrice := toHVH(1) // 1 USDT == 1 HVH

	owner := common.MustNewAddressFromString("hx1111")
	pm := common.MustNewAddressFromString("hx2222")

	stateCfg := hvhstate.StateConfig{
		TermPeriod:  &common.HexInt64{Value: termPeriod},
		USDTPrice:   new(common.HexInt).SetValue(usdtPrice),
		IssueAmount: new(common.HexInt).SetValue(issueAmount),
	}
	mcc, es := newMockContextAndExtensionState(t, &PlatformConfig{StateConfig: stateCfg})
	mcc.height = 5
	cc := NewCallContext(mcc, pm)

	err := es.StartRewardIssue(cc, issueStart)
	assert.NoError(t, err)

	// Register a Planet
	priceInUSDT := toUSDT(1)
	priceInHVH := toHVH(1)
	err = es.RegisterPlanet(
		cc, id, true, false, owner, priceInUSDT, priceInHVH)
	assert.NoError(t, err)

	// termSeq 0 has just started
	goByHeight(t, issueStart, es, mcc, owner)
	assert.Zero(t, mcc.GetBalance(hvhmodule.PublicTreasury).Cmp(issueAmount))

	// height = issueStart + 5
	goByCount(t, 5, es, mcc, owner)
	assert.Equal(t, issueStart+5, mcc.BlockHeight())

	for i := 0; i < hvhmodule.DayPerYear; i++ {
		ri0, err := es.GetRewardInfoOf(cc, id)
		assert.NoError(t, err)
		total0, remain0, claimable0 := getRewardInfoOf(ri0)
		assert.Zero(t, claimable0.Sign())

		assert.NoError(t, es.ReportPlanetWork(cc, id))

		ri1, err := es.GetRewardInfoOf(cc, id)
		assert.NoError(t, err)
		total1, remain1, claimable1 := getRewardInfoOf(ri1)
		assert.Zero(t, claimable1.Sign())

		assert.Zero(t, issueAmount.Cmp(new(big.Int).Sub(total1, total0)))
		assert.Zero(t, issueAmount.Cmp(new(big.Int).Sub(remain1, remain0)))

		goByCount(t, termPeriod, es, mcc, pm)
	}

	jso, err := es.GetPrivateClaimableRate()
	assert.NoError(t, err)
	num, denom := getPrivateClaimableRate(jso)
	assert.Zero(t, int64(0), num)
	assert.Equal(t, int64(hvhmodule.PrivateClaimableRate), denom)
}

// Case 3
// company planet rewards
// Company planet rewards are distributed to EcoSystem and its planet owner in a ratio of 6:4
func TestExtensionStateImpl_Reward3(t *testing.T) {
	var balance *big.Int
	id := int64(1)
	issueStart := int64(10)
	termPeriod := int64(10)
	issueAmount := toHVH(10) // 10 HVH
	usdtPrice := toHVH(10)   // 1 USDT == 10 HVH
	owner := common.MustNewAddressFromString("hx2222")

	stateCfg := hvhstate.StateConfig{
		TermPeriod:  &common.HexInt64{Value: termPeriod},
		USDTPrice:   new(common.HexInt).SetValue(usdtPrice),
		IssueAmount: new(common.HexInt).SetValue(issueAmount),
	}
	mcc, es := newMockContextAndExtensionState(t, &PlatformConfig{StateConfig: stateCfg})
	mcc.height = 1
	cc := NewCallContext(mcc, owner)

	err := es.StartRewardIssue(cc, issueStart)
	assert.NoError(t, err)
	balance = cc.GetBalance(hvhmodule.PublicTreasury)
	assert.Zero(t, balance.Sign())

	priceInUSDT := toUSDT(5_000)
	priceInHVH := toHVH(50_000)
	err = es.RegisterPlanet(cc, id, false, true, owner, priceInUSDT, priceInHVH)

	// Before reporting a planet work, the balances of related accounts are 0
	assert.Zero(t, cc.GetBalance(hvhmodule.EcoSystem).Sign())
	assert.Zero(t, cc.GetBalance(owner).Sign())

	// Go To issueStart height
	height := issueStart + termPeriod/2
	goByHeight(t, height, es, mcc, nil)

	err = es.ReportPlanetWork(cc, id)
	assert.NoError(t, err)

	// Expected balances after reporting a planet work
	ecoReward := new(big.Int).Mul(issueAmount, hvhmodule.BigRatEcoSystemToCompanyReward.Num())
	ecoReward.Div(ecoReward, hvhmodule.BigRatEcoSystemToCompanyReward.Denom())
	ownerReward := new(big.Int).Sub(issueAmount, ecoReward)

	// Check if reward info is correct
	jso, err := es.GetRewardInfoOf(cc, id)
	assert.NoError(t, err)
	checkRewardInfo(t, jso, cc.BlockHeight(), issueAmount, ownerReward, ownerReward)

	// Claim rewards for company planet owner
	err = es.ClaimPlanetReward(cc, []int64{id})
	assert.NoError(t, err)
	assert.Zero(t, ownerReward.Cmp(cc.GetBalance(owner)))
	// EcoSystem reward will be transferred at the beginning of the next term
	assert.Zero(t, cc.GetBalance(hvhmodule.EcoSystem).Sign())

	// Go to the next term start
	goByHeight(t, issueStart+termPeriod*2, es, mcc, nil)

	// Rewards for EcoSystem will be claimed automatically at the beginning of the next term
	assert.Zero(t, ecoReward.Cmp(cc.GetBalance(hvhmodule.EcoSystem)))
}

// Case4
// private planet
// claim rewards during private release cycle
func TestExtensionStateImpl_Reward_With_PrivateClaimableRate(t *testing.T) {
	var err error
	id := int64(1)
	issueStart := int64(10)
	termPeriod := int64(4)
	issueAmount := big.NewInt(1_000_000)
	usdtPrice := toHVH(1) // 1 USDT == 10 HVH
	repeat := int64(5)
	owner := common.MustNewAddressFromString("hx2222")

	stateCfg := hvhstate.StateConfig{
		TermPeriod:  &common.HexInt64{Value: termPeriod},
		IssueAmount: new(common.HexInt).SetValue(issueAmount),
		USDTPrice:   new(common.HexInt).SetValue(usdtPrice),
	}
	mcc, es := newMockContextAndExtensionState(t, &PlatformConfig{StateConfig: stateCfg})
	mcc.height = 1
	cc := NewCallContext(mcc, owner)

	err = es.StartRewardIssue(cc, issueStart)
	assert.NoError(t, err)

	priceInUSDT := big.NewInt(1)
	priceInHVH := big.NewInt(1)
	err = es.RegisterPlanet(cc, id, true, false, owner, priceInUSDT, priceInHVH)
	assert.NoError(t, err)

	// Term 0
	goByHeight(t, issueStart, es, mcc, nil)

	// Go to the next term
	goByCount(t, termPeriod, es, mcc, nil)

	jso, err := es.GetRewardInfoOf(cc, id)
	assert.NoError(t, err)
	total, remain, claimable := getRewardInfoOf(jso)
	assert.Zero(t, total.Sign())
	assert.Zero(t, remain.Sign())
	assert.Zero(t, claimable.Sign())

	termHeight := mcc.height
	for i := int64(0); i < repeat; i++ {
		// Go to the next term + 1 height
		termHeight += termPeriod
		goByHeight(t, termHeight+1, es, mcc, nil)

		err = es.ReportPlanetWork(cc, id)
		assert.NoError(t, err)

		err = es.SetPrivateClaimableRate(i+1, repeat)
		assert.NoError(t, err)

		jso, err := es.GetRewardInfoOf(cc, id)
		assert.NoError(t, err)
		_, _, claimable = getRewardInfoOf(jso)
		assert.True(t, claimable.Sign() > 0)

		oldBalance := mcc.GetBalance(owner)

		err = es.ClaimPlanetReward(cc, []int64{id})
		assert.NoError(t, err)

		balance := mcc.GetBalance(owner)
		assert.Zero(t, balance.Cmp(new(big.Int).Add(oldBalance, claimable)))
	}

	goByCount(t, 1, es, mcc, nil)

	jso, err = es.GetPrivateClaimableRate()
	assert.NoError(t, err)
	num, denom := getPrivateClaimableRate(jso)
	assert.Equal(t, repeat, num)
	assert.Equal(t, repeat, denom)

	claimed := new(big.Int).Mul(issueAmount, big.NewInt(repeat))
	jso, err = es.GetRewardInfoOf(cc, id)
	assert.NoError(t, err)
	total, remain, claimable = getRewardInfoOf(jso)
	assert.Zero(t, total.Cmp(claimed))
	assert.Zero(t, remain.Sign())
	assert.Zero(t, claimable.Sign())

	balance := mcc.GetBalance(owner)
	assert.Zero(t, balance.Cmp(claimed))
}

// Case5
// Reversed PrivateClaimableRate
// 2 planets: user planet(1), private planet(1), company planet(0)
func TestExtensionSnapshotImpl_Reward5_ReversedPrivateClaimableRate(t *testing.T) {
	var err error
	id := int64(1)
	issueStart := int64(10)
	termPeriod := int64(4)
	issueAmount := big.NewInt(100)
	usdtPrice := toHVH(1) // 1 USDT == 10 HVH
	owner := common.MustNewAddressFromString("hx2222")

	stateCfg := hvhstate.StateConfig{
		TermPeriod:  &common.HexInt64{Value: termPeriod},
		IssueAmount: new(common.HexInt).SetValue(issueAmount),
		USDTPrice:   new(common.HexInt).SetValue(usdtPrice),
	}
	mcc, es := newMockContextAndExtensionState(t, &PlatformConfig{StateConfig: stateCfg})
	mcc.height = 1
	cc := NewCallContext(mcc, owner)

	err = es.StartRewardIssue(cc, issueStart)
	assert.NoError(t, err)

	priceInUSDT := big.NewInt(1)
	priceInHVH := big.NewInt(1)
	err = es.RegisterPlanet(cc, id, true, false, owner, priceInUSDT, priceInHVH)
	assert.NoError(t, err)

	// Term 0
	goByHeight(t, issueStart, es, mcc, nil)

	// Go to the next term
	goByCount(t, termPeriod, es, mcc, nil)

	jso, err := es.GetRewardInfoOf(cc, id)
	assert.NoError(t, err)
	total, remain, claimable := getRewardInfoOf(jso)
	assert.Zero(t, total.Sign())
	assert.Zero(t, remain.Sign())
	assert.Zero(t, claimable.Sign())

	err = es.SetPrivateClaimableRate(4, 5)
	assert.NoError(t, err)

	expRemains := []int64{100, 120, 220, 320, 420}
	expClaimables := []int64{80, 0, 0, 0, 20}

	for i := 0; i < 5; i++ {
		goByCount(t, termPeriod, es, mcc, nil)

		err = es.ReportPlanetWork(cc, id)
		assert.NoError(t, err)

		jso, err = es.GetRewardInfoOf(cc, id)
		assert.NoError(t, err)
		total, remain, claimable = getRewardInfoOf(jso)
		es.Logger().Debugf(
			"i=%d total=%d remain=%d claimable=%d", i, total, remain, claimable)
		assert.Equal(t, issueAmount.Int64()*int64(i+1), total.Int64())
		assert.Equal(t, expRemains[i], remain.Int64())
		assert.Equal(t, expClaimables[i], claimable.Int64())

		oldBalance := mcc.GetBalance(owner)

		err = es.ClaimPlanetReward(cc, []int64{id})
		assert.NoError(t, err)

		balance := mcc.GetBalance(owner)
		assert.Zero(t, balance.Cmp(new(big.Int).Add(oldBalance, claimable)))

		if i == 0 {
			err = es.SetPrivateClaimableRate(1, 5)
			assert.NoError(t, err)
		}
	}
}

func TestExtensionStateImpl_DistributeFee(t *testing.T) {
	var balance *big.Int
	issueStart := int64(10)
	termPeriod := int64(10)
	issueAmount := toHVH(10)
	usdtPrice := toHVH(1) // 1 USDT == 1 HVH
	owner := common.MustNewAddressFromString("hx2222")

	stateCfg := hvhstate.StateConfig{
		TermPeriod:  &common.HexInt64{Value: termPeriod},
		USDTPrice:   new(common.HexInt).SetValue(usdtPrice),
		IssueAmount: new(common.HexInt).SetValue(issueAmount),
	}
	mcc, es := newMockContextAndExtensionState(t, &PlatformConfig{StateConfig: stateCfg})
	mcc.height = 1
	cc := NewCallContext(mcc, owner)

	err := es.StartRewardIssue(cc, issueStart)
	assert.NoError(t, err)
	balance = cc.GetBalance(hvhmodule.PublicTreasury)
	assert.Zero(t, balance.Sign())

	// Initial condition setting
	mcc.SetBalance(hvhmodule.ServiceTreasury, toHVH(10))
	mcc.SetBalance(hvhmodule.HooverFund, hvhmodule.BigIntHooverBudget)
	mcc.SetBalance(cc.Treasury(), toHVH(20))
	es.Logger().Infof("Treasury: %s", cc.Treasury())

	assert.Zero(t, mcc.GetBalance(hvhmodule.EcoSystem).Sign())
	assert.Zero(t, mcc.GetBalance(hvhmodule.SustainableFund).Sign())

	// 1st term start
	goByHeight(t, issueStart, es, mcc, nil)

	// No fee distribution at the first term
	assert.Zero(t, mcc.GetBalance(hvhmodule.EcoSystem).Sign())
	assert.Zero(t, mcc.GetBalance(hvhmodule.SustainableFund).Sign())

	// 2nd term start
	goByCount(t, termPeriod, es, mcc, nil)

	assert.Zero(t, mcc.GetBalance(hvhmodule.EcoSystem).Cmp(toHVH(6)))
	assert.Zero(t, mcc.GetBalance(hvhmodule.SustainableFund).Cmp(new(big.Int).Add(issueAmount, toHVH(24))))
	assert.Zero(t, mcc.GetBalance(hvhmodule.HooverFund).Cmp(hvhmodule.BigIntHooverBudget))
	assert.Zero(t, mcc.GetBalance(hvhmodule.ServiceTreasury).Sign())
	assert.Zero(t, mcc.GetBalance(cc.Treasury()).Sign())
}

func TestExtensionStateImpl_IssueReduction(t *testing.T) {
	var balance *big.Int
	issueStart := int64(10)
	termPeriod := int64(10)
	issueReductionCycle := int64(10)
	issueAmount := toHVH(100)
	usdtPrice := toHVH(1) // 1 USDT == 1 HVH

	// pm := common.MustNewAddressFromString("hx1111")
	owner := common.MustNewAddressFromString("hx2222")

	stateCfg := hvhstate.StateConfig{
		TermPeriod:          &common.HexInt64{Value: termPeriod},
		USDTPrice:           new(common.HexInt).SetValue(usdtPrice),
		IssueAmount:         new(common.HexInt).SetValue(issueAmount),
		IssueReductionCycle: &common.HexInt64{Value: issueReductionCycle},
	}
	mcc, es := newMockContextAndExtensionState(t, &PlatformConfig{StateConfig: stateCfg})
	mcc.height = 1
	cc := NewCallContext(mcc, owner)

	err := es.StartRewardIssue(cc, issueStart)
	assert.NoError(t, err)
	balance = cc.GetBalance(hvhmodule.PublicTreasury)
	assert.Zero(t, balance.Sign())

	height := issueStart
	for i := int64(0); i < issueReductionCycle; i++ {
		goByHeight(t, height, es, mcc, owner)
		assert.Zero(t, (mcc.height-issueStart)%termPeriod)

		balance = mcc.GetBalance(hvhmodule.PublicTreasury)
		assert.Zero(t, balance.Cmp(issueAmount))

		height += termPeriod
		fmt.Println(height)
	}

	totalIssued := new(big.Int).Mul(issueAmount, big.NewInt(issueReductionCycle))
	totalSupply := cc.GetTotalSupply()
	assert.Zero(t, totalSupply.Cmp(totalIssued))

	hfBalance := mcc.GetBalance(hvhmodule.HooverFund)
	assert.Zero(t, hfBalance.Cmp(new(big.Int).Mul(issueAmount, big.NewInt(issueReductionCycle-1))))

	sfBalance := mcc.GetBalance(hvhmodule.SustainableFund)
	assert.Zero(t, sfBalance.Sign())

	es.Logger().Errorf("hf=%d sf=%d", hfBalance, sfBalance)

	// The first term start of the 2nd issueReductionCycle
	issueAmount = calcNewIssueAmount(issueAmount, hvhmodule.BigRatIssueReductionRate)
	goByHeight(t, height, es, mcc, owner)
	balance = mcc.GetBalance(hvhmodule.PublicTreasury)
	assert.Zero(t, balance.Cmp(issueAmount))

	for i := int64(0); i < issueReductionCycle; i++ {
		goByHeight(t, height, es, mcc, owner)
		assert.Zero(t, (mcc.height-issueStart)%termPeriod)

		balance = mcc.GetBalance(hvhmodule.PublicTreasury)
		assert.Zero(t, balance.Cmp(issueAmount))

		height += termPeriod
		fmt.Println(height)
	}

	// The first term start of the 3rd issueReductionCycle
	issueAmount = calcNewIssueAmount(issueAmount, hvhmodule.BigRatIssueReductionRate)
	goByHeight(t, height, es, mcc, owner)
	balance = mcc.GetBalance(hvhmodule.PublicTreasury)
	assert.Zero(t, balance.Cmp(issueAmount))
}

func TestExtensionStateImpl_SetPlanetOwner(t *testing.T) {
	issueStart := int64(10)
	termPeriod := int64(10)
	issueReductionCycle := int64(10)
	issueAmount := toHVH(100)
	usdtPrice := toHVH(1) // 1 USDT == 1 HVH

	stateCfg := hvhstate.StateConfig{
		TermPeriod:          &common.HexInt64{Value: termPeriod},
		USDTPrice:           new(common.HexInt).SetValue(usdtPrice),
		IssueAmount:         new(common.HexInt).SetValue(issueAmount),
		IssueReductionCycle: &common.HexInt64{Value: issueReductionCycle},
	}
	mcc, es := newMockContextAndExtensionState(t, &PlatformConfig{StateConfig: stateCfg})
	mcc.height = 1
	cc := NewCallContext(mcc, nil)

	err := es.StartRewardIssue(cc, issueStart)
	assert.NoError(t, err)

	oldOwners := []module.Address{
		common.MustNewAddressFromString("hx11"),
		common.MustNewAddressFromString("hx12"),
		common.MustNewAddressFromString("hx13"),
	}
	newOwners := []module.Address{
		common.MustNewAddressFromString("hx21"),
		common.MustNewAddressFromString("hx22"),
		common.MustNewAddressFromString("hx23"),
	}

	priceInUSDT := toUSDT(5_000)
	priceInHVH := toHVH(50_000)

	for i := 0; i < len(oldOwners); i++ {
		id := int64(i + 1)
		err = es.RegisterPlanet(
			cc, id, false, false,
			oldOwners[i], priceInUSDT, priceInHVH)
		assert.NoError(t, err)
		jso, err := es.GetPlanetInfo(cc, id)
		assert.NoError(t, err)
		assert.True(t, jso["owner"].(module.Address).Equal(oldOwners[i]))
	}

	goByHeight(t, issueStart, es, mcc, nil)
	goByCount(t, termPeriod, es, mcc, nil)

	for i := 0; i < len(oldOwners); i++ {
		id := int64(i + 1)
		jso, err := es.GetPlanetInfo(cc, id)
		assert.NoError(t, err)
		assert.True(t, jso["owner"].(module.Address).Equal(oldOwners[i]))

		err = es.SetPlanetOwner(cc, id, newOwners[i])
		assert.NoError(t, err)

		jso2, err := es.GetPlanetInfo(cc, id)
		assert.NoError(t, err)
		assert.True(t, jso2["owner"].(module.Address).Equal(newOwners[i]))

		assert.Equal(t, jso["isPrivate"].(bool), jso2["isPrivate"].(bool))
		assert.Equal(t, jso["isCompany"].(bool), jso2["isCompany"].(bool))
		assert.Equal(t, jso["usdtPrice"].(*big.Int), jso2["usdtPrice"].(*big.Int))
		assert.Equal(t, jso["havahPrice"].(*big.Int), jso2["havahPrice"].(*big.Int))
		assert.Equal(t, jso["height"].(int64), jso2["height"].(int64))
	}
}

func TestExtensionStateImpl_ReportPlanetWork_BeforeStartRewardIssue(t *testing.T) {
	var err error
	id := int64(1)
	termPeriod := int64(10)
	issueReductionCycle := int64(10)
	issueAmount := toHVH(100)
	usdtPrice := toHVH(1) // 1 USDT == 1 HVH
	owner := common.MustNewAddressFromString("hx1234")

	stateCfg := hvhstate.StateConfig{
		TermPeriod:          &common.HexInt64{Value: termPeriod},
		USDTPrice:           new(common.HexInt).SetValue(usdtPrice),
		IssueAmount:         new(common.HexInt).SetValue(issueAmount),
		IssueReductionCycle: &common.HexInt64{Value: issueReductionCycle},
	}
	mcc, es := newMockContextAndExtensionState(t, &PlatformConfig{StateConfig: stateCfg})
	mcc.height = 1
	cc := NewCallContext(mcc, nil)

	priceInUSDT := toUSDT(5_000)
	priceInHVH := toHVH(50_000)
	err = es.RegisterPlanet(cc, id, false, false, owner, priceInUSDT, priceInHVH)
	assert.NoError(t, err)

	goByCount(t, 100, es, mcc, nil)

	err = es.ReportPlanetWork(cc, id)
	assert.Error(t, err)
}
