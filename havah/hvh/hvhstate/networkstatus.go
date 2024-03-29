package hvhstate

import (
	"fmt"

	"github.com/icon-project/goloop/common/codec"
	"github.com/icon-project/goloop/common/errors"
)

type NetMode int

const (
	NetModeInit NetMode = iota
	NetModeDecentralized
)

type NetworkStatus struct {
	version              int
	mode                 NetMode
	blockVoteCheckPeriod int64
	nonVoteAllowance     int64
	activeValidatorCount int64
}

func (ns *NetworkStatus) Version() int {
	return ns.version
}

func (ns *NetworkStatus) Mode() NetMode {
	return ns.mode
}

func (ns *NetworkStatus) SetMode(mode NetMode) {
	ns.mode = mode
}

func (ns *NetworkStatus) BlockVoteCheckPeriod() int64 {
	return ns.blockVoteCheckPeriod
}

func (ns *NetworkStatus) SetBlockVoteCheckPeriod(period int64) error {
	if period < 0 {
		return errors.IllegalArgumentError.Errorf(
			"Invalid blockVoteCheckPeriod: %d", period)
	}
	ns.blockVoteCheckPeriod = period
	return nil
}

func (ns *NetworkStatus) NonVoteAllowance() int64 {
	return ns.nonVoteAllowance
}

func (ns *NetworkStatus) SetNonVoteAllowance(allowance int64) error {
	if allowance < 0 {
		return errors.IllegalArgumentError.Errorf(
			"Invalid nonVoteAllowance: %d", allowance)
	}
	ns.nonVoteAllowance = allowance
	return nil
}

func (ns *NetworkStatus) ActiveValidatorCount() int64 {
	return ns.activeValidatorCount
}

func (ns *NetworkStatus) SetActiveValidatorCount(count int64) error {
	if count < 1 {
		return errors.IllegalArgumentError.Errorf("Invalid activeValidatorCount: %d", count)
	}
	ns.activeValidatorCount = count
	return nil
}

func (ns *NetworkStatus) Equal(other *NetworkStatus) bool {
	return ns.version == other.version &&
		ns.mode == other.mode &&
		ns.blockVoteCheckPeriod == other.blockVoteCheckPeriod &&
		ns.nonVoteAllowance == other.nonVoteAllowance &&
		ns.activeValidatorCount == other.activeValidatorCount
}

func (ns *NetworkStatus) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"mode":                 int(ns.mode),
		"blockVoteCheckPeriod": ns.blockVoteCheckPeriod,
		"nonVoteAllowance":     ns.nonVoteAllowance,
		"activeValidatorCount": ns.activeValidatorCount,
	}
}

func (ns *NetworkStatus) RLPDecodeSelf(d codec.Decoder) error {
	return d.DecodeListOf(
		&ns.version, &ns.mode, &ns.blockVoteCheckPeriod, &ns.nonVoteAllowance, &ns.activeValidatorCount)
}

func (ns *NetworkStatus) RLPEncodeSelf(e codec.Encoder) error {
	return e.EncodeListOf(
		ns.version, ns.mode, ns.blockVoteCheckPeriod, ns.nonVoteAllowance, ns.activeValidatorCount)
}

func (ns *NetworkStatus) Bytes() []byte {
	return codec.BC.MustMarshalToBytes(ns)
}

func (ns *NetworkStatus) IsDecentralized() bool {
	return ns.mode == NetModeDecentralized
}

func (ns *NetworkStatus) SetDecentralized() {
	ns.mode = NetModeDecentralized
}

func (ns *NetworkStatus) String() string {
	return fmt.Sprintf("NetworkStatus(" +
			"version=%d,mode=%d,blockVoteCheckPeriod=%d,nonVoteAllowance=%d,activeValidatorCount=%d)",
			ns.version, ns.mode, ns.blockVoteCheckPeriod, ns.nonVoteAllowance, ns.activeValidatorCount)
}

func NewNetworkStatus() *NetworkStatus {
	return &NetworkStatus{}
}

func NewNetworkStatusFromBytes(bs []byte) (*NetworkStatus, error) {
	ns := &NetworkStatus{}
	if _, err := codec.BC.UnmarshalFromBytes(bs, ns); err != nil {
		return nil, err
	}
	return ns, nil
}
