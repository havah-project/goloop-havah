package hvhstate

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/codec"
	"github.com/icon-project/goloop/common/crypto"
	"github.com/icon-project/goloop/havah/hvhmodule"
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

func newDummyAddresses(startId int, contract bool, size int) []module.Address {
	addrs := make([]module.Address, size)
	for i := 0; i < size; i++ {
		id := startId + i
		addrs[i] = newDummyAddress(id, contract)
	}
	return addrs
}

func newDummyValidatorInfo(id int, grade Grade) *ValidatorInfo {
	owner := newDummyAddress(id, false)
	name := fmt.Sprintf("name-%02d", id)
	_, pubKey := crypto.GenerateKeyPair()
	vi, _ := NewValidatorInfo(owner, pubKey.SerializeUncompressed(), grade, name, nil)
	return vi
}

func TestNewValidatorInfo(t *testing.T) {
	owner := common.MustNewAddressFromString("hx1")
	_, pubKey := crypto.GenerateKeyPair()
	name := "name01"
	address := common.NewAccountAddressFromPublicKey(pubKey)
	url := fmt.Sprintf("https://www.%s.com/details.json", name)

	vi0, err := NewValidatorInfo(owner, pubKey.SerializeCompressed(), GradeMain, name, &url)
	assert.NoError(t, err)
	assert.NotNil(t, vi0)

	assert.Zero(t, vi0.Version())
	assert.True(t, owner.Equal(vi0.Owner()))
	assert.Equal(t, GradeMain, vi0.Grade())
	assert.Equal(t, name, vi0.Name())
	assert.True(t, pubKey.Equal(vi0.PublicKey()))
	assert.True(t, address.Equal(vi0.Address()))
	assert.False(t, address.Equal(owner))
	assert.Equal(t, url, vi0.Url())

	vi1, err := NewValidatorInfo(owner, pubKey.SerializeUncompressed(), GradeMain, name, &url)
	assert.NoError(t, err)
	assert.True(t, vi0.Equal(vi1))

	vi2, err := NewValidatorInfoFromBytes(vi0.Bytes())
	assert.True(t, vi0.Equal(vi2))
}

func TestValidatorInfo_SetPublicKey(t *testing.T) {
	vi0 := newDummyValidatorInfo(1, GradeSub)
	vi1 := newDummyValidatorInfo(1, GradeSub)
	assert.False(t, vi0.Equal(vi1))
	assert.False(t, vi0.Address().Equal(vi1.Address()))
	assert.NotZero(t, bytes.Compare(vi0.Bytes(), vi1.Bytes()))

	err := vi1.SetPublicKey(vi0.PublicKey().SerializeCompressed())
	assert.NoError(t, err)
	assert.True(t, vi0.Equal(vi1))
	assert.True(t, vi0.Address().Equal(vi1.Address()))
	assert.Zero(t, bytes.Compare(vi0.Bytes(), vi1.Bytes()))
}

func TestValidatorInfo_SetName(t *testing.T) {
	vi := newDummyValidatorInfo(1, GradeSub)
	assert.Equal(t, "name-01", vi.Name())

	err := vi.SetName("hello")
	assert.NoError(t, err)
	assert.Equal(t, "hello", vi.Name())

	var buf bytes.Buffer
	for i := 0; i < hvhmodule.MaxValidatorNameLen; i++ {
		buf.WriteString("a")
	}
	name := buf.String()
	err = vi.SetName(name)
	assert.NoError(t, err)
	assert.Equal(t, name, vi.Name())

	tooLongName := name + "a"
	err = vi.SetName(tooLongName)
	assert.Error(t, err)
	assert.Equal(t, name, vi.Name())
}

func TestValidatorInfo_SetUrl(t *testing.T) {
	vi := newDummyValidatorInfo(1, GradeSub)
	assert.Equal(t, "", vi.Url())

	url := "https://www.example.com/info"
	err := vi.SetUrl(&url)
	assert.NoError(t, err)
	assert.Equal(t, url, vi.Url())

	for ; len(url) < hvhmodule.MaxValidatorUrlLen; {
		url += "a"
	}
	err = vi.SetUrl(&url)
	assert.NoError(t, err)
	assert.Equal(t, url, vi.Url())

	tooLongUrl := url + "a"
	err = vi.SetName(tooLongUrl)
	assert.Error(t, err)
	assert.Equal(t, url, vi.Url())

	url = vi.Url()
	assert.NoError(t, vi.SetUrl(nil))
	assert.Equal(t, url, vi.Url())
}

func TestValidatorInfo_RLPDecodeSelf(t *testing.T) {
	var err error
	vi := newDummyValidatorInfo(1, GradeSub)

	buf := bytes.NewBuffer(nil)
	e := codec.BC.NewEncoder(buf)

	err = vi.RLPEncodeSelf(e)
	assert.NoError(t, err)

	assert.Zero(t, bytes.Compare(vi.Bytes(), buf.Bytes()))

	d := codec.BC.NewDecoder(buf)
	vi2 := &ValidatorInfo{}
	err = vi2.RLPDecodeSelf(d)
	assert.NoError(t, err)

	assert.True(t, vi2.Equal(vi))
}
