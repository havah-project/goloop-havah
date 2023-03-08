package hvhstate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToKey(t *testing.T) {
	addr := newDummyAddress(1, false)
	key := ToKey(addr)
	assert.Equal(t, string(addr.Bytes()), key)

	key = ToKey("hello")
	assert.Equal(t, "hello", key)

	o := []byte{0, 1, 2, 3}
	key = ToKey(o)
	assert.Equal(t, string(o), key)
}
