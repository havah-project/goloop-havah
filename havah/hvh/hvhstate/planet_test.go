package hvhstate

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/codec"
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

func TestNewPlanetState(t *testing.T) {
	owner := common.MustNewAddressFromString("hx1234")
	usdt := big.NewInt(1000)
	price := new(big.Int).Mul(usdt, big.NewInt(10))
	height := int64(123)
	ps := NewPlanetState(Private|Company, owner, usdt, price, height)
	if ps == nil {
		t.Errorf("Failed to create a PlanetState")
	}
	checkPlanetProperties(t, &ps.planet, true, true, owner, usdt, price, height)
	if !ps.IsDirty() {
		t.Errorf("NewPlanetState() error")
	}
}

func TestNewPlanetStateFromSnapshot(t *testing.T) {
	flags := Private
	owner := common.MustNewAddressFromString("hx1234")
	usdt := big.NewInt(1000)
	price := new(big.Int).Mul(usdt, big.NewInt(10))
	height := int64(2345)

	pss := &PlanetSnapshot{planet{flags, owner, usdt, price, height}}
	ps := NewPlanetStateFromSnapshot(pss)
	if ps.IsDirty() {
		t.Errorf("PlanetState.IsDirty() error")
	}
	checkPlanetProperties(t, &ps.planet, true, false, owner, usdt, price, height)

	newOwner := common.MustNewAddressFromString("hx5678")
	ps.SetOwner(newOwner)
	if !ps.IsDirty() {
		t.Errorf("PlanetState.IsDirty() error")
	}
	if !ps.Owner().Equal(newOwner) {
		t.Errorf("PlanetState.SetOwner() error")
	}
}

func TestPlanet_EncodeAndDecode(t *testing.T) {
	flags := Company | Private
	owner := common.MustNewAddressFromString("hx1234")
	usdt := big.NewInt(1000)
	price := new(big.Int).Mul(usdt, big.NewInt(10))
	height := int64(123)
	p := newPlanet(flags, owner, usdt, price, height)

	var buf []byte
	e := codec.BC.NewEncoderBytes(&buf)
	if err := e.Encode(p); err != nil {
		t.Errorf("Failed to encode a planet")
	}

	p2 := &planet{}
	if p.equal(p2) {
		t.Errorf("planet.equal() error")
	}
	d := codec.BC.NewDecoder(bytes.NewReader(buf))
	if err := d.Decode(p2); err != nil {
		t.Errorf("Failed to decode a planet")
	}
	if !p.equal(p2) {
		t.Errorf("Failed to decode a planet")
	}
}
