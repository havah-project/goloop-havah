package hvh

import (
	"encoding/json"
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
	return hvhmodule.ServiceTreasury
}

func (cc *mockCallContext) GetBalance(address module.Address) *big.Int {
	as := cc.GetAccountState(address.ID())
	return as.GetBalance()
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

	for i := int64(0); i < count; i++ {
		mcc.height++
		if hvhstate.IsIssueStarted(mcc.height, is) {
			data := es.NewBaseTransactionData(mcc.height, is)
			bs, err := json.Marshal(data)
			assert.NoError(t, err)

			err = es.OnBaseTx(cc, bs)
			assert.NoError(t, err)
		}
	}
}

func checkRewardInfo(t *testing.T,
	jso map[string]interface{},
	height int64, total, remain, claimable *big.Int) {
	assert.Equal(t, height, jso["height"].(int64))

	keys := []string{"total", "remain", "claimable"}
	values := []*big.Int{total, remain, claimable}
	for i, key := range keys {
		assert.Zero(t, jso[key].(*big.Int).Cmp(values[i]))
	}
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
	termPeriod := int64(hvhmodule.TermPeriod)
	mcc, es := newMockContextAndExtensionState(t, newSimplePlatformConfig(termPeriod, 1))
	mcc.height = 100
	cc := NewCallContext(mcc, nil)
	err := es.StartRewardIssue(cc, 101)
	assert.NoError(t, err)
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

	ri0, err := es.GetRewardInfo(cc, id)
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

	ri1, err := es.GetRewardInfo(cc, id)
	assert.NoError(t, err)
	checkRewardInfo(t, ri1, mcc.height,
		ri0["total"].(*big.Int), hvhmodule.BigIntZero, hvhmodule.BigIntZero)

	goByCount(t, termPeriod, es, mcc, owner)

	balance = cc.GetBalance(hvhmodule.SustainableFund)
	assert.Zero(t, balance.Sign())
	balance = cc.GetBalance(hvhmodule.HooverFund)
	assert.Zero(t, balance.Sign())
}
