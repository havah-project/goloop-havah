package hvh

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/service/scoredb"
	"github.com/icon-project/goloop/service/state"
)

func TestAccountStateProxy_Balance(t *testing.T) {
	addr := common.MustNewAddressFromString("hx1234")
	id := addr.ID()

	as := newMockAccount(id)
	as.SetBalance(big.NewInt(100))

	asp := newAccountStateProxy(id, as)
	assert.Equal(t, int64(100), asp.GetBalance().Int64())

	asp.SetBalance(big.NewInt(200))
	assert.Equal(t, int64(200), asp.GetBalance().Int64())
	assert.Equal(t, int64(100), as.GetBalance().Int64())
}

func TestAccountStateProxy_Value(t *testing.T) {
	id := state.SystemID
	as := newMockAccount(id)
	aps := newAccountStateProxy(id, as)

	db := scoredb.NewVarDB(aps, state.VarTotalSupply)

	assert.NoError(t, db.Set(100))
	assert.Equal(t, int64(100), db.Int64())
}
