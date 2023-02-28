package hvhstate

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/icon-project/goloop/common/codec"
)

func TestAddressSet_Add(t *testing.T) {
	size := 5
	addrSet := NewAddressSet(0)

	for i := 1; i <= size; i++ {
		addr := newDummyAddress(i, false)
		err := addrSet.Add(addr)
		assert.NoError(t, err)
	}
	assert.Equal(t, size, addrSet.Len())

	for i := 1; i <= size; i++ {
		addr := newDummyAddress(i, false)
		err := addrSet.Add(addr)
		assert.Error(t, err)
		assert.Equal(t, size, addrSet.Len())
		assert.True(t, addrSet.Get(i-1).Equal(addr))
	}
}

func TestAddressSet_RLPDecodeSelf(t *testing.T) {
	size := 5
	addrSet := NewAddressSet(0)

	for i := 1; i <= size; i++ {
		addr := newDummyAddress(i, false)
		err := addrSet.Add(addr)
		assert.NoError(t, err)
	}

	buf := bytes.NewBuffer(nil)
	e := codec.BC.NewEncoder(buf)
	err := addrSet.RLPEncodeSelf(e)
	assert.NoError(t, err)

	e.Close()
	assert.True(t, buf.Len() > 0)

	d := codec.BC.NewDecoder(buf)
	addrSet2 := NewAddressSet(0)
	assert.Zero(t, addrSet2.Len())

	err = addrSet2.RLPDecodeSelf(d)
	assert.NoError(t, err)
	assert.Equal(t, size, addrSet2.Len())

	assert.True(t, addrSet.Equal(addrSet2))
}

func TestAddressSet_ToBytes(t *testing.T) {
	size := 3
	addrSet := NewAddressSet(0)

	for i := 1; i <= size; i++ {
		addr := newDummyAddress(i, false)
		err := addrSet.Add(addr)
		assert.NoError(t, err)
	}
	assert.Equal(t, size, addrSet.Len())

	buf := bytes.NewBuffer(nil)
	e := codec.BC.NewEncoder(buf)
	err := addrSet.RLPEncodeSelf(e)
	assert.NoError(t, err)

	bs := addrSet.Bytes()
	assert.Zero(t, 0, bytes.Compare(bs, buf.Bytes()))
}

func TestAddressSet_Clear(t *testing.T) {
	size := 3
	addrSet := NewAddressSet(0)

	for i := 1; i <= size; i++ {
		addr := newDummyAddress(i, false)
		err := addrSet.Add(addr)
		assert.NoError(t, err)
	}
	assert.Equal(t, size, addrSet.Len())

	addrSet.Clear()
	assert.Zero(t, addrSet.Len())

	addr := addrSet.Get(0)
	assert.Nil(t, addr)
}
