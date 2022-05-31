package hvhstate

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/module"
)

func checkPlanetProperties(
	t *testing.T, p *planet,
	isPrivate, isCompany bool, owner module.Address, usdt, price *big.Int,
	height int64,
) {
	if p.IsPrivate() != isPrivate {
		t.Errorf("Planet.IsPrivate() error")
	}
	if p.IsCompany() != isCompany {
		t.Errorf("Planet.IsCompany() error")
	}
	if !p.Owner().Equal(owner) {
		t.Errorf("Planet.Owner() error")
	}
	if p.USDT().Cmp(usdt) != 0 {
		t.Errorf("Planet.USDT() error")
	}
	if p.Price().Cmp(price) != 0 {
		t.Errorf("Planet.Price() error")
	}
	if p.Height() != height {
		t.Errorf("Planet.Height() error")
	}
}

//func TestNewPlanetState(t *testing.T) {
//	owner := common.MustNewAddressFromString("hx1234")
//	usdt := big.NewInt(1000)
//	price := new(big.Int).Mul(usdt, big.NewInt(10))
//	height := int64(123)
//	ps := NewPlanetState(Private|Company, owner, usdt, price, height)
//	if ps == nil {
//		t.Errorf("Failed to create a PlanetState")
//	}
//	checkPlanetProperties(t, &ps.planet, true, true, owner, usdt, price, height)
//	if !ps.IsDirty() {
//		t.Errorf("NewPlanetState() error")
//	}
//}

//func TestNewPlanetStateFromSnapshot(t *testing.T) {
//	flags := Private
//	owner := common.MustNewAddressFromString("hx1234")
//	usdt := big.NewInt(1000)
//	price := new(big.Int).Mul(usdt, big.NewInt(10))
//	height := int64(2345)
//
//	pss := &PlanetSnapshot{planet{false, flags, owner, usdt, price, height}}
//	ps := NewPlanetStateFromSnapshot(pss)
//	if ps.IsDirty() {
//		t.Errorf("PlanetState.IsDirty() error")
//	}
//	checkPlanetProperties(t, &ps.planet, true, false, owner, usdt, price, height)
//
//	newOwner := common.MustNewAddressFromString("hx5678")
//	ps.SetOwner(newOwner)
//	if !ps.IsDirty() {
//		t.Errorf("PlanetState.IsDirty() error")
//	}
//	if !ps.Owner().Equal(newOwner) {
//		t.Errorf("PlanetState.SetOwner() error")
//	}
//}

func TestPlanet_Bytes(t *testing.T) {
	isPrivate := true
	isCompany := true
	owner := common.MustNewAddressFromString("hx1234")
	usdt := big.NewInt(1000)
	price := new(big.Int).Mul(usdt, big.NewInt(10))
	height := int64(123)

	p := newPlanet(isPrivate, isCompany, owner, usdt, price, height)
	checkPlanetProperties(t, p, isPrivate, isCompany, owner, usdt, price, height)

	p2, err := newPlanetFromBytes(p.Bytes())
	if err != nil {
		t.Errorf(err.Error())
	}
	if p2.isDirty() {
		t.Errorf("Incorrect initiala dirty state")
	}
	checkPlanetProperties(t, p, isPrivate, isCompany, owner, usdt, price, height)

	if !p.equal(p2) {
		t.Errorf("Failed to decode a planet")
	}
	if bytes.Compare(p.Bytes(), p2.Bytes()) != 0 {
		t.Errorf("plant.Bytes() error")
	}
}

func TestPlanet_setOwner(t *testing.T) {
	isPrivate := true
	isCompany := true
	owner := common.MustNewAddressFromString("hx1234")
	usdt := big.NewInt(1000)
	price := new(big.Int).Mul(usdt, big.NewInt(10))
	height := int64(123)

	p := newPlanet(isPrivate, isCompany, owner, usdt, price, height)
	if p.isDirty() {
		t.Errorf("Incorrect initial dirty state")
	}

	var newOwner module.Address
	if err := p.setOwner(newOwner); err == nil {
		t.Errorf("Nil owner is accepted by planet.setOwner()")
	}

	newOwner = owner
	if err := p.setOwner(newOwner); err != nil {
		t.Errorf("The same owner is not allowed in planet.setOwner()")
	}
	if p.isDirty() {
		t.Errorf("dirty is set to true even though the owner is not changed")
	}

	newOwner = common.MustNewAddressFromString("hx5678")
	if err := p.setOwner(newOwner); err != nil {
		t.Errorf("setOwner() failure")
	}
	if !p.isDirty() {
		t.Errorf("dirty should be set to true after owner is changed")
	}
}
