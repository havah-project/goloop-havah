package hvhstate

import (
	"fmt"
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

type Planet struct {
	dirty bool

	flags  PlanetFlag
	owner  *common.Address
	usdt   *big.Int // priceInUSDT
	price  *big.Int // priceInHVH
	height int64    // regTime
}

func newPlanet(isPrivate, isCompany bool, owner module.Address, usdt, price *big.Int, height int64) *Planet {
	var flags PlanetFlag
	if isPrivate {
		flags |= Private
	}
	if isCompany {
		flags |= Company
	}
	return &Planet{
		flags:  flags,
		owner:  common.AddressToPtr(owner),
		usdt:   usdt,
		price:  price,
		height: height,
	}
}

func newPlanetFromBytes(b []byte) (*Planet, error) {
	p := &Planet{}
	if _, err := codec.BC.UnmarshalFromBytes(b, p); err != nil {
		return nil, scoreresult.UnknownFailureError.Wrap(err, "Failed to create a Planet from bytes")
	}
	return p, nil
}

func (p *Planet) IsPrivate() bool {
	return p.flags&Private != 0
}

func (p *Planet) IsCompany() bool {
	return p.flags&Company != 0
}

func (p *Planet) Owner() module.Address {
	return p.owner
}

func (p *Planet) USDT() *big.Int {
	return p.usdt
}

func (p *Planet) Price() *big.Int {
	return p.price
}

func (p *Planet) Height() int64 {
	return p.height
}

func (p *Planet) turnFlags(flags PlanetFlag, on bool) {
	if on {
		p.flags |= flags
	} else {
		p.flags &= ^flags
	}
}

func (p *Planet) equal(p2 *Planet) bool {
	return p.flags == p2.flags &&
		p.owner.Equal(p2.owner) &&
		p.usdt.Cmp(p2.usdt) == 0 &&
		p.price.Cmp(p2.price) == 0 &&
		p.height == p2.height
}

func (p *Planet) clone() *Planet {
	return &Planet{
		p.dirty,
		p.flags,
		p.owner,
		p.usdt,
		p.price,
		p.height,
	}
}

func (p *Planet) setOwner(owner module.Address) error {
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

func (p *Planet) isDirty() bool {
	return p.dirty
}

func (p *Planet) setDirty() {
	p.dirty = true
}

func (p *Planet) RLPDecodeSelf(d codec.Decoder) error {
	return d.DecodeListOf(&p.flags, &p.owner, &p.usdt, &p.price, &p.height)
}

func (p *Planet) RLPEncodeSelf(e codec.Encoder) error {
	return e.EncodeListOf(p.flags, p.owner, p.usdt, p.price, p.height)
}

func (p *Planet) Bytes() []byte {
	return codec.BC.MustMarshalToBytes(p)
}

func (p *Planet) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"isPrivate":  p.IsPrivate(),
		"isCompany":  p.IsCompany(),
		"owner":      p.owner,
		"usdtPrice":  p.usdt,
		"havahPrice": p.price,
		"height":     p.height,
	}
}

func (p *Planet) String() string {
	return fmt.Sprintf("Planet(flags=%d,owner=%s,usdt=%d,price=%d,height=%d)",
			p.flags, p.owner, p.usdt, p.price, p.height)
}
