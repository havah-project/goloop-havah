package hvhstate

import (
	"github.com/icon-project/goloop/common/codec"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/service/scoreresult"
)

const (
	FlagDisabled = 1 << iota
	FlagDisqualified
)

type ValidatorStatus struct {
	version     int
	flags       int
	nonVotes    int64
	enableCount int
}

func (vs *ValidatorStatus) Version() int {
	return vs.version
}

func (vs *ValidatorStatus) NonVotes() int64 {
	return vs.nonVotes
}

func (vs *ValidatorStatus) EnableCount() int {
	return vs.enableCount
}

func (vs *ValidatorStatus) IncrementNonVotes() int64 {
	vs.nonVotes++
	return vs.nonVotes
}

func (vs *ValidatorStatus) ResetNonVotes() {
	vs.nonVotes = 0
}

func (vs *ValidatorStatus) Enable(calledByGov bool) error {
	if calledByGov {
		return vs.enableByGov()
	} else {
		return vs.enableByOwner()
	}
}

func (vs *ValidatorStatus) enableByGov() error {
	err := vs.enable()
	vs.enableCount = 0
	return err
}

func (vs *ValidatorStatus) enableByOwner() error {
	if vs.enableCount >= hvhmodule.MaxEnableCount {
		return scoreresult.AccessDeniedError.Errorf(
			"MaxEnableCount exceeded: %d", vs.enableCount)
	}
	return vs.enable()
}

func (vs *ValidatorStatus) enable() error {
	if vs.Disabled() {
		vs.enableCount++
		vs.setFlags(FlagDisabled, false)
	}
	return nil
}

func (vs *ValidatorStatus) SetDisabled() {
	vs.setFlags(FlagDisabled, true)
}

func (vs *ValidatorStatus) Disabled() bool {
	return vs.all(FlagDisabled)
}

func (vs *ValidatorStatus) Enabled() bool {
	return vs.flags == 0
}

func (vs *ValidatorStatus) setFlags(flags int, on bool) {
	if on {
		vs.flags |= flags
	} else {
		vs.flags &= ^flags
	}
}

func (vs *ValidatorStatus) SetDisqualified() {
	vs.setFlags(FlagDisqualified, true)
}

func (vs *ValidatorStatus) Disqualified() bool {
	return vs.all(FlagDisqualified)
}

func (vs *ValidatorStatus) all(flags int) bool {
	return vs.flags&flags == flags
}

func (vs *ValidatorStatus) RLPDecodeSelf(d codec.Decoder) error {
	return d.DecodeListOf(&vs.version, &vs.flags, &vs.nonVotes, &vs.enableCount)
}

func (vs *ValidatorStatus) RLPEncodeSelf(e codec.Encoder) error {
	return e.EncodeListOf(vs.version, vs.flags, vs.nonVotes, vs.enableCount)
}

func (vs *ValidatorStatus) Equal(other *ValidatorStatus) bool {
	return vs.version == other.version &&
		vs.flags == other.flags &&
		vs.nonVotes == other.nonVotes &&
		vs.enableCount == other.enableCount
}

func (vs *ValidatorStatus) Bytes() []byte {
	return codec.BC.MustMarshalToBytes(vs)
}

func (vs *ValidatorStatus) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"flags":       vs.flags,
		"nonVotes":    vs.nonVotes,
		"enableCount": vs.enableCount,
	}
}

func NewValidatorStatus() *ValidatorStatus {
	return &ValidatorStatus{}
}

func NewValidatorStatusFromBytes(b []byte) (*ValidatorStatus, error) {
	vs := &ValidatorStatus{}
	if b != nil && len(b) > 0 {
		if _, err := codec.UnmarshalFromBytes(b, vs); err != nil {
			return nil, err
		}
	}
	return vs, nil
}
