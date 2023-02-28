package hvhstate

import (
	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/codec"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/scoreresult"
)

type AddressSet struct {
	addrs   []module.Address
	addrMap map[string]struct{}
}

func (as *AddressSet) Get(i int) module.Address {
	if i < 0 || i >= len(as.addrs) {
		return nil
	}
	return as.addrs[i]
}

func (as *AddressSet) Add(address module.Address) error {
	if as.addrs == nil {
		as.addrs = make([]module.Address, 0, 10)
	}

	key := ToKey(address)
	if _, ok := as.addrMap[key]; ok {
		return scoreresult.Errorf(
			hvhmodule.StatusDuplicate, "Address already exists: %s", address)
	}
	as.addrs = append(as.addrs, address)
	as.addrMap[key] = struct{}{}
	return nil
}

func (as *AddressSet) Len() int {
	return len(as.addrs)
}

func (as *AddressSet) Clear() {
	if as.Len() > 0 {
		as.addrs = nil
		as.addrMap = make(map[string]struct{})
	}
}

func (as *AddressSet) RLPEncodeSelf(e codec.Encoder) error {
	return e.Encode(as.addrs)
}

func (as *AddressSet) RLPDecodeSelf(d codec.Decoder) error {
	var addrs []*common.Address
	if err := d.Decode(&addrs); err != nil {
		return err
	}
	for _, addr := range addrs {
		if err := as.Add(addr); err != nil {
			return err
		}
	}
	return nil
}

func (as *AddressSet) Equal(other *AddressSet) bool {
	if as.Len() != other.Len() {
		return false
	}
	for i := 0; i < as.Len(); i++ {
		if !as.Get(i).Equal(other.Get(i)) {
			return false
		}
	}
	return true
}

func (as *AddressSet) Bytes() []byte {
	return codec.MustMarshalToBytes(as)
}

func NewAddressSet(capacity int) *AddressSet {
	return &AddressSet{
		addrs:   make([]module.Address, 0, capacity),
		addrMap: make(map[string]struct{}),
	}
}

func NewAddressSetFromBytes(b []byte) (*AddressSet, error) {
	addrSet := NewAddressSet(10)
	if b != nil && len(b) > 0 {
		if _, err := codec.BC.UnmarshalFromBytes(b, addrSet); err != nil {
			return nil, err
		}
	}
	return addrSet, nil
}
