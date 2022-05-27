package hvhstate

import (
	"math/big"
	"testing"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/db"
	"github.com/icon-project/goloop/module"
)

func newDummyState() *State {
	mdb := db.NewMapDB()
	snapshot := NewSnapshot(mdb, nil)
	return NewStateFromSnapshot(snapshot, false, nil)
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
