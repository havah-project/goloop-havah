package hvhstate

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/crypto"
	"github.com/icon-project/goloop/module"
)

func newDummyAddress(id int, contract bool) module.Address {
	bs := make([]byte, common.AddressBytes)
	if contract {
		bs[0] = 1
	}
	for i := 20; i >= 0 && id > 0; i-- {
		bs[i] = byte(id & 0xff)
		id >>= 8
	}
	return common.MustNewAddress(bs)
}

func newDummyValidatorInfo(id, grade int) *ValidatorInfo {
	owner := newDummyAddress(id, false)
	name := fmt.Sprintf("name-%02d", id)
	_, pubKey := crypto.GenerateKeyPair()
	vi, _ := NewValidatorInfo(owner, grade, name, pubKey.SerializeUncompressed())
	return vi
}

func newDummyValidatorInfos(size int, grade int) []*ValidatorInfo {
	vis := make([]*ValidatorInfo, size)
	for i := 1; i <= size; i++ {
		vis[i] = newDummyValidatorInfo(i, grade)
	}
	return vis
}

func TestNewValidatorInfo(t *testing.T) {
	owner := common.MustNewAddressFromString("hx1")
	_, pubKey := crypto.GenerateKeyPair()
	name := "name01"
	address := common.NewAccountAddressFromPublicKey(pubKey)

	vi0, err := NewValidatorInfo(owner, GradeMain, name, pubKey.SerializeCompressed())
	assert.NoError(t, err)
	assert.NotNil(t, vi0)

	assert.Zero(t, vi0.Version())
	assert.True(t, owner.Equal(vi0.Owner()))
	assert.Equal(t, GradeMain, vi0.Grade())
	assert.Equal(t, name, vi0.Name())
	assert.True(t, pubKey.Equal(vi0.PublicKey()))
	assert.True(t, address.Equal(vi0.Address()))
	assert.False(t, address.Equal(owner))

	vi1, err := NewValidatorInfo(owner, GradeMain, name, pubKey.SerializeUncompressed())
	assert.NoError(t, err)
	assert.True(t, vi0.Equal(vi1))

	vi2, err := NewValidatorInfoFromBytes(vi0.Bytes())
	assert.True(t, vi0.Equal(vi2))
}

func TestValidatorInfo_SetPublicKey(t *testing.T) {
	vi0 := newDummyValidatorInfo(1, GradeNormal)
	vi1 := newDummyValidatorInfo(1, GradeNormal)
	assert.False(t, vi0.Equal(vi1))
	assert.False(t, vi0.Address().Equal(vi1.Address()))
	assert.NotZero(t, bytes.Compare(vi0.Bytes(), vi1.Bytes()))

	err := vi1.SetPublicKey(vi0.PublicKey().SerializeCompressed())
	assert.NoError(t, err)
	assert.True(t, vi0.Equal(vi1))
	assert.True(t, vi0.Address().Equal(vi1.Address()))
	assert.Zero(t, bytes.Compare(vi0.Bytes(), vi1.Bytes()))
}
