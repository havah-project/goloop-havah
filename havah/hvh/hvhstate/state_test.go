package hvhstate

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/crypto"
	"github.com/icon-project/goloop/common/db"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/havah/hvhutils"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/state"
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

func checkAllPlanet(t *testing.T, s *State, expected int64) {
	cur := s.getVarDB(hvhmodule.VarAllPlanet).Int64()
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
	s := newDummyState()

	price = s.GetUSDTPrice()
	assert.Zero(t, price.Sign())

	// In case of valid prices
	prices := []*big.Int{new(big.Int), big.NewInt(1234)}
	for _, price = range prices {
		assert.NoError(t, s.SetUSDTPrice(price))
		price2 := s.GetUSDTPrice()
		assert.Zero(t, price.Cmp(price2))
	}

	// In case of invalid prices
	invalidPrices := []*big.Int{nil, big.NewInt(-1000)}
	for _, price = range invalidPrices {
		assert.Error(t, s.SetUSDTPrice(price))
	}
}

func TestState_AddPlanetManager(t *testing.T) {
	var err error
	var address module.Address

	s := newDummyState()
	addresses := []module.Address{
		common.MustNewAddressFromString("cx1"),
		common.MustNewAddressFromString("cx2"),
		common.MustNewAddressFromString("hx1"),
		common.MustNewAddressFromString("hx2"),
		common.MustNewAddressFromString("hx3"),
	}

	// Call AddPlanetManager() with a valid argument
	for _, address = range addresses {
		err = s.AddPlanetManager(address)
		assert.NoError(t, err)

		ok, err := s.IsPlanetManager(address)
		assert.True(t, ok)
		assert.NoError(t, err)
	}

	// Call AddPlanetManager with an invalid argument
	for _, address = range []module.Address{nil, addresses[0], addresses[1]} {
		err = s.AddPlanetManager(address)
		assert.Error(t, err)
	}

	// Remove the first item with RemovePlanetManager()
	address = addresses[0]
	err = s.RemovePlanetManager(address)
	assert.NoError(t, err)

	ok, err := s.IsPlanetManager(address)
	assert.NoError(t, err)
	assert.False(t, ok)

	for i := 1; i < len(addresses); i++ {
		ok, err = s.IsPlanetManager(addresses[i])
		assert.True(t, ok)
		assert.NoError(t, err)
	}

	// Remove the last item with RemovePlanetManager()
	address = addresses[len(addresses)-1]
	err = s.RemovePlanetManager(address)
	assert.NoError(t, err)

	ok, err = s.IsPlanetManager(address)
	assert.NoError(t, err)
	assert.False(t, ok)

	for i := 1; i < len(addresses)-1; i++ {
		ok, err = s.IsPlanetManager(addresses[i])
		assert.True(t, ok)
		assert.NoError(t, err)
	}

	// Invalid cases of RemovePlanetManager()
	for _, address = range []module.Address{nil, addresses[0], common.MustNewAddressFromString("hx1234")} {
		err = s.RemovePlanetManager(address)
		assert.Error(t, err)
	}

	ok, err = s.IsPlanetManager(nil)
	assert.Error(t, err)
	assert.False(t, ok)
}

func TestState_SetIssueStart(t *testing.T) {
	var startBH, curBH, height int64
	s := newDummyState()

	// Success case: startBH > 0 && startBH > curBH
	startBH, curBH = 2000, 1000
	err := s.SetIssueStart(curBH, startBH)
	assert.NoError(t, err)

	height = s.getVarDB(hvhmodule.VarIssueStart).Int64()
	assert.Equal(t, startBH, height)
	assert.Equal(t, startBH, s.GetIssueStart())

	// Failure case: startBH <= 0 || startBH <= curBH
	curBH = 1000
	height = s.getVarDB(hvhmodule.VarIssueStart).Int64()
	for _, startBH = range []int64{-100, 0, 100, 500, curBH} {
		err = s.SetIssueStart(curBH, startBH)
		assert.Error(t, err)

		startBH = s.getVarDB(hvhmodule.VarIssueStart).Int64()
		assert.Equal(t, height, startBH)
	}
}

func TestState_RegisterPlanet(t *testing.T) {
	s := newDummyState()
	owner := common.MustNewAddressFromString("hx1234")
	isCompany := true
	isPrivate := true
	usdt := big.NewInt(1000)
	price := new(big.Int).Mul(usdt, big.NewInt(10))
	height := int64(200)
	var planetCount int64
	rev := hvhmodule.RevisionPlanetIDReuse

	checkAllPlanet(t, s, 0)
	expectedCount := int64(0)
	for i := 0; i < 3; i++ {
		err := s.RegisterPlanet(rev, int64(i), isPrivate, isCompany, owner, usdt, price, height)
		assert.NoError(t, err)

		expectedCount++
		planetCount = s.getVarDB(hvhmodule.VarAllPlanet).Int64()
		assert.Equal(t, expectedCount, planetCount)
	}

	for i := 0; i < 2; i++ {
		lostDelta, err := s.UnregisterPlanet(rev, int64(i))
		assert.NoError(t, err)
		assert.Zero(t, lostDelta.Sign())
		expectedCount--
		checkAllPlanet(t, s, expectedCount)
	}

	planetCount = s.getVarDB(hvhmodule.VarAllPlanet).Int64()
	lostDelta, err := s.UnregisterPlanet(rev, int64(100))
	assert.Error(t, err)
	assert.Nil(t, lostDelta)
	checkAllPlanet(t, s, planetCount)
}

func TestState_UnregisterPlanet_InAbnormalCase(t *testing.T) {
	rev := hvhmodule.RevisionPlanetIDReuse
	s := newDummyState()
	checkAllPlanet(t, s, int64(0))
	lostDelta, err := s.UnregisterPlanet(rev, int64(-1))
	assert.Error(t, err)
	assert.Nil(t, lostDelta)

	lostDelta, err = s.UnregisterPlanet(rev, int64(100))
	assert.Error(t, err)
	assert.Nil(t, lostDelta)

	checkAllPlanet(t, s, int64(0))
}

func TestState_GetBigInt(t *testing.T) {
	key := "key"
	s := newDummyState()
	varDB := s.getVarDB(key)

	assert.Nil(t, varDB.BigInt())

	value := s.getBigInt(key)
	assert.NotNil(t, value)

	defValue := hvhmodule.BigIntHooverBudget
	value = s.getBigIntOrDefault(key, defValue)
	assert.Zero(t, defValue.Cmp(value))

	newValue := new(big.Int).Add(defValue, big.NewInt(100))
	err := s.setBigInt(key, newValue)
	assert.NoError(t, err)

	value = s.getBigInt(key)
	assert.Zero(t, newValue.Cmp(value))

	value = s.getBigIntOrDefault(key, defValue)
	assert.Zero(t, value.Cmp(newValue))
}

func TestState_GetPlanet(t *testing.T) {
	id := int64(1)

	s := newDummyState()
	p, err := s.GetPlanet(id)
	assert.Nil(t, p)
	assert.Error(t, err)

	dictDB := s.getDictDB(hvhmodule.DictPlanet, 1)

	p = newDummyPlanet(false, false, 1234)
	err = s.setPlanet(dictDB, id, p)
	assert.NoError(t, err)

	p2, err := s.getPlanet(dictDB, id)
	assert.NotNil(t, p)
	assert.NoError(t, err)
	assert.True(t, p.equal(p2))
}

func TestState_GetPlanetReward(t *testing.T) {
	id := int64(1)
	reward := big.NewInt(100)

	s := newDummyState()
	pr, err := s.GetPlanetReward(id)
	assert.NotNil(t, pr)
	assert.NoError(t, err)
	assert.Zero(t, pr.Total().Sign())
	assert.Zero(t, pr.Current().Sign())
	assert.Zero(t, pr.LastTermNumber())

	err = pr.increment(10, reward, reward)
	assert.NoError(t, err)

	err = s.setPlanetReward(id, pr)
	assert.NoError(t, err)

	pr2, err := s.GetPlanetReward(id)
	assert.NotNil(t, pr2)
	assert.NoError(t, err)
	assert.True(t, pr.equal(pr2))
}

func TestState_IncrementWorkingPlanet(t *testing.T) {
	s := newDummyState()

	ov := s.getInt64(hvhmodule.VarWorkingPlanet)
	err := s.IncrementWorkingPlanet()
	assert.NoError(t, err)

	nv := s.getInt64(hvhmodule.VarWorkingPlanet)
	assert.Equal(t, ov+1, nv)
}

func TestState_InitState(t *testing.T) {
	s := newDummyState()

	assert.Equal(t, int64(hvhmodule.TermPeriod), s.GetTermPeriod())
	assert.Zero(t, s.GetHooverBudget().Cmp(hvhmodule.BigIntHooverBudget))
	assert.Equal(t, int64(hvhmodule.IssueReductionCycle), s.GetIssueReductionCycle())
	assert.Zero(t, s.GetIssueStart())
	assert.Zero(t, s.GetIssueReductionRate().Cmp(hvhmodule.BigRatIssueReductionRate))

	cfg := &StateConfig{}
	cfg.TermPeriod = &common.HexInt64{Value: 100}
	cfg.HooverBudget = common.NewHexInt(1_000_000)
	cfg.IssueReductionCycle = &common.HexInt64{Value: 180}
	cfg.USDTPrice = common.NewHexInt(200_000)

	assert.NoError(t, s.InitState(cfg))
	assert.Equal(t, cfg.TermPeriod.Value, s.GetTermPeriod())
	assert.Zero(t, s.GetHooverBudget().Cmp(cfg.HooverBudget.Value()))
	assert.Equal(t, cfg.IssueReductionCycle.Value, s.GetIssueReductionCycle())
	assert.Zero(t, s.GetIssueReductionRate().Cmp(hvhmodule.BigRatIssueReductionRate))
}

func TestState_DecreaseRewardRemain(t *testing.T) {
	const initRewardRemain = 1_000_000
	const key = hvhmodule.VarRewardRemain
	s := newDummyState()

	safeSetBigInt(t, s, key, big.NewInt(initRewardRemain))

	amount := int64(1000)
	err := s.DecreaseRewardRemain(big.NewInt(amount))
	assert.NoError(t, err)

	newRewardRemain := s.getBigInt(key)
	assert.Zero(t, big.NewInt(initRewardRemain).Cmp(new(big.Int).Add(newRewardRemain, big.NewInt(amount))))
}

func TestState_calcClaimableReward(t *testing.T) {
	height := int64(100)
	owner := common.MustNewAddressFromString("hx12345")
	usdt := toUSDT(5000)
	price := toHVH(50000)
	p := newPlanet(false, false, owner, usdt, price, 100)
	pr := newEmpyPlanetReward()
	s := newDummyState()

	amount := toHVH(1)
	err := pr.increment(1, amount, amount)
	assert.NoError(t, err)

	p2 := deepCopyPlanet(t, p)
	pr2 := deepCopyPlanetReward(t, pr)

	reward, err := s.calcClaimableReward(height, p, pr)
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

	s := newDummyState()

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
		err := s.setPlanetReward(int64(i), pr)
		assert.NoError(t, err)
	}

	planetDictDB := s.getDictDB(hvhmodule.DictPlanet, 1)
	for i := 1; i < 4; i++ {
		p := deepCopyPlanet(t, ps[i])
		err := s.setPlanet(planetDictDB, int64(i), p)
		assert.NoError(t, err)
	}

	err := s.SetIssueStart(1, issueStart)
	assert.NoError(t, err)

	height := s.GetIssueStart() + s.GetTermPeriod()
	for _, id := range []int64{publicId, companyId} {
		reward, err := s.ClaimPlanetReward(id, height, owner)
		assert.NoError(t, err)
		assert.Zero(t, reward.Cmp(prs[id].Current()))

		pr, err := s.GetPlanetReward(id)
		assert.NoError(t, err)
		assert.Zero(t, pr.Total().Cmp(prs[id].Total()))
		assert.Zero(t, pr.Current().Sign())
		assert.Equal(t, pr.LastTermNumber(), prs[id].LastTermNumber())
	}

	id := int64(privateId)
	reward, err := s.ClaimPlanetReward(id, height, owner)
	assert.NoError(t, err)
	assert.Zero(t, reward.Sign())
}

func TestState_SetPrivateClaimableRate(t *testing.T) {
	var err error
	var num, denom int64
	var expNum, expDenom int64
	s := newDummyState()

	// Check default value
	expNum = int64(0)
	expDenom = int64(hvhmodule.PrivateClaimableRate)
	num, denom = s.GetPrivateClaimableRate()
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
		err = s.SetPrivateClaimableRate(in[0], in[1])
		assert.Error(t, err)

		num, denom = s.GetPrivateClaimableRate()
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
		err = s.SetPrivateClaimableRate(expNum, expDenom)
		assert.NoError(t, err)

		num, denom = s.GetPrivateClaimableRate()
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

func TestState_Lost(t *testing.T) {
	var err error
	var lost, amount *big.Int
	expLost := new(big.Int)
	s := newDummyState()

	lost, err = s.GetLost()
	assert.NoError(t, err)
	assert.Zero(t, lost.Sign())

	for i := 0; i < 3; i++ {
		amount = big.NewInt(100)
		expLost.Add(expLost, amount)

		err = s.addLost(amount)
		assert.NoError(t, err)
		lost, err = s.GetLost()
		assert.NoError(t, err)
		assert.Zero(t, lost.Cmp(expLost))
	}

	lost, err = s.DeleteLost()
	assert.NoError(t, err)
	assert.Zero(t, lost.Cmp(expLost))

	lost, err = s.GetLost()
	assert.NoError(t, err)
	assert.Zero(t, lost.Sign())

	amount = big.NewInt(10)
	err = s.addLost(amount)
	assert.NoError(t, err)
	lost, err = s.GetLost()
	assert.NoError(t, err)
	assert.Zero(t, lost.Cmp(amount))
}

func TestState_UnregisterPlanet(t *testing.T) {
	var err error
	rev := hvhmodule.Revision2
	usdt := big.NewInt(10)
	price := big.NewInt(1)
	expLost := new(big.Int)

	type planet struct {
		isPrivate bool
		isCompany bool
		owner     module.Address
	}
	planets := []planet{
		{true, false, common.MustNewAddressFromString("hx1")},
		{false, true, common.MustNewAddressFromString("hx2")},
		{false, false, common.MustNewAddressFromString("hx3")},
	}
	rewards := []*big.Int{
		big.NewInt(rand.Int63()),
		big.NewInt(rand.Int63()),
		big.NewInt(rand.Int63()),
	}

	s := newDummyState()

	for i := 0; i < len(planets); i++ {
		id := int64(i + 1)
		p := planets[i]
		err = s.RegisterPlanet(
			rev,
			id, p.isPrivate, p.isCompany, p.owner, usdt, price, 5)
		assert.NoError(t, err)

		pr, err := s.GetPlanetReward(id)
		assert.NoError(t, err)
		assert.Zero(t, pr.Total().Sign())
		assert.Zero(t, pr.Current().Sign())

		reward := rewards[i]
		err = s.OfferReward(1, id, pr, reward, reward)
		assert.NoError(t, err)

		expLost.Add(expLost, reward)
	}

	for i := 0; i < len(planets); i++ {
		id := int64(i + 1)
		lostDelta, err := s.UnregisterPlanet(rev, id)
		assert.NoError(t, err)
		assert.Zero(t, lostDelta.Cmp(rewards[i]))

		pr, err := s.GetPlanetReward(id)
		assert.NoError(t, err)
		assert.Zero(t, pr.Total().Sign())
		assert.Zero(t, pr.Current().Sign())
	}

	lost, err := s.GetLost()
	assert.NoError(t, err)
	assert.True(t, lost.Sign() > 0)
	assert.Zero(t, expLost.Cmp(lost))
}

func TestState_RegisterValidator(t *testing.T) {
	owner := newDummyAddress(1, false)
	name := "name-01"
	url := fmt.Sprintf("https://www.%s.com/details.json", name)
	_, pubKey := crypto.GenerateKeyPair()
	s := newDummyState()

	err := s.RegisterValidator(owner, pubKey.SerializeCompressed(), GradeSub, name, &url)
	assert.NoError(t, err)

	err = s.RegisterValidator(owner, pubKey.SerializeCompressed(), GradeSub, name, &url)
	assert.Error(t, err)

}

func TestState_UnregisterValidator(t *testing.T) {
	var node module.Address
	owner := newDummyAddress(1, false)
	name := "name-01"
	url := fmt.Sprintf("https://www.%s.com/details.json", name)
	_, pubKey := crypto.GenerateKeyPair()
	s := newDummyState()

	err := s.RegisterValidator(owner, pubKey.SerializeCompressed(), GradeSub, name, &url)
	assert.NoError(t, err)

	vi, err := s.GetValidatorInfo(owner)
	assert.NoError(t, err)
	assert.True(t, vi.Owner().Equal(owner))

	vs, err := s.GetValidatorStatus(owner)
	assert.NoError(t, err)
	assert.False(t, vs.Disabled())
	assert.False(t, vs.Disqualified())
	assert.True(t, vs.Enabled())

	owners, err := s.GetDisqualifiedValidators()
	assert.Zero(t, len(owners))
	assert.NoError(t, err)

	// Try to unregister a not-registered validator
	invalidOwner := newDummyAddress(2, false)
	node, err = s.UnregisterValidator(invalidOwner)
	assert.Error(t, err)
	assert.Nil(t, node)

	// Success case
	node, err = s.UnregisterValidator(owner)
	assert.NoError(t, err)
	assert.True(t, node.Equal(vi.Address()))

	vs, err = s.GetValidatorStatus(owner)
	assert.True(t, vs.Disqualified())

	owners, err = s.GetDisqualifiedValidators()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(owners))
	assert.True(t, owners[0].Equal(owner))
}

func TestState_SetValidatorInfo(t *testing.T) {
	owner := newDummyAddress(1, false)
	name := "name-01"
	url := fmt.Sprintf("https://www.%s.com/details.json", name)
	_, pubKey := crypto.GenerateKeyPair()
	s := newDummyState()

	err := s.RegisterValidator(owner, pubKey.SerializeCompressed(), GradeSub, name, &url)
	assert.NoError(t, err)

	newName := "newName"
	newUrl := "http://www.example.com/details.json"
	// _, newPubKey := crypto.GenerateKeyPair()
	// newNodePublicKey := "0x" + hex.EncodeToString(newPubKey.SerializeCompressed())
	values := make(map[string]string)
	values["name"] = newName
	values["url"] = newUrl
	// values["nodePublicKey"] = newNodePublicKey

	err = s.SetValidatorInfo(owner, values)
	assert.NoError(t, err)

	vi, err := s.GetValidatorInfo(owner)
	assert.NoError(t, err)
	assert.Equal(t, newName, vi.Name())
	assert.Equal(t, newUrl, vi.Url())
	// assert.True(t, newPubKey.Equal(vi.PublicKey()))
}

func TestState_EnableValidator(t *testing.T) {
	owner := newDummyAddress(1, false)
	name := "name-01"
	url := fmt.Sprintf("https://www.%s.com/details.json", name)
	_, pubKey := crypto.GenerateKeyPair()
	s := newDummyState()

	err := s.RegisterValidator(owner, pubKey.SerializeCompressed(), GradeSub, name, &url)
	assert.NoError(t, err)

	vs, err := s.GetValidatorStatus(owner)
	enableCount := vs.EnableCount()
	assert.NoError(t, err)
	assert.Equal(t, hvhmodule.MaxEnableCount, enableCount)

	// Called by an invalid owner
	NoOwner := newDummyAddress(2, false)
	err = s.EnableValidator(NoOwner, false)
	assert.Error(t, err)
	assert.Equal(t, enableCount, vs.EnableCount())

	// Case where calling EnableValidator() to an enabled validator
	err = s.EnableValidator(owner, false)
	assert.NoError(t, err)
	assert.Equal(t, hvhmodule.MaxEnableCount, vs.EnableCount())

	for i := 1; i <= hvhmodule.MaxEnableCount; i++ {
		assert.NoError(t, s.DisableValidator(owner))
		vs, err = s.GetValidatorStatus(owner)
		assert.NoError(t, err)
		assert.True(t, vs.Disabled())

		err = s.EnableValidator(owner, false)
		assert.NoError(t, err)

		vs, err = s.GetValidatorStatus(owner)
		assert.NoError(t, err)
		assert.Equal(t, hvhmodule.MaxEnableCount-i, vs.EnableCount())
		assert.True(t, vs.Enabled())
	}

	assert.NoError(t, s.DisableValidator(owner))
	vs, err = s.GetValidatorStatus(owner)
	assert.NoError(t, err)
	assert.True(t, vs.Disabled())

	err = s.EnableValidator(owner, false)
	assert.Error(t, err)

	vs, err = s.GetValidatorStatus(owner)
	assert.NoError(t, err)
	assert.True(t, vs.Disabled())
	assert.Zero(t, vs.EnableCount())
}

func TestState_SetValidatorCount(t *testing.T) {
	s := newDummyState()
	count := s.GetActiveValidatorCount()
	assert.Zero(t, count)

	newCount := int64(10)
	err := s.SetActiveValidatorCount(newCount)
	assert.NoError(t, err)
	count = s.GetActiveValidatorCount()
	assert.Equal(t, newCount, count)

	assert.Error(t, s.SetActiveValidatorCount(0))
	count = s.GetActiveValidatorCount()
	assert.Equal(t, newCount, count)
}

func TestState_IsDecentralizationPossible(t *testing.T) {
	var err error
	s := newDummyState()

	// Decentralization is not possible if revision is less than hvhmodule.RevisionDecentralization
	for rev := 0; rev < hvhmodule.RevisionDecentralization; rev++ {
		assert.False(t, s.IsDecentralizationPossible(rev))
	}

	rev := hvhmodule.RevisionDecentralization

	validatorCount := s.GetActiveValidatorCount()
	assert.Zero(t, validatorCount)

	validatorCount = 10
	err = s.SetActiveValidatorCount(validatorCount)
	assert.NoError(t, err)
	assert.False(t, s.IsDecentralizationPossible(rev))

	for i := 0; i < int(validatorCount); i++ {
		name := fmt.Sprintf("name-%02d", i)
		url := fmt.Sprintf("https://www.%s.com/details.json", name)
		owner := newDummyAddress(i+1, false)
		_, pubKey := crypto.GenerateKeyPair()
		err = s.RegisterValidator(owner, pubKey.SerializeCompressed(), GradeSub, name, &url)
		assert.NoError(t, err)
	}

	assert.True(t, s.IsDecentralizationPossible(rev))
}

func TestState_GetValidatorsOf(t *testing.T) {
	var err error
	var validators []module.Address
	mainCount := 7
	subCount := 5
	validatorCount := mainCount + subCount
	mainOwners := make([]module.Address, 0, mainCount)
	subOwners := make([]module.Address, 0, subCount)

	s := newDummyState()

	for i := 0; i < mainCount; i++ {
		name := fmt.Sprintf("name-%02d", i)
		owner := newDummyAddress(i+1, false)
		_, pubKey := crypto.GenerateKeyPair()
		url := fmt.Sprintf("https://www.%s.com/details.json", name)

		err = s.RegisterValidator(owner, pubKey.SerializeCompressed(), GradeMain, name, &url)
		assert.NoError(t, err)

		mainOwners = append(mainOwners, owner)
	}
	for i := mainCount; i < validatorCount; i++ {
		name := fmt.Sprintf("name-%02d", i)
		owner := newDummyAddress(i+1, false)
		_, pubKey := crypto.GenerateKeyPair()
		url := fmt.Sprintf("https://www.%s.com/details.json", name)

		err = s.RegisterValidator(owner, pubKey.SerializeCompressed(), GradeSub, name, &url)
		assert.NoError(t, err)

		subOwners = append(subOwners, owner)
	}

	allOwners := make([]module.Address, 0, mainCount+subCount)
	allOwners = append(allOwners, mainOwners...)
	allOwners = append(allOwners, subOwners...)
	ownersOf := [][]module.Address{subOwners, mainOwners, allOwners}

	for _, gFilter := range []GradeFilter{GradeFilterSub, GradeFilterMain, GradeFilterAll} {
		owners := ownersOf[gFilter]
		validators, err = s.GetValidatorsOf(gFilter)
		assert.NoError(t, err)
		assert.Equal(t, len(owners), len(validators))
		for i, v := range validators {
			assert.True(t, v.Equal(owners[i]))
		}
	}

	// Unregister a main validator
	idx := 1
	ownerToRemove := mainOwners[idx]
	_, err = s.UnregisterValidator(ownerToRemove)
	assert.NoError(t, err)
	mainOwners = append(mainOwners[:idx], mainOwners[idx+1:]...)
	assert.Equal(t, mainCount-1, len(mainOwners))

	validators, err = s.GetValidatorsOf(GradeFilterMain)
	assert.NoError(t, err)
	assert.Equal(t, len(mainOwners), len(validators))
	for i := 0; i < len(validators); i++ {
		assert.False(t, validators[i].Equal(ownerToRemove))
	}

	// Unregister a sub validator
	idx = 1
	ownerToRemove = subOwners[1]
	_, err = s.UnregisterValidator(ownerToRemove)
	assert.NoError(t, err)
	subOwners = append(subOwners[:idx], subOwners[idx+1:]...)
	assert.Equal(t, subCount-1, len(subOwners))

	validators, err = s.GetValidatorsOf(GradeFilterSub)
	assert.NoError(t, err)
	assert.Equal(t, subCount-1, len(validators))
	for i := 0; i < len(validators); i++ {
		assert.False(t, validators[i].Equal(ownerToRemove))
	}

	// Unregister all main validators
	for _, owner := range mainOwners {
		_, err = s.UnregisterValidator(owner)
		assert.NoError(t, err)
	}
	validators, err = s.GetValidatorsOf(GradeFilterMain)
	assert.NoError(t, err)
	assert.Zero(t, len(validators))

	// Unregister all sub validators
	for _, owner := range subOwners {
		_, err = s.UnregisterValidator(owner)
		assert.NoError(t, err)
	}
	validators, err = s.GetValidatorsOf(GradeFilterSub)
	assert.NoError(t, err)
	assert.Zero(t, len(validators))

	validators, err = s.GetValidatorsOf(GradeFilterAll)
	assert.NoError(t, err)
	assert.Zero(t, len(validators))
}

func TestState_OnBlockVote(t *testing.T) {
	var err error
	var grade Grade
	var vi *ValidatorInfo
	var vs *ValidatorStatus
	var owner module.Address
	var idx int

	mainCount := 7
	subCount := 3
	validatorCount := mainCount + subCount

	s := newDummyState()
	assert.NoError(t, s.SetDecentralized())
	assert.NoError(t, s.RenewNetworkStatusOnTermStart())

	owners := make([]module.Address, validatorCount)
	validators := make([]module.Address, validatorCount)

	for i := 0; i < validatorCount; i++ {
		name := fmt.Sprintf("name-%02d", i)
		owner = newDummyAddress(i+1, false)
		_, pubKey := crypto.GenerateKeyPair()
		url := fmt.Sprintf("https://www.%s.com/details.json", name)

		if i < mainCount {
			grade = GradeMain
		} else {
			grade = GradeSub
		}

		err = s.RegisterValidator(owner, pubKey.SerializeCompressed(), grade, name, &url)
		assert.NoError(t, err)

		vi, err = s.GetValidatorInfo(owner)
		assert.NoError(t, err)

		owners[i] = owner
		validators[i] = vi.Address()
	}

	var penalized bool
	for i, v := range validators {
		penalized, owner, err = s.OnBlockVote(v, true)
		assert.NoError(t, err)
		assert.False(t, penalized)
		assert.True(t, owners[i].Equal(owner))

		owner, err = s.GetOwnerByNode(v)
		assert.NoError(t, err)

		vs, err = s.GetValidatorStatus(owner)
		assert.NoError(t, err)
		assert.Zero(t, vs.NonVotes())
		assert.True(t, vs.Enabled())
		assert.Equal(t, hvhmodule.MaxEnableCount, vs.EnableCount())
	}

	// Case: a main validator does not get penalized even though its nonVotes is larger than nonVoteAllowance
	idx = 0
	node := validators[idx]
	size := int(hvhmodule.NonVoteAllowance) + 1
	owner, err = s.GetOwnerByNode(node)
	assert.NoError(t, err)
	for i := 0; i < size; i++ {
		penalized, owner2, err := s.OnBlockVote(node, false)
		assert.NoError(t, err)
		assert.False(t, penalized)
		assert.True(t, owner2.Equal(owner))

		vs, err = s.GetValidatorStatus(owner)
		assert.NoError(t, err)
		assert.Equal(t, int64(i + 1), vs.NonVotes())
		assert.True(t, vs.Enabled())
		assert.Equal(t, hvhmodule.MaxEnableCount, vs.EnableCount())
	}

	// Case: a sub validator gets penalized if its nonVotes is larger than nonVoteAllowance
	idx = mainCount
	node = validators[idx]
	size = int(hvhmodule.NonVoteAllowance) + 1
	owner = owners[idx]
	assert.NoError(t, err)
	for i := 0; i < size; i++ {
		expectedPenalized := i == size - 1

		penalized, owner2, err := s.OnBlockVote(node, false)
		assert.NoError(t, err)
		assert.Equal(t, expectedPenalized, penalized)
		assert.True(t, owner2.Equal(owner))

		vs, err = s.GetValidatorStatus(owner)
		assert.NoError(t, err)
		assert.Equal(t, int64(i + 1), vs.NonVotes())
		assert.Equal(t, !expectedPenalized, vs.Enabled())
		assert.Equal(t, hvhmodule.MaxEnableCount, vs.EnableCount())
	}
}

func TestState_GetMainValidators(t *testing.T) {
	var err error
	s := newDummyState()
	size := 7
	owners := make([]module.Address, size)
	expValidators := make([]module.Address, size)

	for i := 0; i < size; i++ {
		name := fmt.Sprintf("name-%02d", i)
		owner := newDummyAddress(i+1, false)
		_, pubKey := crypto.GenerateKeyPair()
		url := fmt.Sprintf("https://www.%s.com/details.json", name)

		err = s.RegisterValidator(owner, pubKey.SerializeCompressed(), GradeMain, name, &url)
		assert.NoError(t, err)

		expValidators[i] = common.NewAccountAddressFromPublicKey(pubKey)
		owners[i] = owner
	}

	count := 5
	validators, err := s.GetMainValidators(count)
	assert.NoError(t, err)
	assert.Equal(t, count, len(validators))

	for i := 0; i < count; i++ {
		assert.True(t, validators[i].Equal(expValidators[i]))
	}

	idx := size / 2
	_, err = s.UnregisterValidator(owners[idx])
	assert.NoError(t, err)

	count = 10
	validators, err = s.GetMainValidators(count)
	assert.NoError(t, err)
	assert.Equal(t, size-1, len(validators))

	addrMap := make(map[string]struct{})
	for _, v := range validators {
		addrMap[ToKey(v)] = struct{}{}
	}

	expValidators = append(expValidators[:idx], expValidators[idx+1:]...)
	for _, ev := range expValidators {
		_, found := addrMap[ToKey(ev)]
		assert.True(t, found)
	}
}

type dummyValidatorState struct {
	state.ValidatorState
	validators   []module.Address
	validatorMap map[string]int
}

func (vs *dummyValidatorState) IndexOf(addr module.Address) int {
	if vs.validatorMap == nil {
		vs.validatorMap = make(map[string]int)
		for i, v := range vs.validators {
			vs.validatorMap[ToKey(v)] = i
		}
	}
	if idx, ok := vs.validatorMap[ToKey(addr)]; ok {
		return idx
	} else {
		return -1
	}
}

func (vs *dummyValidatorState) Get(i int) (module.Validator, bool) {
	if i >= 0 && i < vs.Len() {
		addr := vs.validators[i]
		v, _ := state.ValidatorFromAddress(addr)
		return v, true
	}
	return nil, false
}

func (vs *dummyValidatorState) Len() int {
	return len(vs.validators)
}

func newDummyValidatorState(validators []module.Address) state.ValidatorState {
	return &dummyValidatorState{
		validators: validators,
	}
}

func TestState_GetNextActiveValidatorsAndChangeIndex(t *testing.T) {
	var err error
	s := newDummyState()
	size := 7
	owners := make([]module.Address, size)
	expValidators := make([]module.Address, size)

	for i := 0; i < size; i++ {
		name := fmt.Sprintf("name-%02d", i+1)
		owner := newDummyAddress(i+1, false)
		_, pubKey := crypto.GenerateKeyPair()
		url := fmt.Sprintf("https://www.%s.com/details.json", name)

		err = s.RegisterValidator(owner, pubKey.SerializeCompressed(), GradeSub, name, &url)
		assert.NoError(t, err)

		expValidators[i] = common.NewAccountAddressFromPublicKey(pubKey)
		owners[i] = owner
	}

	type arg struct {
		count, expLen, expSVIndex int
	}
	args := []arg{
		{count: -1, expLen: 0, expSVIndex: 0},
		{count: 0, expLen: 0, expSVIndex: 0},
		{count: 1, expLen: 1, expSVIndex: 4},
		{count: 2, expLen: 2, expSVIndex: 6},
		{count: 2, expLen: 2, expSVIndex: 4},
	}

	validators, err := s.GetNextActiveValidatorsAndChangeIndex(nil, size)
	assert.NoError(t, err)
	assert.Equal(t, size, len(validators))
	for i := 0; i < size; i++ {
		assert.True(t, expValidators[i].Equal(validators[i]))
	}
	assert.Zero(t, s.GetSubValidatorsIndex())

	activeValidators := newDummyValidatorState(expValidators[:3])
	for i, test := range args {
		name := fmt.Sprintf("no-disabled-%d", i)
		t.Run(name, func(t *testing.T) {
			validators, err = s.GetNextActiveValidatorsAndChangeIndex(activeValidators, test.count)
			if test.count >= 0 {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
			assert.Equal(t, test.expLen, len(validators))
			assert.Equal(t, test.expSVIndex, int(s.GetSubValidatorsIndex()))
			for _, v := range validators {
				assert.True(t, activeValidators.IndexOf(v) < 0)
				owner, err := s.GetOwnerByNode(v)
				assert.NoError(t, err)
				vs, err := s.GetValidatorStatus(owner)
				assert.NoError(t, err)
				assert.True(t, vs.Enabled())
			}
		})
	}

	args = []arg{
		{count: 3, expLen: 3, expSVIndex: 4},
		{count: 1, expLen: 1, expSVIndex: 6},
	}
	owner, _ := s.GetOwnerByNode(expValidators[4])
	assert.NoError(t, s.DisableValidator(owner))
	for i, test := range args {
		name := fmt.Sprintf("one-disabled-%d", i)
		t.Run(name, func(t *testing.T) {
			validators, err = s.GetNextActiveValidatorsAndChangeIndex(activeValidators, test.count)
			assert.NoError(t, err)
			assert.Equal(t, test.expLen, len(validators))
			assert.Equal(t, test.expSVIndex, int(s.GetSubValidatorsIndex()))
			for _, v := range validators {
				assert.True(t, activeValidators.IndexOf(v) < 0)
				owner, _ = s.GetOwnerByNode(v)
				vs, _ := s.GetValidatorStatus(owner)
				assert.True(t, vs.Enabled())
			}
		})
	}

	assert.Equal(t, 6, int(s.GetSubValidatorsIndex()))
	args = []arg{
		{count: 1, expLen: 1, expSVIndex: 4},
	}
	owner, _ = s.GetOwnerByNode(expValidators[4])
	_, err = s.UnregisterValidator(owner)
	assert.NoError(t, err)
	// sub_validators: |0|1|2|3|5|6|
	for i, test := range args {
		name := fmt.Sprintf("one-unregistered-%d", i)
		t.Run(name, func(t *testing.T) {
			validators, err = s.GetNextActiveValidatorsAndChangeIndex(activeValidators, test.count)
			assert.NoError(t, err)
			assert.Equal(t, test.expLen, len(validators))
			assert.Equal(t, test.expSVIndex, int(s.GetSubValidatorsIndex()))
			for _, v := range validators {
				assert.True(t, activeValidators.IndexOf(v) < 0)
				owner, _ = s.GetOwnerByNode(v)
				vs, _ := s.GetValidatorStatus(owner)
				assert.True(t, vs.Enabled())
			}
		})
	}
}

func TestState_SetBlockVoteCheckParameters(t *testing.T) {
	s := newDummyState()
	assert.Equal(t, hvhmodule.BlockVoteCheckPeriod, s.GetBlockVoteCheckPeriod())
	assert.Equal(t, hvhmodule.NonVoteAllowance, s.GetNonVoteAllowance())

	tests := []struct{
		period int64
		allowance int64
		success bool
	}{
		{0, 0, true},
		{0, 100, true},
		{100, 0, true},
		{100, 70, true},
		{200, 80, true},
		{-10, 80, false},
		{10, -80, false},
		{-100, -80, false},
	}

	for i, test := range tests {
		period := test.period
		allowance := test.allowance
		success := test.success
		name := fmt.Sprintf("test%02d_%d_%d_%t", i, period, allowance, success)
		t.Run(name, func(t *testing.T){
			oPeriod := s.GetBlockVoteCheckPeriod()
			oAllowance := s.GetNonVoteAllowance()

			err := s.SetBlockVoteCheckParameters(period, allowance)
			assert.Equal(t, success, err == nil)
			if success {
				assert.Equal(t, period, s.GetBlockVoteCheckPeriod())
				assert.Equal(t, allowance, s.GetNonVoteAllowance())
			} else {
				assert.Equal(t, oPeriod, s.GetBlockVoteCheckPeriod())
				assert.Equal(t, oAllowance, s.GetNonVoteAllowance())
			}
		})
	}
}

func TestState_SetNetworkStatus(t *testing.T) {
	s := newDummyState()
	emptyNS := NewNetworkStatus()
	ns, err := s.GetNetworkStatus()
	assert.NoError(t, err)
	assert.True(t, ns.Equal(emptyNS))

	ns.SetDecentralized()
	assert.NoError(t, ns.SetActiveValidatorCount(25))
	assert.NoError(t, ns.SetNonVoteAllowance(100))
	assert.NoError(t, ns.SetBlockVoteCheckPeriod(100))
	assert.NoError(t, s.SetNetworkStatus(ns))

	ns2, err := s.GetNetworkStatus()
	assert.NoError(t, err)
	assert.True(t, ns != ns2)
	assert.True(t, ns2.Equal(ns))
}

func TestState_IsItTimeToCheckBlockVote(t *testing.T) {
	s := newDummyState()
	tests := []struct{
		blockIndex int64
		mode NetMode
		period int64
		allowance int64
		result bool
	}{
		{0, NetModeInit, 0, 0, false},
		{0, NetModeDecentralized, 0, 5, false},
		{0, NetModeInit, 10, 5, false},
		{7, NetModeDecentralized, 10, 5, false},
		{0, NetModeDecentralized, 10, 0, true},
		{10, NetModeDecentralized, 10, 0, true},
		{20, NetModeDecentralized, 20, 3, true},
		{40, NetModeDecentralized, 20, 3, true},
	}

	ns, _ := s.GetNetworkStatus()
	for i, test := range tests {
		name := fmt.Sprintf("name-%02d", i)
		t.Run(name, func(t *testing.T){
			ns.SetMode(test.mode)
			assert.NoError(t, ns.SetBlockVoteCheckPeriod(test.period))
			assert.NoError(t, ns.SetNonVoteAllowance(test.allowance))
			assert.NoError(t, s.SetNetworkStatus(ns))
			assert.Equal(t, test.result, s.IsItTimeToCheckBlockVote(test.blockIndex))
		})
	}
}

func TestState_RenewNetworkStatusOnTermStart(t *testing.T) {
	s := newDummyState()
	period := int64(10)
	allowance := int64(3)
	avCount := int64(10)

	oldNs, _ := s.GetNetworkStatus()
	assert.False(t, oldNs.IsDecentralized())

	assert.NoError(t, s.SetBlockVoteCheckParameters(period, allowance))
	assert.NoError(t, s.SetActiveValidatorCount(avCount))
	assert.NoError(t, s.RenewNetworkStatusOnTermStart())
	ns, _ := s.GetNetworkStatus()
	assert.True(t, ns.Equal(oldNs))

	ns.SetDecentralized()
	assert.NoError(t, s.SetNetworkStatus(ns))

	assert.NoError(t, s.RenewNetworkStatusOnTermStart())
	ns, _ = s.GetNetworkStatus()
	assert.False(t, ns.Equal(oldNs))
}

func TestState_GetOwnerByNode(t *testing.T) {
	size := 5
	owners := newDummyAddresses(1, false, size)
	nodes := newDummyAddresses(100, false, size)
	s := newDummyState()

	// Register node address: (node : owner)
	for i := 0; i < size; i++ {
		err := s.addNodeToOwnerMap(nodes[i], owners[i])
		assert.NoError(t, err)
	}

	// Well registered
	for i, node := range nodes {
		owner, err := s.GetOwnerByNode(node)
		assert.NoError(t, err)
		assert.True(t, owner.Equal(owners[i]))
	}

	// Change node address: (owner : owner)
	for _, owner := range owners {
		node := owner
		err := s.addNodeToOwnerMap(node, owner)
		assert.NoError(t, err)

		owner, err = s.GetOwnerByNode(node)
		assert.NoError(t, err)
		assert.True(t, owner.Equal(node))
	}

	// Failed if a duplicate node address is used
	for i, node := range nodes {
		err := s.addNodeToOwnerMap(node, owners[i])
		assert.Error(t, err)
		assert.False(t, node.Equal(owners[i]))
	}
}

func TestState_SetNodePublicKey(t *testing.T) {
	var err error
	var idx int

	size := 5
	owners := newDummyAddresses(1, false, size)
	nodes := make([]module.Address, size)
	var publicKeys []*crypto.PublicKey
	s := newDummyState()

	// Register validators
	for i := 0; i < size; i++ {
		name := fmt.Sprintf("name-%02d", i)
		owner := owners[i]
		_, publicKey := crypto.GenerateKeyPair()
		publicKeys = append(publicKeys, publicKey)
		url := fmt.Sprintf("https://www.%s.com/details.json", name)

		err = s.RegisterValidator(owner, publicKey.SerializeCompressed(), GradeSub, name, &url)
		assert.NoError(t, err)

		nodes[i] = common.NewAccountAddressFromPublicKey(publicKey)
	}

	// Error case: non-existent owner
	owner := newDummyAddress(1234, false)
	_, publicKey := crypto.GenerateKeyPair()
	oNode, nNode, err := s.SetNodePublicKey(owner, publicKey.SerializeCompressed())
	assert.Nil(t, oNode)
	assert.Nil(t, nNode)
	assert.Error(t, err)

	// Error case: already used node address
	idx = 1
	owner = owners[idx]
	publicKey = publicKeys[3]
	oNode, nNode, err = s.SetNodePublicKey(owner, publicKey.SerializeUncompressed())
	assert.Nil(t, oNode)
	assert.Nil(t, nNode)
	assert.Error(t, err)

	// Success case: Replace old node with the same one
	idx = 2
	owner = owners[idx]
	publicKey = publicKeys[idx]
	oNode, nNode, err = s.SetNodePublicKey(owner, publicKey.SerializeUncompressed())
	assert.True(t, oNode.Equal(nNode))
	assert.NoError(t, err)

	// Success case
	idx = 0
	owner = owners[idx]
	_, publicKey = crypto.GenerateKeyPair()
	oNode, nNode, err = s.SetNodePublicKey(owner, publicKey.SerializeUncompressed())
	vi, err := s.GetValidatorInfo(owner)
	assert. NoError(t, err)
	assert.True(t, oNode.Equal(nodes[idx]))
	assert.True(t, publicKey.Equal(vi.PublicKey()))
	assert.True(t, vi.Address().Equal(common.NewAccountAddressFromPublicKey(publicKey)))
	assert.False(t, vi.Address().Equal(oNode))
}
