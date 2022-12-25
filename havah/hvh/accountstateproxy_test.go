package hvh

import (
	"fmt"
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
	asp := newAccountStateProxy(id, as)

	db := scoredb.NewVarDB(asp, state.VarTotalSupply)

	assert.NoError(t, db.Set(100))
	assert.Equal(t, int64(100), db.Int64())
}

func TestAccountStateProxy_DeleteValue(t *testing.T) {
	id := state.SystemID
	oldVs := [][]byte{nil, []byte("old")}

	for i, oldV := range oldVs {
		name := fmt.Sprintf("name-%d", i)
		t.Run(name, func(t *testing.T) {
			as := newMockAccount(id)
			asp := newAccountStateProxy(id, as)

			k := []byte("key")
			v, err := as.SetValue(k, oldV)
			assert.NoError(t, err)
			assert.Nil(t, v)

			v, err = as.GetValue(k)
			assert.NoError(t, err)
			assert.Equal(t, oldV, v)

			v, err = asp.GetValue(k)
			assert.NoError(t, err)
			assert.Equal(t, oldV, v)

			v, err = asp.DeleteValue(k)
			assert.NoError(t, err)
			assert.Equal(t, oldV, v)

			v, err = asp.GetValue(k)
			assert.NoError(t, err)
			assert.Nil(t, v)

			v, err = as.GetValue(k)
			assert.NoError(t, err)
			assert.Equal(t, oldV, v)
		})
	}
}

func TestAccountStateProxy_SetValue(t *testing.T) {
	id := state.SystemID
	oldVs := [][]byte{nil, []byte("old")}
	newV := []byte("new")

	for i, oldV := range oldVs {
		name := fmt.Sprintf("name-%d", i)
		t.Run(name, func(t *testing.T) {
			as := newMockAccount(id)
			asp := newAccountStateProxy(id, as)

			k := []byte("key")
			v, err := as.SetValue(k, oldV)
			assert.NoError(t, err)
			assert.Nil(t, v)

			v, err = as.GetValue(k)
			assert.NoError(t, err)
			assert.Equal(t, oldV, v)

			v, err = asp.GetValue(k)
			assert.NoError(t, err)
			assert.Equal(t, oldV, v)

			v, err = asp.SetValue(k, newV)
			assert.NoError(t, err)
			assert.Equal(t, oldV, v)

			v, err = asp.GetValue(k)
			assert.NoError(t, err)
			assert.Equal(t, newV, v)

			v, err = as.GetValue(k)
			assert.NoError(t, err)
			assert.Equal(t, oldV, v)
		})
	}
}
