package hvhutils

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/crypto"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/module"
)

func TestIsNil(t *testing.T) {
	var a *big.Int
	var b interface{}
	b = a

	assert.True(t, a == nil)
	assert.False(t, b == nil)
	assert.True(t, IsNil(a))
	assert.True(t, IsNil(b))
}

func TestCheckNameLength(t *testing.T) {
	name := make([]byte, hvhmodule.MaxValidatorNameLen)
	err := CheckNameLength(string(name))
	assert.NoError(t, err)

	name = make([]byte, hvhmodule.MaxValidatorNameLen+1)
	err = CheckNameLength(string(name))
	assert.Error(t, err)
}

func TestCheckUrlLength(t *testing.T) {
	name := make([]byte, hvhmodule.MaxValidatorUrlLen)
	err := CheckUrlLength(string(name))
	assert.NoError(t, err)

	name = make([]byte, hvhmodule.MaxValidatorUrlLen+1)
	err = CheckUrlLength(string(name))
	assert.Error(t, err)
}

func TestCheckAddressArgument(t *testing.T) {
	eoa := common.MustNewAddressFromString("hx1")
	ca := common.MustNewAddressFromString("cx1")

	args := []struct{
		addr module.Address
		eoaOnly bool
		success bool
	}{
		{eoa, true, true},
		{eoa, false, true},
		{ca, true, false},
		{ca, false, true},
	}

	for i, arg := range args {
		name := fmt.Sprintf("name%d", i)
		t.Run(name, func(t *testing.T){
			err := CheckAddressArgument(arg.addr, arg.eoaOnly)
			if arg.success {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestCheckCompressedPublicKeyFormat(t *testing.T) {
	var err error
	_, publicKey := crypto.GenerateKeyPair()

	args := []struct{
		pubKey []byte
		success bool
	}{
		{nil, false},
		{ []byte{}, false},
		{publicKey.SerializeUncompressed(), false},
		{ publicKey.SerializeCompressed(), true},
	}

	for i, arg := range args {
		name := fmt.Sprintf("name-%02d", i)
		t.Run(name, func(t *testing.T){
			err = CheckCompressedPublicKeyFormat(arg.pubKey)
			if arg.success {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
