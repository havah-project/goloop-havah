package hvhstate

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/icon-project/goloop/common/codec"
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

func TestNetworkStatus_IsDecentralized(t *testing.T) {
	ns := NewNetworkStatus()
	assert.False(t, ns.IsDecentralized())
	ns.SetDecentralized()
	assert.True(t, ns.IsDecentralized())
}

func TestNetworkStatus_RLPDecodeSelf(t *testing.T) {
	var err error
	allowance := int64(10)
	period := int64(100)

	ns := NewNetworkStatus()
	ns.SetDecentralized()
	assert.NoError(t, ns.SetNonVoteAllowance(allowance))
	assert.NoError(t, ns.SetBlockVoteCheckPeriod(period))

	buf := bytes.NewBuffer(nil)
	e := codec.BC.NewEncoder(buf)

	err = ns.RLPEncodeSelf(e)
	assert.NoError(t, err)

	assert.Zero(t, bytes.Compare(ns.Bytes(), buf.Bytes()))

	d := codec.BC.NewDecoder(buf)
	ns2 := NewNetworkStatus()
	err = ns2.RLPDecodeSelf(d)
	assert.NoError(t, err)

	assert.True(t, ns2.Equal(ns))
	assert.True(t, ns2.IsDecentralized())
	assert.Equal(t, allowance, ns2.NonVoteAllowance())
	assert.Equal(t, period, ns2.BlockVoteCheckPeriod())
}
