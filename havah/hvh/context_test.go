package hvh

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/state"
)

func newMockCallContextAndFrom() (hvhmodule.CallContext, module.Address) {
	from := common.MustNewAddressFromString("hx1234")
	mcc := newMockCallContext()
	return NewCallContext(mcc, from), from
}

func TestCallContextImpl_Issue(t *testing.T) {
	cc, from := newMockCallContextAndFrom()
	amount := big.NewInt(1000)

	ots := cc.GetTotalSupply()
	assert.Zero(t, ots.Sign())

	balance := cc.GetBalance(from)
	assert.Zero(t, balance.Sign())

	nts, err := cc.Issue(from, amount)
	assert.NoError(t, err)
	assert.Zero(t, nts.Cmp(new(big.Int).Add(ots, amount)))
	assert.Zero(t, cc.GetBalance(from).Cmp(amount))

	ts := cc.GetTotalSupply()
	assert.Zero(t, ts.Cmp(new(big.Int).Add(ots, amount)))

	_, err = cc.Issue(from, new(big.Int).Neg(amount))
	assert.Error(t, err)
}

func TestCallContextImpl_Burn(t *testing.T) {
	cc, _ := newMockCallContextAndFrom()
	amount := big.NewInt(1000)

	ts, err := cc.Issue(state.SystemAddress, amount)
	assert.Zero(t, ts.Cmp(amount))

	nts, err := cc.Burn(amount)
	assert.NoError(t, err)
	assert.Zero(t, nts.Sign())

	balance := cc.GetBalance(state.SystemAddress)
	assert.Zero(t, balance.Sign())
}

func TestCallContextImpl_Transfer(t *testing.T) {
	to := common.MustNewAddressFromString("hx2222")
	cc, from := newMockCallContextAndFrom()
	amount := big.NewInt(10_000)

	_, err := cc.Issue(from, amount)
	assert.NoError(t, err)

	balance := cc.GetBalance(from)
	assert.Zero(t, amount.Cmp(balance))

	balance = cc.GetBalance(to)
	assert.Zero(t, balance.Sign())

	err = cc.Transfer(from, to, amount, module.Transfer)
	assert.NoError(t, err)

	balance = cc.GetBalance(from)
	assert.Zero(t, balance.Sign())

	balance = cc.GetBalance(to)
	assert.Zero(t, balance.Cmp(amount))
}
