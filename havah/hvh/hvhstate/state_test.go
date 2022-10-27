package hvhstate

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/db"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/havah/hvhutils"
	"github.com/icon-project/goloop/module"
)

func newDummyState() *State {
	mdb := db.NewMapDB()
	snapshot := NewSnapshot(mdb, nil)
	logger := hvhutils.NewLogger(nil)
	return NewStateFromSnapshot(snapshot, false, logger)
}

func newDummyPlanet(isPrivate, isCompany bool, height int64) *Planet {
	owner := common.MustNewAddressFromString("hx123")
	usdt := big.NewInt(1000)
	price := big.NewInt(2000)
	return newPlanet(isPrivate, isCompany, owner, usdt, price, height)
}

func newPlanetReward(total, current *big.Int, lastTN int64) *planetReward {
	return &planetReward{
		total:   total,
		lastTN:  lastTN,
		current: current,
	}
}

func checkAllPlanet(t *testing.T, state *State, expected int64) {
	cur := state.getVarDB(hvhmodule.VarAllPlanet).Int64()
	assert.Equal(t, expected, cur)
}

func safeSetBigInt(t *testing.T, state *State, key string, value *big.Int) {
	orgValue := new(big.Int).Set(value)
	assert.NoError(t, state.setBigInt(key, value))

	value = state.getBigInt(key)
	assert.Zero(t, orgValue.Cmp(value))
}

func deepCopyPlanet(t *testing.T, p *Planet) *Planet {
	p2 := &Planet{
		p.dirty,
		p.flags,
		p.owner,
		new(big.Int).Set(p.usdt),
		new(big.Int).Set(p.price),
		p.height,
	}
	assert.True(t, p2.equal(p))
	return p2
}

func deepCopyPlanetReward(t *testing.T, pr *planetReward) *planetReward {
	pr2 := &planetReward{
		new(big.Int).Set(pr.Total()),
		pr.LastTermNumber(),
		new(big.Int).Set(pr.Current()),
	}
	assert.True(t, pr2.equal(pr))
	return pr2
}

func toUSDT(value int64) *big.Int {
	usdt := big.NewInt(value)
	return usdt.Mul(usdt, hvhmodule.BigIntUSDTDecimal)
}

func toHVH(value int64) *big.Int {
	hvh := big.NewInt(value)
	return hvh.Mul(hvh, hvhmodule.BigIntCoinDecimal)
}

func TestState_SetUSDTPrice(t *testing.T) {
	var price *big.Int
	state := newDummyState()

	price = state.GetUSDTPrice()
	assert.Zero(t, price.Sign())

	// In case of valid prices
	prices := []*big.Int{new(big.Int), big.NewInt(1234)}
	for _, price = range prices {
		assert.NoError(t, state.SetUSDTPrice(price))
		price2 := state.GetUSDTPrice()
		assert.Zero(t, price.Cmp(price2))
	}

	// In case of invalid prices
	invalidPrices := []*big.Int{nil, big.NewInt(-1000)}
	for _, price = range invalidPrices {
		assert.Error(t, state.SetUSDTPrice(price))
	}
}

func TestState_AddPlanetManager(t *testing.T) {
	var err error
	var address module.Address

	state := newDummyState()
	addresses := []module.Address{
		common.MustNewAddressFromString("cx1"),
		common.MustNewAddressFromString("cx2"),
		common.MustNewAddressFromString("hx1"),
		common.MustNewAddressFromString("hx2"),
		common.MustNewAddressFromString("hx3"),
	}

	// Call AddPlanetManager() with a valid argument
	for _, address = range addresses {
		err = state.AddPlanetManager(address)
		assert.NoError(t, err)

		ok, err := state.IsPlanetManager(address)
		assert.True(t, ok)
		assert.NoError(t, err)
	}

	// Call AddPlanetManager with an invalid argument
	for _, address = range []module.Address{nil, addresses[0], addresses[1]} {
		err = state.AddPlanetManager(address)
		assert.Error(t, err)
	}

	// Remove the first item with RemovePlanetManager()
	address = addresses[0]
	err = state.RemovePlanetManager(address)
	assert.NoError(t, err)

	ok, err := state.IsPlanetManager(address)
	assert.NoError(t, err)
	assert.False(t, ok)

	for i := 1; i < len(addresses); i++ {
		ok, err = state.IsPlanetManager(addresses[i])
		assert.True(t, ok)
		assert.NoError(t, err)
	}

	// Remove the last item with RemovePlanetManager()
	address = addresses[len(addresses)-1]
	err = state.RemovePlanetManager(address)
	assert.NoError(t, err)

	ok, err = state.IsPlanetManager(address)
	assert.NoError(t, err)
	assert.False(t, ok)

	for i := 1; i < len(addresses)-1; i++ {
		ok, err = state.IsPlanetManager(addresses[i])
		assert.True(t, ok)
		assert.NoError(t, err)
	}

	// Invalid cases of RemovePlanetManager()
	for _, address = range []module.Address{nil, addresses[0], common.MustNewAddressFromString("hx1234")} {
		err = state.RemovePlanetManager(address)
		assert.Error(t, err)
	}

	ok, err = state.IsPlanetManager(nil)
	assert.Error(t, err)
	assert.False(t, ok)
}

func TestState_SetIssueStart(t *testing.T) {
	var startBH, curBH, height int64
	state := newDummyState()

	// Success case: startBH > 0 && startBH > curBH
	startBH, curBH = 2000, 1000
	err := state.SetIssueStart(curBH, startBH)
	assert.NoError(t, err)

	height = state.getVarDB(hvhmodule.VarIssueStart).Int64()
	assert.Equal(t, startBH, height)
	assert.Equal(t, startBH, state.GetIssueStart())

	// Failure case: startBH <= 0 || startBH <= curBH
	curBH = 1000
	height = state.getVarDB(hvhmodule.VarIssueStart).Int64()
	for _, startBH = range []int64{-100, 0, 100, 500, curBH} {
		err = state.SetIssueStart(curBH, startBH)
		assert.Error(t, err)

		startBH = state.getVarDB(hvhmodule.VarIssueStart).Int64()
		assert.Equal(t, height, startBH)
	}
}

func TestState_RegisterPlanet(t *testing.T) {
	state := newDummyState()
	owner := common.MustNewAddressFromString("hx1234")
	isCompany := true
	isPrivate := true
	usdt := big.NewInt(1000)
	price := new(big.Int).Mul(usdt, big.NewInt(10))
	height := int64(200)
	var planetCount int64

	checkAllPlanet(t, state, 0)
	expectedCount := int64(0)
	for i := 0; i < 3; i++ {
		err := state.RegisterPlanet(int64(i), isPrivate, isCompany, owner, usdt, price, height)
		assert.NoError(t, err)

		expectedCount++
		planetCount = state.getVarDB(hvhmodule.VarAllPlanet).Int64()
		assert.Equal(t, expectedCount, planetCount)
	}

	for i := 0; i < 2; i++ {
		err := state.UnregisterPlanet(int64(i))
		assert.NoError(t, err)
		expectedCount--
		checkAllPlanet(t, state, expectedCount)
	}

	planetCount = state.getVarDB(hvhmodule.VarAllPlanet).Int64()
	err := state.UnregisterPlanet(int64(100))
	assert.Error(t, err)
	checkAllPlanet(t, state, planetCount)
}

func TestState_UnregisterPlanet_InAbnormalCase(t *testing.T) {
	state := newDummyState()
	checkAllPlanet(t, state, int64(0))
	err := state.UnregisterPlanet(int64(-1))
	assert.Error(t, err)

	err = state.UnregisterPlanet(int64(100))
	assert.Error(t, err)

	checkAllPlanet(t, state, int64(0))
}

func TestState_GetBigInt(t *testing.T) {
	key := "key"
	state := newDummyState()
	varDB := state.getVarDB(key)

	assert.Nil(t, varDB.BigInt())

	value := state.getBigInt(key)
	assert.NotNil(t, value)

	defValue := hvhmodule.BigIntHooverBudget
	value = state.getBigIntOrDefault(key, defValue)
	assert.Zero(t, defValue.Cmp(value))

	newValue := new(big.Int).Add(defValue, big.NewInt(100))
	err := state.setBigInt(key, newValue)
	assert.NoError(t, err)

	value = state.getBigInt(key)
	assert.Zero(t, newValue.Cmp(value))

	value = state.getBigIntOrDefault(key, defValue)
	assert.Zero(t, value.Cmp(newValue))
}

func TestState_GetPlanet(t *testing.T) {
	id := int64(1)

	state := newDummyState()
	p, err := state.GetPlanet(id)
	assert.Nil(t, p)
	assert.Error(t, err)

	dictDB := state.getDictDB(hvhmodule.DictPlanet, 1)

	p = newDummyPlanet(false, false, 1234)
	err = state.setPlanet(dictDB, id, p)
	assert.NoError(t, err)

	p2, err := state.getPlanet(dictDB, id)
	assert.NotNil(t, p)
	assert.NoError(t, err)
	assert.True(t, p.equal(p2))
}

func TestState_GetPlanetReward(t *testing.T) {
	id := int64(1)
	reward := big.NewInt(100)

	state := newDummyState()
	pr, err := state.GetPlanetReward(id)
	assert.NotNil(t, pr)
	assert.NoError(t, err)
	assert.Zero(t, pr.Total().Sign())
	assert.Zero(t, pr.Current().Sign())
	assert.Zero(t, pr.LastTermNumber())

	err = pr.increment(10, reward, reward)
	assert.NoError(t, err)

	err = state.setPlanetReward(id, pr)
	assert.NoError(t, err)

	pr2, err := state.GetPlanetReward(id)
	assert.NotNil(t, pr2)
	assert.NoError(t, err)
	assert.True(t, pr.equal(pr2))
}

func TestState_IncrementWorkingPlanet(t *testing.T) {
	state := newDummyState()

	ov := state.getInt64(hvhmodule.VarWorkingPlanet)
	err := state.IncrementWorkingPlanet()
	assert.NoError(t, err)

	nv := state.getInt64(hvhmodule.VarWorkingPlanet)
	assert.Equal(t, ov+1, nv)
}

func TestState_InitState(t *testing.T) {
	state := newDummyState()

	assert.Equal(t, int64(hvhmodule.TermPeriod), state.GetTermPeriod())
	assert.Zero(t, state.GetHooverBudget().Cmp(hvhmodule.BigIntHooverBudget))
	assert.Equal(t, int64(hvhmodule.IssueReductionCycle), state.GetIssueReductionCycle())
	assert.Zero(t, state.GetIssueStart())
	assert.Zero(t, state.GetIssueReductionRate().Cmp(hvhmodule.BigRatIssueReductionRate))

	cfg := &StateConfig{}
	cfg.TermPeriod = &common.HexInt64{Value: 100}
	cfg.HooverBudget = common.NewHexInt(1_000_000)
	cfg.IssueReductionCycle = &common.HexInt64{Value: 180}
	cfg.USDTPrice = common.NewHexInt(200_000)

	assert.NoError(t, state.InitState(cfg))
	assert.Equal(t, cfg.TermPeriod.Value, state.GetTermPeriod())
	assert.Zero(t, state.GetHooverBudget().Cmp(cfg.HooverBudget.Value()))
	assert.Equal(t, cfg.IssueReductionCycle.Value, state.GetIssueReductionCycle())
	assert.Zero(t, state.GetIssueReductionRate().Cmp(hvhmodule.BigRatIssueReductionRate))
}

func TestState_DecreaseRewardRemain(t *testing.T) {
	const initRewardRemain = 1_000_000
	const key = hvhmodule.VarRewardRemain
	state := newDummyState()

	safeSetBigInt(t, state, key, big.NewInt(initRewardRemain))

	amount := int64(1000)
	err := state.DecreaseRewardRemain(big.NewInt(amount))
	assert.NoError(t, err)

	newRewardRemain := state.getBigInt(key)
	assert.Zero(t, big.NewInt(initRewardRemain).Cmp(new(big.Int).Add(newRewardRemain, big.NewInt(amount))))
}

func TestState_calcClaimableReward(t *testing.T) {
	height := int64(100)
	owner := common.MustNewAddressFromString("hx12345")
	usdt := toUSDT(5000)
	price := toHVH(50000)
	p := newPlanet(false, false, owner, usdt, price, 100)
	pr := newEmpyPlanetReward()
	state := newDummyState()

	amount := toHVH(1)
	err := pr.increment(1, amount, amount)
	assert.NoError(t, err)

	p2 := deepCopyPlanet(t, p)
	pr2 := deepCopyPlanetReward(t, pr)

	reward, err := state.calcClaimableReward(height, p, pr)
	assert.NoError(t, err)
	assert.Zero(t, reward.Cmp(amount))
	assert.True(t, p.equal(p2))
	assert.True(t, pr.equal(pr2))
}

func TestState_ClaimPlanetReward(t *testing.T) {
	const (
		publicId  = 1
		companyId = 2
		privateId = 3
	)
	issueStart := int64(100)
	owner := common.MustNewAddressFromString("hx1234")

	state := newDummyState()

	ps := []*Planet{
		nil,
		newPlanet(false, false, owner, toUSDT(5000), toHVH(50000), 10), // public
		newPlanet(false, true, owner, toUSDT(2000), toHVH(20000), 10),  // company
		newPlanet(true, false, owner, toUSDT(3000), toHVH(30000), 10),  // private
	}
	prs := []*planetReward{
		nil,
		newPlanetReward(toHVH(10), toHVH(3), 0),
		newPlanetReward(toHVH(10), toHVH(1), 0),
		newPlanetReward(toHVH(10), toHVH(2), 0),
	}
	for i := 1; i < 4; i++ {
		pr := deepCopyPlanetReward(t, prs[i])
		err := state.setPlanetReward(int64(i), pr)
		assert.NoError(t, err)
	}

	planetDictDB := state.getDictDB(hvhmodule.DictPlanet, 1)
	for i := 1; i < 4; i++ {
		p := deepCopyPlanet(t, ps[i])
		err := state.setPlanet(planetDictDB, int64(i), p)
		assert.NoError(t, err)
	}

	err := state.SetIssueStart(1, issueStart)
	assert.NoError(t, err)

	height := state.GetIssueStart() + state.GetTermPeriod()
	for _, id := range []int64{publicId, companyId} {
		reward, err := state.ClaimPlanetReward(id, height, owner)
		assert.NoError(t, err)
		assert.Zero(t, reward.Cmp(prs[id].Current()))

		pr, err := state.GetPlanetReward(id)
		assert.NoError(t, err)
		assert.Zero(t, pr.Total().Cmp(prs[id].Total()))
		assert.Zero(t, pr.Current().Sign())
		assert.Equal(t, pr.LastTermNumber(), prs[id].LastTermNumber())
	}

	id := int64(privateId)
	reward, err := state.ClaimPlanetReward(id, height, owner)
	assert.NoError(t, err)
	assert.Zero(t, reward.Sign())
}

func TestState_SetPrivateClaimableRate(t *testing.T) {
	var err error
	var num, denom int64
	var expNum, expDenom int64
	state := newDummyState()

	// Check default value
	expNum = int64(0)
	expDenom = int64(hvhmodule.PrivateClaimableRate)
	num, denom = state.GetPrivateClaimableRate()
	assert.Zero(t, num)
	assert.Equal(t, expDenom, denom)

	// Error cases
	// num, denom
	ins := [][]int64{
		{0, 0},
		{1, 0},
		{-2, 10},
		{2, -10},
		{-3, -20},
		{10, 5},
		{10, 10001},
		{10001, 10001},
	}
	for _, in := range ins {
		err = state.SetPrivateClaimableRate(in[0], in[1])
		assert.Error(t, err)

		num, denom = state.GetPrivateClaimableRate()
		assert.Zero(t, num)
		assert.Equal(t, expDenom, denom)
	}

	ins = [][]int64{
		{0, 100},
		{1, 100},
		{10000, 10000},
		{23, 24},
		{24, 24},
	}
	for _, in := range ins {
		expNum, expDenom = in[0], in[1]
		err = state.SetPrivateClaimableRate(expNum, expDenom)
		assert.NoError(t, err)

		num, denom = state.GetPrivateClaimableRate()
		assert.Equal(t, expNum, num)
		assert.Equal(t, expDenom, denom)
	}
}

func TestGetTermSequenceAndBlockIndex(t *testing.T) {
	termSeq, blockIndex := GetTermSequenceAndBlockIndex(5, 20, 10)
	assert.True(t, termSeq < 0)
	assert.True(t, blockIndex < 0)

	termSeq, blockIndex = GetTermSequenceAndBlockIndex(20, 20, 10)
	assert.Zero(t, termSeq)
	assert.Zero(t, blockIndex)

	termSeq, blockIndex = GetTermSequenceAndBlockIndex(21, 20, 10)
	assert.Equal(t, int64(0), termSeq)
	assert.Equal(t, int64(1), blockIndex)

	termSeq, blockIndex = GetTermSequenceAndBlockIndex(30, 20, 10)
	assert.Equal(t, int64(1), termSeq)
	assert.Equal(t, int64(0), blockIndex)

	termSeq, blockIndex = GetTermSequenceAndBlockIndex(34, 20, 10)
	assert.Equal(t, int64(1), termSeq)
	assert.Equal(t, int64(4), blockIndex)
}
