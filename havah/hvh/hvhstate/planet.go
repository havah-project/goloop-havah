package hvhstate

import (
	"math/big"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/codec"
	"github.com/icon-project/goloop/module"
)

type PlanetFlag int

const (
	None    = PlanetFlag(0)
	Company = PlanetFlag(1 << iota)
	Private
)

type planet struct {
	flags  PlanetFlag
	owner  *common.Address
	usdt   *big.Int // priceInUSDT
	price  *big.Int
	height int64
}

func newPlanet(flags PlanetFlag, owner *common.Address, usdt, price *big.Int, height int64) *planet {
	return &planet{
		flags,
		owner,
		usdt,
		price,
		height,
	}
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
	return p.usdt
}

func (p *planet) Price() *big.Int {
	return p.price
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

func (p *planet) RLPDecodeSelf(d codec.Decoder) error {
	return d.DecodeListOf(&p.flags, &p.owner, &p.usdt, &p.price, &p.height)
}

func (p *planet) RLPEncodeSelf(e codec.Encoder) error {
	return e.EncodeListOf(p.flags, p.owner, p.usdt, p.price, p.height)
}

func (p *planet) Bytes() []byte {
	var buf []byte
	e := codec.BC.NewEncoderBytes(&buf)
	if err := e.Encode(p); err != nil {
		panic("planet.Bytes() error")
	}
	return buf
}

// ====================================================================

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
