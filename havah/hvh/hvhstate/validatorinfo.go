package hvhstate

import (
	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/codec"
	"github.com/icon-project/goloop/common/crypto"
	"github.com/icon-project/goloop/common/errors"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/module"
)

type ValidatorInfo struct {
	version   int
	owner     module.Address
	publicKey *crypto.PublicKey
	grade     Grade
	name      string
	url       string

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

func (vi *ValidatorInfo) SetName(name string) error {
	if len(name) > hvhmodule.MaxValidatorNameLen {
		return errors.IllegalArgumentError.Errorf("Too long name: %s", name)
	}
	vi.name = name
	return nil
}

func (vi *ValidatorInfo) Grade() Grade {
	return vi.grade
}

func (vi *ValidatorInfo) Url() string {
	return vi.url
}

func (vi *ValidatorInfo) SetUrl(url string) error {
	if len(url) > hvhmodule.MaxValidatorUrlLen {
		return errors.IllegalArgumentError.Errorf("Too long url: %s", url)
	}
	vi.url = url
	return nil
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

	err := d.DecodeListOf(
		&vi.version, &owner, &pubKey, &vi.grade, &vi.name, &vi.url)
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
		vi.publicKey.SerializeCompressed(),
		vi.grade,
		vi.name,
		vi.url)
}

func (vi *ValidatorInfo) Bytes() []byte {
	return codec.BC.MustMarshalToBytes(vi)
}

func (vi *ValidatorInfo) Equal(other *ValidatorInfo) bool {
	return vi.version == other.version &&
		vi.owner.Equal(other.owner) &&
		vi.publicKey.Equal(other.publicKey) &&
		vi.grade == other.grade &&
		vi.name == other.name &&
		vi.url == other.url
}

func (vi *ValidatorInfo) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"owner":         vi.owner,
		"nodePublicKey": vi.publicKey.SerializeCompressed(),
		"node":          vi.Address(),
		"grade":         vi.grade,
		"name":          vi.name,
		"url":           vi.url,
	}
}

func NewValidatorInfo(
	owner module.Address, pubKey []byte, grade Grade, name string) (*ValidatorInfo, error) {
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
