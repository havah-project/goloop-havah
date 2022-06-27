package hvhstate

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/module"
)

func checkPlanetProperties(
	t *testing.T, p *Planet,
	isPrivate, isCompany bool, owner module.Address, usdt, price *big.Int,
	height int64,
) {
	assert.Equal(t, isPrivate, p.IsPrivate())
	assert.Equal(t, isCompany, p.IsCompany())
	assert.True(t, p.Owner().Equal(owner))
	assert.Zero(t, p.USDT().Cmp(usdt))
	assert.Zero(t, p.Price().Cmp(price))
	assert.Equal(t, height, p.Height())
}

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
	assert.NoError(t, err)
	assert.False(t, p2.isDirty())

	checkPlanetProperties(t, p, isPrivate, isCompany, owner, usdt, price, height)

	assert.True(t, p.equal(p2))
	assert.Zero(t, bytes.Compare(p.Bytes(), p2.Bytes()))
}

func TestPlanet_setOwner(t *testing.T) {
	isPrivate := true
	isCompany := true
	owner := common.MustNewAddressFromString("hx1234")
	usdt := big.NewInt(1000)
	price := new(big.Int).Mul(usdt, big.NewInt(10))
	height := int64(123)

	p := newPlanet(isPrivate, isCompany, owner, usdt, price, height)
	assert.False(t, p.isDirty())

	var newOwner module.Address
	assert.Error(t, p.setOwner(newOwner))

	newOwner = owner
	assert.NoError(t, p.setOwner(newOwner))
	assert.False(t, p.isDirty())

	newOwner = common.MustNewAddressFromString("hx5678")
	assert.NoError(t, p.setOwner(newOwner))
	assert.True(t, p.isDirty())
}

func TestPlanet_ToJSON(t *testing.T) {
	isPrivate := true
	isCompany := true
	owner := common.MustNewAddressFromString("hx1234")
	usdt := big.NewInt(1000)
	price := new(big.Int).Mul(usdt, big.NewInt(10))
	height := int64(123)

	p := newPlanet(isPrivate, isCompany, owner, usdt, price, height)
	jso := p.ToJSON()
	assert.Equal(t, 6, len(jso))
	assert.Equal(t, isPrivate, jso["isPrivate"].(bool))
	assert.Equal(t, isCompany, jso["isCompany"].(bool))
	assert.True(t, owner.Equal(jso["owner"].(*common.Address)))
	assert.Zero(t, usdt.Cmp(jso["usdtPrice"].(*big.Int)))
	assert.Zero(t, price.Cmp(jso["havahPrice"].(*big.Int)))
	assert.Equal(t, height, jso["height"].(int64))
}
