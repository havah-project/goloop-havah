package hvhstate

import (
	"fmt"
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

func TestIsItTimeToCheckBlockVote(t *testing.T) {
	tests := []struct{
		blockIndex int64
		period int64
		expResult bool
	}{
		{0, 0, false},
		{50, 0, false},
		{0, 100, true},
		{1, 100, false},
		{100, 100, true},
		{150, 100, false},
		{200, 100, true},
	}
	for i, test := range tests {
		name := fmt.Sprintf("name-%d", i)
		t.Run(name, func(t *testing.T) {
			ok := IsItTimeToCheckBlockVote(test.blockIndex, test.period)
			assert.Equal(t, test.expResult, ok)
		})
	}
}
