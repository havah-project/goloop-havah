package hvhstate

import (
	"math/big"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/codec"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/scoreresult"
)

type PlanetFlag int

const (
	Company = PlanetFlag(1 << iota)
	Private
)

type planet struct {
	dirty bool

	flags  PlanetFlag
	owner  *common.Address
	usdt   *big.Int // priceInUSDT
	price  *big.Int
	height int64
}

func newPlanet(isPrivate, isCompany bool, owner module.Address, usdt, price *big.Int, height int64) *planet {
	var flags PlanetFlag
	if isPrivate {
		flags |= Private
	}
	if isCompany {
		flags |= Company
	}
	return &planet{
		flags:  flags,
		owner:  common.AddressToPtr(owner),
		usdt:   usdt,
		price:  price,
		height: height,
	}
}

func newPlanetFromBytes(b []byte) (*planet, error) {
	p := &planet{}
	if _, err := codec.BC.UnmarshalFromBytes(b, p); err != nil {
		return nil, scoreresult.UnknownFailureError.Wrap(err, "Failed to create a planet from bytes")
	}
	return p, nil
}

func (p *planet) IsPrivate() bool {
	return p.flags&Private != 0
}

func (p *planet) IsCompany() bool {
	return p.flags&Company != 0
}

func (p *planet) Owner() module.Address {
	return p.owner
}

func (p *planet) USDT() *big.Int {
	return new(big.Int).Set(p.usdt)
}

func (p *planet) Price() *big.Int {
	return new(big.Int).Set(p.price)
}

func (p *planet) Height() int64 {
	return p.height
}

func (p *planet) turnFlags(flags PlanetFlag, on bool) {
	if on {
		p.flags |= flags
	} else {
		p.flags &= ^flags
	}
}

func (p *planet) equal(p2 *planet) bool {
	return p.flags == p2.flags &&
		p.owner.Equal(p2.owner) &&
		p.usdt.Cmp(p2.usdt) == 0 &&
		p.price.Cmp(p2.price) == 0 &&
		p.height == p2.height
}

func (p *planet) setOwner(owner module.Address) error {
	if owner == nil {
		return scoreresult.New(hvhmodule.StatusIllegalArgument, "Invalid owner")
	}
	if !p.owner.Equal(owner) {
		p.owner = common.AddressToPtr(owner)
		if p.owner == nil {
			return scoreresult.New(hvhmodule.StatusIllegalArgument, "Invalid owner")
		}
		p.setDirty()
	}
	return nil
}

func (p *planet) isDirty() bool {
	return p.dirty
}

func (p *planet) setDirty() {
	p.dirty = true
}

func (p *planet) RLPDecodeSelf(d codec.Decoder) error {
	return d.DecodeListOf(&p.flags, &p.owner, &p.usdt, &p.price, &p.height)
}

func (p *planet) RLPEncodeSelf(e codec.Encoder) error {
	return e.EncodeListOf(p.flags, p.owner, p.usdt, p.price, p.height)
}

func (p *planet) Bytes() []byte {
	return codec.BC.MustMarshalToBytes(p)
}

func (p *planet) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"isPrivate":  p.IsPrivate(),
		"isCompany":  p.IsCompany(),
		"owner":      p.owner,
		"usdtPrice":  p.usdt,
		"havahPrice": p.price,
		"height":     p.height,
	}
}

// ====================================================================

/*
type PlanetSnapshot struct {
	planet
}

type PlanetState struct {
	snapshot *PlanetSnapshot
	planet
}

func NewPlanetState(flags PlanetFlag, owner module.Address, usdt, price *big.Int, height int64) *PlanetState {
	ps := PlanetState{}
	ps.snapshot = nil
	ps.flags = flags
	ps.owner = common.AddressToPtr(owner)
	ps.usdt = usdt
	ps.price = price
	ps.height = height
	return &ps
}

func NewPlanetStateFromSnapshot(pss *PlanetSnapshot) *PlanetState {
	return &PlanetState{
		snapshot: pss,
		planet:   pss.planet,
	}
}

func NewPlanetStateFromBytes(b []byte) *PlanetState {
	p, err := newPlanetFromBytes(b)
	if err != nil {
		return nil
	}
	return &PlanetState{nil, *p}
}

func (ps *PlanetState) IsDirty() bool {
	return ps.snapshot == nil
}

func (ps *PlanetState) setDirty() {
	ps.snapshot = nil
}

func (ps *PlanetState) SetOwner(address module.Address) {
	if !ps.owner.Equal(address) {
		ps.owner = common.AddressToPtr(address)
		ps.setDirty()
	}
}
*/
