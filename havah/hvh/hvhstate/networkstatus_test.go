package hvhstate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNetMode(t *testing.T) {
	var mode NetMode
	assert.Equal(t, NetModeInit, mode)
	mode++
	assert.Equal(t, NetModeDecentralized, mode)
	mode++
	assert.True(t, mode == 2)
}

func TestNewNetworkStatus(t *testing.T) {
	ns := NewNetworkStatus()
	assert.NotNil(t, ns)
	assert.Zero(t, ns.Version())
	assert.Zero(t, ns.Mode())
	assert.Zero(t, ns.BlockVoteCheckPeriod())
	assert.Zero(t, ns.NonVoteAllowance())

	ns2, err := NewNetworkStatusFromBytes(ns.Bytes())
	assert.NoError(t, err)
	assert.True(t, ns2.Equal(ns))
}

func TestNetworkStatus_SetMode(t *testing.T) {
	ns := NewNetworkStatus()
	assert.Equal(t, NetModeInit, ns.Mode())
	ns.SetMode(NetModeDecentralized)
	assert.Equal(t, NetModeDecentralized, ns.Mode())

	ns2, err := NewNetworkStatusFromBytes(ns.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, NetModeDecentralized, ns2.Mode())
}

func TestNetworkStatus_SetBlockVoteCheckPeriod(t *testing.T) {
	period := int64(100)
	ns := NewNetworkStatus()
	assert.Equal(t, int64(0), ns.BlockVoteCheckPeriod())
	ns.SetBlockVoteCheckPeriod(period)
	assert.Equal(t, period, ns.BlockVoteCheckPeriod())

	assert.Error(t, ns.SetBlockVoteCheckPeriod(int64(-10)))

	ns2, err := NewNetworkStatusFromBytes(ns.Bytes())
	assert.NoError(t, err)
	assert.True(t, ns2.Equal(ns))
	assert.Equal(t, period, ns2.BlockVoteCheckPeriod())
}

func TestNetworkStatus_SetNonVoteAllowance(t *testing.T) {
	allowance := int64(10)
	ns := NewNetworkStatus()
	assert.Equal(t, int64(0), ns.BlockVoteCheckPeriod())
	ns.SetNonVoteAllowance(allowance)
	assert.Equal(t, allowance, ns.NonVoteAllowance())

	assert.Error(t, ns.SetNonVoteAllowance(int64(-10)))

	ns2, err := NewNetworkStatusFromBytes(ns.Bytes())
	assert.NoError(t, err)
	assert.True(t, ns2.Equal(ns))
	assert.Equal(t, allowance, ns2.NonVoteAllowance())
}
