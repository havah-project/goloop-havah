package hvhstate

import (
	"math/big"
	"testing"

	"github.com/icon-project/goloop/common/db"
)

func newState() *State {
	mdb := db.NewMapDB()
	snapshot := NewSnapshot(mdb, nil)
	return NewStateFromSnapshot(snapshot, false, nil)
}

func TestState_SetUSDTPrice(t *testing.T) {
	var price *big.Int
	state := newState()

	price = state.GetUSDTPrice()
	if price == nil || price.Sign() != 0 {
		t.Errorf("GetUSDTPrice() is failed: %s != 0", price)
	}
	price = big.NewInt(1234)
	if err := state.SetUSDTPrice(price); err != nil {
		t.Errorf("SetUSDTPrice() is failed: %#v", price)
	}
	price2 := state.GetUSDTPrice()
	if price2 == nil || price.Cmp(price2) != 0 {
		t.Errorf("GetUSDTPrice() is failed: %s != %s", price, price2)
	}
	if price == price2 {
		t.Errorf("Reused price")
	}
}
