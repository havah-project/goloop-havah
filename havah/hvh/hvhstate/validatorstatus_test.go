package hvhstate

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/icon-project/goloop/common/codec"
	"github.com/icon-project/goloop/havah/hvhmodule"
)

func TestNewValidatorStatus(t *testing.T) {
	vs0 := NewValidatorStatus()
	assert.Zero(t, vs0.Version())
	assert.Zero(t, vs0.NonVotes())
	assert.False(t, vs0.Disabled())
	assert.False(t, vs0.Disqualified())

	vs1, err := NewValidatorStatusFromBytes(vs0.Bytes())
	assert.NoError(t, err)
	assert.True(t, vs1.Equal(vs0))

	vs0.SetDisabled()
	vs0.SetDisqualified()
	assert.False(t, vs0.Equal(vs1))

	vs2, err := NewValidatorStatusFromBytes(vs0.Bytes())
	assert.NoError(t, err)
	assert.True(t, vs2.Equal(vs0))
}

func TestValidatorStatus_IncrementNonVotes(t *testing.T) {
	vs := NewValidatorStatus()
	assert.Zero(t, vs.NonVotes())
	for i := 0; i < 3; i++ {
		nonVotes := vs.IncrementNonVotes()
		assert.Equal(t, int64(i+1), nonVotes)
		assert.Equal(t, nonVotes, vs.NonVotes())
	}

	vs.ResetNonVotes()
	assert.Zero(t, vs.NonVotes())
}

func TestValidatorStatus_Disabled(t *testing.T) {
	vs := NewValidatorStatus()
	assert.False(t, vs.Disabled())

	vs.SetDisabled()
	assert.True(t, vs.Disabled())
	assert.False(t, vs.Disqualified())

	err := vs.Enable(false)
	assert.NoError(t, err)
	assert.False(t, vs.Disabled())
	assert.False(t, vs.Disqualified())
}

func TestValidatorStatus_Enable(t *testing.T) {
	var err error
	vs := NewValidatorStatus()
	assert.False(t, vs.Disabled())

	for i := 0; i < hvhmodule.MaxEnableCount; i++ {
		vs.SetDisabled()
		assert.True(t, vs.Disabled())

		err = vs.Enable(false)
		assert.NoError(t, err)
		assert.False(t, vs.Disabled())
	}
	assert.Zero(t, vs.EnableCount())

	// Success case: enableCount=0, Enabled, Call Enable()
	err = vs.Enable(false)
	assert.NoError(t, err)
	assert.Zero(t, vs.EnableCount())
	assert.True(t, vs.Enabled())

	vs.SetDisabled()
	assert.True(t, vs.Disabled())

	err = vs.Enable(false)
	assert.Error(t, err)
	assert.True(t, vs.Disabled())
	assert.Zero(t, vs.EnableCount())

	err = vs.Enable(true)
	assert.NoError(t, err)
	assert.False(t, vs.Disabled())
	assert.Equal(t, hvhmodule.MaxEnableCount, vs.EnableCount())

	vs.SetDisabled()
	vs.SetDisqualified()
	for _, calledByGov := range []bool{true, false} {
		err = vs.Enable(calledByGov)
		assert.Error(t, err)
		assert.True(t, vs.Disabled())
		assert.True(t, vs.Disqualified())
	}
}

func TestValidatorStatus_Disqualified(t *testing.T) {
	vs := NewValidatorStatus()
	assert.False(t, vs.Disqualified())
	assert.False(t, vs.Disabled())

	vs.SetDisqualified()
	assert.True(t, vs.Disqualified())
	assert.False(t, vs.Disabled())
}

func TestValidatorStatus_RLPDecodeSelf(t *testing.T) {
	var err error

	vs := NewValidatorStatus()
	vs.IncrementNonVotes()
	vs.SetDisabled()
	err = vs.Enable(false)
	assert.NoError(t, err)

	buf := bytes.NewBuffer(nil)
	e := codec.BC.NewEncoder(buf)

	err = vs.RLPEncodeSelf(e)
	assert.NoError(t, err)

	assert.Zero(t, bytes.Compare(vs.Bytes(), buf.Bytes()))

	d := codec.BC.NewDecoder(buf)
	vs2 := NewValidatorStatus()
	err = vs2.RLPDecodeSelf(d)
	assert.NoError(t, err)

	assert.True(t, vs2.Equal(vs))
	assert.True(t, vs2.Enabled())
	assert.Equal(t, int64(1), vs2.NonVotes())
	assert.Equal(t, hvhmodule.MaxEnableCount-1, vs2.EnableCount())
}
