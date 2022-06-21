package hvhstate

import (
	"math/big"
	"testing"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/db"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/module"
)

func newDummyState() *State {
	mdb := db.NewMapDB()
	snapshot := NewSnapshot(mdb, nil)
	return NewStateFromSnapshot(snapshot, false, nil)
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
	if cur != expected {
		t.Errorf("Incorrect all_planet count: cur(%d) != expected(%d)", cur, expected)
	}
}

func safeSetBigInt(t *testing.T, state *State, key string, value *big.Int) {
	var err error
	orgValue := new(big.Int).Set(value)

	if err = state.setBigInt(key, value); err != nil {
		t.Errorf(err.Error())
	}
	value = state.getBigInt(key)
	if orgValue.Cmp(value) != 0 {
		t.Errorf("state.setBigInt() error")
	}
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
	if !p2.equal(p) {
		t.Errorf("deepCopyPlanet() error")
	}
	return p2
}

func deepCopyPlanetReward(t *testing.T, pr *planetReward) *planetReward {
	pr2 := &planetReward{
		new(big.Int).Set(pr.Total()),
		pr.LastTermNumber(),
		new(big.Int).Set(pr.Current()),
	}
	if !pr2.equal(pr) {
		t.Errorf("deepCopyPlanetReward() error")
	}
	return pr2
}

func toUSDT(value int64) *big.Int {
	usdt := big.NewInt(value)
	return usdt.Mul(usdt, hvhmodule.BigIntUSDTDecimal)
}

func toHVH(value int64) *big.Int {
	hvh := big.NewInt(value)
	return hvh.Mul(hvh, hvhmodule.BigIntUSDTDecimal)
}

func TestState_SetUSDTPrice(t *testing.T) {
	var price *big.Int
	state := newDummyState()

	price = state.GetUSDTPrice()
	if price == nil || price.Sign() != 0 {
		t.Errorf("GetUSDTPrice() is failed: %s != 0", price)
	}

	// In case of valid prices
	prices := []*big.Int{new(big.Int), big.NewInt(1234)}
	for _, price = range prices {
		if err := state.SetUSDTPrice(price); err != nil {
			t.Errorf("SetUSDTPrice() is failed: %#v", price)
		}
		price2 := state.GetUSDTPrice()
		if price2 == nil || price.Cmp(price2) != 0 {
			t.Errorf("GetUSDTPrice() is failed: %s != %s", price, price2)
		}
	}

	// In case of invalid prices
	invalidPrices := []*big.Int{nil, big.NewInt(-1000)}
	for _, price = range invalidPrices {
		if err := state.SetUSDTPrice(price); err == nil {
			t.Errorf("Invalid price is accepted")
		}
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
		if err != nil {
			t.Errorf("AddPlanetManager is failed: %s", address)
		}
		if ok, err := state.IsPlanetManager(address); !ok || err != nil {
			t.Errorf("IsPlanetManager is failed: %s", address)
		}
	}

	// Call AddPlanetManager with an invalid argument
	for _, address = range []module.Address{nil, addresses[0], addresses[1]} {
		if err = state.AddPlanetManager(address); err == nil {
			t.Errorf("Duplicate address is accepted in AddPlanetManager(): %s", address)
		}
	}

	// Remove the first item with RemovePlanetManager()
	address = addresses[0]
	if err = state.RemovePlanetManager(address); err != nil {
		t.Errorf("RemovePlanetManager is failed: %s", address)
	}
	if ok, err := state.IsPlanetManager(address); ok || err != nil {
		t.Errorf("IsPlanetManager is failed: %s", address)
	}
	for i := 1; i < len(addresses); i++ {
		ok, err := state.IsPlanetManager(addresses[i])
		if !(ok && err == nil) {
			t.Errorf("IsPlanetManager is failed: %s", addresses[i])
		}
	}

	// Remove the last item with RemovePlanetManager()
	address = addresses[len(addresses)-1]
	if err = state.RemovePlanetManager(address); err != nil {
		t.Errorf("RemovePlanetManager is failed: %s", address)
	}
	if ok, err := state.IsPlanetManager(address); ok || err != nil {
		t.Errorf("IsPlanetManager is failed: %s", address)
	}
	for i := 1; i < len(addresses)-1; i++ {
		ok, err := state.IsPlanetManager(addresses[i])
		if !(ok && err == nil) {
			t.Errorf("IsPlanetManager is failed: %s", addresses[i])
		}
	}

	// Invalid cases of RemovePlanetManager()
	for _, address = range []module.Address{nil, addresses[0], common.MustNewAddressFromString("hx1234")} {
		if err = state.RemovePlanetManager(address); err == nil {
			t.Errorf("Invalid argument is accepted in RemovePlanetManager(): %s", address)
		}
	}

	ok, err := state.IsPlanetManager(nil)
	if ok || err == nil {
		t.Errorf("IsPlanetManager() accpets nil")
	}
}

func TestState_SetIssueStart(t *testing.T) {
	var startBH, curBH, height int64
	state := newDummyState()

	// Success case: startBH > 0 && startBH > curBH
	startBH, curBH = 2000, 1000
	if err := state.SetIssueStart(curBH, startBH); err != nil {
		t.Errorf("SetIssueStart() is failed: startBH=%d curBH=%d", startBH, curBH)
	}
	height = state.getVarDB(hvhmodule.VarIssueStart).Int64()
	if height != startBH {
		t.Errorf("SetIssueStart() is failed")
	}
	if state.GetIssueStart() != startBH {
		t.Errorf("SetIssueStart() is failed")
	}

	// Failure case: startBH <= 0 || startBH <= curBH
	curBH = 1000
	height = state.getVarDB(hvhmodule.VarIssueStart).Int64()
	for _, startBH = range []int64{-100, 0, 100, 500, curBH} {
		if err := state.SetIssueStart(curBH, startBH); err == nil {
			t.Errorf("Invalid argument is accepted in SetIssueStart(): %d", startBH)
		}
		startBH = state.getVarDB(hvhmodule.VarIssueStart).Int64()
		if startBH != height {
			t.Errorf("SetIssueStart() is failed")
		}
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
		if err := state.RegisterPlanet(int64(i), isPrivate, isCompany, owner, usdt, price, height); err != nil {
			t.Errorf(err.Error())
		}
		expectedCount++
		planetCount = state.getVarDB(hvhmodule.VarAllPlanet).Int64()
		if planetCount != expectedCount {
			t.Errorf(
				"planetCount is not increased: cur(%d) != expected(%d)",
				planetCount, expectedCount,
			)
		}
	}

	for i := 0; i < 2; i++ {
		if err := state.UnregisterPlanet(int64(i)); err != nil {
			t.Errorf(err.Error())
		}
		expectedCount--
		checkAllPlanet(t, state, expectedCount)
	}

	planetCount = state.getVarDB(hvhmodule.VarAllPlanet).Int64()
	if err := state.UnregisterPlanet(int64(100)); err == nil {
		t.Errorf("No error while unregistering a non-existent Planet")
	}
	checkAllPlanet(t, state, planetCount)
}

func TestState_UnregisterPlanet_InAbnormalCase(t *testing.T) {
	state := newDummyState()
	checkAllPlanet(t, state, int64(0))
	if err := state.UnregisterPlanet(int64(-1)); err == nil {
		t.Errorf("Invalid id is allowed by UnregisterPlanet()")
	}
	if err := state.UnregisterPlanet(int64(100)); err == nil {
		t.Errorf("No error while unregistering a non-existent Planet")
	}
	checkAllPlanet(t, state, int64(0))
}

func TestState_GetBigInt(t *testing.T) {
	key := "key"
	state := newDummyState()
	varDB := state.getVarDB(key)

	if varDB.BigInt() != nil {
		t.Errorf("value is not nil")
	}

	value := state.getBigInt(key)
	if value == nil {
		t.Errorf("GetBigInt() error")
	}

	defValue := hvhmodule.BigIntHooverBudget
	value = state.getBigIntOrDefault(key, defValue)
	if value == nil || value.Cmp(defValue) != 0 {
		t.Errorf("GetBigIntOrDefault() error")
	}

	newValue := new(big.Int).Add(defValue, big.NewInt(100))
	err := state.setBigInt(key, newValue)
	if err != nil {
		t.Errorf(err.Error())
	}

	value = state.getBigInt(key)
	if value == nil || value.Cmp(newValue) != 0 {
		t.Errorf("GetBigIntOrDefault() error")
	}

	value = state.getBigIntOrDefault(key, defValue)
	if value == nil || value.Cmp(newValue) != 0 {
		t.Errorf("GetBigIntOrDefault() error")
	}
}

func TestState_GetPlanet(t *testing.T) {
	id := int64(1)

	state := newDummyState()
	p, err := state.GetPlanet(id)
	if p != nil || err == nil {
		t.Errorf("GetPlanet() error")
	}

	dictDB := state.getDictDB(hvhmodule.DictPlanet, 1)

	p = newDummyPlanet(false, false, 1234)
	if err = state.setPlanet(dictDB, id, p); err != nil {
		t.Errorf(err.Error())
	}

	p2, err := state.getPlanet(dictDB, id)
	if p2 == nil || err != nil {
		t.Errorf("getPlanet() error")
	}
	if !p.equal(p2) {
		t.Errorf("getPlanet() error")
	}
}

func TestState_GetPlanetReward(t *testing.T) {
	id := int64(1)
	reward := big.NewInt(100)

	state := newDummyState()
	pr, err := state.GetPlanetReward(id)
	if pr == nil || err != nil {
		t.Errorf("GetPlanetReward() error: pr=%#v err=%s", pr, err)
	}
	if !(pr.Total().Sign() == 0 && pr.Current().Sign() == 0 && pr.LastTermNumber() == 0) {
		t.Errorf("Unexpected PlanetReward")
	}

	if err = pr.increment(10, reward, reward); err != nil {
		t.Errorf(err.Error())
	}
	if err = state.setPlanetReward(id, pr); err != nil {
		t.Errorf(err.Error())
	}

	pr2, err := state.GetPlanetReward(id)
	if pr2 == nil || err != nil {
		t.Errorf("GetPlanetReward() error: pr2=%#v err=%s", pr2, err)
	}
	if !pr.equal(pr2) {
		t.Errorf("setPlanetReward() error")
	}
}

func TestState_IncrementWorkingPlanet(t *testing.T) {
	state := newDummyState()

	ov := state.getInt64(hvhmodule.VarWorkingPlanet)
	err := state.IncrementWorkingPlanet()
	if err != nil {
		t.Errorf("IncrementWorkingPlanet() error")
	}

	nv := state.getInt64(hvhmodule.VarWorkingPlanet)
	if nv != ov+1 {
		t.Errorf("IncrementWorkingPlanet() error")
	}
}

func TestState_InitState(t *testing.T) {
	state := newDummyState()

	if state.GetTermPeriod() != hvhmodule.TermPeriod {
		t.Errorf("GetTermPeriod() error")
	}
	if state.GetHooverBudget().Cmp(hvhmodule.BigIntHooverBudget) != 0 {
		t.Errorf("GetHooverBudget() error")
	}
	if state.GetIssueReductionCycle() != hvhmodule.IssueReductionCycle {
		t.Errorf("GetIssueReductionCycle() error")
	}
	if state.GetIssueStart() != 0 {
		t.Errorf("GetIssueStart() error")
	}
	if state.GetIssueReductionRate().Cmp(hvhmodule.BigRatIssueReductionRate) != 0 {
		t.Errorf("GetIssueReductionRate() error")
	}

	cfg := &StateConfig{}
	cfg.TermPeriod = &common.HexInt64{Value: 100}
	cfg.HooverBudget = common.NewHexInt(1_000_000)
	cfg.IssueReductionCycle = &common.HexInt64{Value: 180}
	cfg.PrivateLockup = &common.HexInt64{Value: 100}
	cfg.PrivateReleaseCycle = &common.HexInt64{Value: 12}
	cfg.USDTPrice = common.NewHexInt(200_000)

	if err := state.InitState(cfg); err != nil {
		t.Errorf(err.Error())
	}

	if state.GetTermPeriod() != cfg.TermPeriod.Value {
		t.Errorf("GetTermPeriod() error")
	}
	if state.GetHooverBudget().Cmp(cfg.HooverBudget.Value()) != 0 {
		t.Errorf("GetHooverBudget() error")
	}
	if state.GetIssueReductionCycle() != cfg.IssueReductionCycle.Value {
		t.Errorf("GetIssueReductionCycle() error")
	}
	if state.getInt64(hvhmodule.VarPrivateLockup) != cfg.PrivateLockup.Value {
		t.Errorf("VarPrivateLockup error")
	}
	if state.getInt64(hvhmodule.VarPrivateReleaseCycle) != cfg.PrivateReleaseCycle.Value {
		t.Errorf("VarPrivateReleaseCycle error")
	}
	if state.GetIssueReductionRate().Cmp(hvhmodule.BigRatIssueReductionRate) != 0 {
		t.Errorf("GetIssueReductionRate() error")
	}
}

func TestState_DecreaseRewardRemain(t *testing.T) {
	const initRewardRemain = 1_000_000
	const key = hvhmodule.VarRewardRemain
	state := newDummyState()

	safeSetBigInt(t, state, key, big.NewInt(initRewardRemain))

	amount := int64(1000)
	if err := state.DecreaseRewardRemain(big.NewInt(amount)); err != nil {
		t.Errorf(err.Error())
	}

	newRewardRemain := state.getBigInt(key)
	if big.NewInt(initRewardRemain).Cmp(new(big.Int).Add(newRewardRemain, big.NewInt(amount))) != 0 {
		t.Errorf("DecreaseRewardRemain() error")
	}
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
	if err := pr.increment(1, amount, amount); err != nil {
		t.Errorf("planetReward.increment() error")
	}

	p2 := deepCopyPlanet(t, p)
	pr2 := deepCopyPlanetReward(t, pr)

	reward, err := state.calcClaimableReward(height, p, pr)
	if err != nil {
		t.Errorf(err.Error())
	}
	if reward.Cmp(amount) != 0 {
		t.Errorf("calcClaimableReward() error")
	}
	if !p.equal(p2) {
		t.Errorf("calcClaimableReward() MUST NOT change the Planet state")
	}
	if !pr.equal(pr2) {
		t.Errorf("calcClaimableReward() MUST NOT change the planetReward state")
	}
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
		if err := state.setPlanetReward(int64(i), pr); err != nil {
			t.Errorf(err.Error())
		}
	}

	planetDictDB := state.getDictDB(hvhmodule.DictPlanet, 1)
	for i := 1; i < 4; i++ {
		p := deepCopyPlanet(t, ps[i])
		if err := state.setPlanet(planetDictDB, int64(i), p); err != nil {
			t.Errorf(err.Error())
		}
	}

	if err := state.SetIssueStart(1, issueStart); err != nil {
		t.Errorf(err.Error())
	}

	height := state.GetIssueStart() + state.GetTermPeriod()
	for _, id := range []int64{publicId, companyId} {
		reward, err := state.ClaimPlanetReward(id, height, owner)
		if reward == nil && err != nil {
			t.Errorf("ClaimPlanetReward() error")
		}
		if reward.Cmp(prs[id].Current()) != 0 {
			t.Errorf("ClaimPlanetReward() error")
		}
		pr, err := state.GetPlanetReward(id)
		if err != nil {
			t.Errorf(err.Error())
		}
		if pr.Total().Cmp(prs[id].Total()) != 0 {
			t.Errorf("Incorrect total")
		}
		if pr.Current().Sign() != 0 {
			t.Errorf("Incorrect current")
		}
		if pr.LastTermNumber() != prs[id].LastTermNumber() {
			t.Errorf("Incorrect lastTN")
		}
	}

	id := int64(privateId)
	reward, err := state.ClaimPlanetReward(id, height, owner)
	if reward == nil && err != nil {
		t.Errorf("ClaimPlanetReward() error")
	}
	if reward.Sign() != 0 {
		t.Errorf("ClaimPlanetReward() error")
	}
}
