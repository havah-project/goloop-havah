package hvhstate

import (
	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/codec"
	"github.com/icon-project/goloop/common/crypto"
	"github.com/icon-project/goloop/module"
)

const (
	GradeNormal = iota
	GradeMain

	GradeReserved
)

func isGradeValid(grade int) bool {
	return grade >= 0 && grade < GradeReserved
}

type ValidatorInfo struct {
	version   int
	owner     module.Address
	grade     int
	name      string
	publicKey *crypto.PublicKey

	// Node address is derived from publicKey
	node module.Address
}

func (vi *ValidatorInfo) Version() int {
	return vi.version
}

func (vi *ValidatorInfo) SetPublicKey(pubKey []byte) error {
	if publicKey, err := crypto.ParsePublicKey(pubKey); err != nil {
		return err
	} else {
		vi.publicKey = publicKey
		vi.node = nil
	}
	return nil
}

func (vi *ValidatorInfo) Owner() module.Address {
	return vi.owner
}

func (vi *ValidatorInfo) Name() string {
	return vi.name
}

func (vi *ValidatorInfo) Grade() int {
	return vi.grade
}

func (vi *ValidatorInfo) Address() module.Address {
	if vi.node == nil {
		vi.node = common.NewAccountAddressFromPublicKey(vi.publicKey)
	}
	return vi.node
}

func (vi *ValidatorInfo) PublicKey() *crypto.PublicKey {
	return vi.publicKey
}

func (vi *ValidatorInfo) RLPDecodeSelf(d codec.Decoder) error {
	var owner *common.Address
	var pubKey []byte

	err := d.DecodeListOf(&vi.version, &owner, &vi.grade, &vi.name, &pubKey)
	if err != nil {
		return err
	}

	vi.owner = owner
	if err = vi.SetPublicKey(pubKey); err != nil {
		return err
	}
	return err
}

func (vi *ValidatorInfo) RLPEncodeSelf(e codec.Encoder) error {
	return e.EncodeListOf(
		vi.version,
		vi.owner.(*common.Address),
		vi.grade,
		vi.name,
		vi.publicKey.SerializeCompressed())
}

func (vi *ValidatorInfo) Bytes() []byte {
	return codec.BC.MustMarshalToBytes(vi)
}

func (vi *ValidatorInfo) Equal(other *ValidatorInfo) bool {
	return vi.version == other.version &&
		vi.owner.Equal(other.owner) &&
		vi.grade == other.grade &&
		vi.name == other.name &&
		vi.publicKey.Equal(other.publicKey)
}

func NewValidatorInfo(owner module.Address, grade int, name string, pubKey []byte) (*ValidatorInfo, error) {
	vi := &ValidatorInfo{
		owner: owner,
		grade: grade,
		name:  name,
	}
	if err := vi.SetPublicKey(pubKey); err != nil {
		return nil, err
	}
	return vi, nil
}

func NewValidatorInfoFromBytes(b []byte) (*ValidatorInfo, error) {
	vi := &ValidatorInfo{}
	if b != nil && len(b) > 0 {
		if _, err := codec.UnmarshalFromBytes(b, vi); err != nil {
			return nil, err
		}
	}
	return vi, nil
}
