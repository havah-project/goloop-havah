package hvhstate

import (
	"github.com/icon-project/goloop/common/codec"
)

const (
	FlagDisabled = 1 << iota
	FlagUnregistered
)

type ValidatorStatus struct {
	version  int
	flags    int
	nonVotes int
}

func (vs *ValidatorStatus) Version() int {
	return vs.version
}

func (vs *ValidatorStatus) NonVotes() int {
	return vs.nonVotes
}

func (vs *ValidatorStatus) IncrementNonVotes() int {
	vs.nonVotes++
	return vs.nonVotes
}

func (vs *ValidatorStatus) ResetNonVotes() {
	vs.nonVotes = 0
}

func (vs *ValidatorStatus) Enable() {
	vs.setFlags(FlagDisabled, false)
}

func (vs *ValidatorStatus) SetDisabled() {
	vs.setFlags(FlagDisabled, true)
}

func (vs *ValidatorStatus) Disabled() bool {
	return vs.all(FlagDisabled)
}

func (vs *ValidatorStatus) setFlags(flags int, on bool) {
	if on {
		vs.flags |= flags
	} else {
		vs.flags &= ^flags
	}
}

func (vs *ValidatorStatus) SetUnregistered() {
	vs.setFlags(FlagUnregistered, true)
}

func (vs *ValidatorStatus) Unregistered() bool {
	return vs.all(FlagUnregistered)
}

func (vs *ValidatorStatus) all(flags int) bool {
	return vs.flags&flags == flags
}

func (vs *ValidatorStatus) RLPDecodeSelf(d codec.Decoder) error {
	return d.DecodeListOf(&vs.version, &vs.flags, &vs.nonVotes)
}

func (vs *ValidatorStatus) RLPEncodeSelf(e codec.Encoder) error {
	return e.EncodeListOf(vs.version, vs.flags, vs.nonVotes)
}

func (vs *ValidatorStatus) Equal(other *ValidatorStatus) bool {
	return vs.version == other.version &&
		vs.flags == other.flags &&
		vs.nonVotes == other.nonVotes
}

func (vs *ValidatorStatus) Bytes() []byte {
	return codec.BC.MustMarshalToBytes(vs)
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
