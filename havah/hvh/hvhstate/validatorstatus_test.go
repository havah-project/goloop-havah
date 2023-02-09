package hvhstate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewValidatorStatus(t *testing.T) {
	vs0 := NewValidatorStatus()
	assert.Zero(t, vs0.Version())
	assert.Zero(t, vs0.NonVotes())
	assert.False(t, vs0.Disabled())
	assert.False(t, vs0.Unregistered())

	vs1, err := NewValidatorStatusFromBytes(vs0.Bytes())
	assert.NoError(t, err)
	assert.True(t, vs1.Equal(vs0))

	vs0.SetDisabled()
	vs0.SetUnregistered()
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
		assert.Equal(t, i+1, nonVotes)
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
	assert.False(t, vs.Unregistered())

	vs.Enable()
	assert.False(t, vs.Disabled())
	assert.False(t, vs.Unregistered())
}

func TestValidatorStatus_Unregistered(t *testing.T) {
	vs := NewValidatorStatus()
	assert.False(t, vs.Unregistered())
	assert.False(t, vs.Disabled())

	vs.SetUnregistered()
	assert.True(t, vs.Unregistered())
	assert.False(t, vs.Disabled())
}
