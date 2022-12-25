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

func newMockCallContextAndFrom(isQuery bool) (hvhmodule.CallContext, module.Address) {
	from := common.MustNewAddressFromString("hx1234")
	mcc := newMockCallContext()
	if isQuery {
		mcc.SetTransactionID(nil)
	}
	return NewCallContext(mcc, from), from
}

func TestCallContext_Issue(t *testing.T) {
	cc, from := newMockCallContextAndFrom(false)
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

func TestCallContext_Burn(t *testing.T) {
	cc, _ := newMockCallContextAndFrom(false)
	amount := big.NewInt(1000)

	ts, err := cc.Issue(state.SystemAddress, amount)
	assert.Zero(t, ts.Cmp(amount))
	assert.NoError(t, err)

	nts, err := cc.Burn(amount)
	assert.NoError(t, err)
	assert.Zero(t, nts.Sign())

	balance := cc.GetBalance(state.SystemAddress)
	assert.Zero(t, balance.Sign())
}

func TestCallContext_Transfer(t *testing.T) {
	to := common.MustNewAddressFromString("hx2222")
	cc, from := newMockCallContextAndFrom(false)
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

func TestCallContext_ActionOnQueryMode(t *testing.T) {
	var err error
	var balance *big.Int
	to := common.MustNewAddressFromString("hx2345")
	cc, from := newMockCallContextAndFrom(true)

	balance = cc.GetBalance(from)
	assert.Zero(t, balance.Sign())

	balance = cc.GetBalance(to)
	assert.Zero(t, balance.Sign())

	totalSupply := cc.GetTotalSupply()
	assert.Zero(t, totalSupply.Sign())

	// Issue
	amount := big.NewInt(10_000)
	totalSupply, err = cc.Issue(from, amount)
	assert.NoError(t, err)
	assert.Zero(t, totalSupply.Cmp(amount))

	fromBalance := cc.GetBalance(from)
	assert.Zero(t, amount.Cmp(fromBalance))

	// Transfer
	amount = big.NewInt(1_000)
	err = cc.Transfer(from, to, amount, module.Transfer)
	assert.NoError(t, err)
	toBalance := cc.GetBalance(to)
	assert.Equal(t, amount.Int64(), toBalance.Int64())
	fromBalance = cc.GetBalance(from)
	assert.Equal(t, int64(9_000), fromBalance.Int64())

	// Burn
	amountToBurn := big.NewInt(5_000)
	err = cc.Transfer(from, state.SystemAddress, amountToBurn, module.Transfer)
	assert.NoError(t, err)

	totalSupply, err = cc.Burn(amountToBurn)
	assert.NoError(t, err)
	assert.Equal(t, int64(5_000), totalSupply.Int64())
	assert.Equal(t, int64(4_000), cc.GetBalance(from).Int64())
	assert.Equal(t, int64(1_000), cc.GetBalance(to).Int64())
}

func TestCallContext_IsBaseTxInvoke(t *testing.T) {
	mcc, _ := newMockCallContextAndFrom(true)
	assert.False(t, mcc.IsBaseTxInvoked())

	mcc.SetBaseTxInvoked()
	assert.True(t, mcc.IsBaseTxInvoked())

	mcc, _ = newMockCallContextAndFrom(false)
	assert.True(t, mcc.IsBaseTxInvoked())
}
