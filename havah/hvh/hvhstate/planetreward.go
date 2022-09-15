package hvhstate

import (
	"math/big"

	"github.com/icon-project/goloop/common/codec"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/service/scoreresult"
)

type planetReward struct {
	// Total received reward
	total *big.Int
	// The last term number when the reward is claimed
	lastTN int64
	// Reward to claim at this moment
	current *big.Int
}

func newEmpyPlanetReward() *planetReward {
	return &planetReward{new(big.Int), 0, new(big.Int)}
}

func newPlanetRewardFromBytes(b []byte) (*planetReward, error) {
	pr := newEmpyPlanetReward()
	if b != nil && len(b) > 0 {
		if _, err := codec.UnmarshalFromBytes(b, pr); err != nil {
			return nil, err
		}
	}
	return pr, nil
}

func (pr *planetReward) RLPDecodeSelf(d codec.Decoder) error {
	return d.DecodeListOf(&pr.total, &pr.lastTN, &pr.current)
}

func (pr *planetReward) RLPEncodeSelf(e codec.Encoder) error {
	return e.EncodeListOf(pr.total, pr.lastTN, pr.current)
}

func (pr *planetReward) Total() *big.Int {
	return pr.total
}

func (pr *planetReward) LastTermNumber() int64 {
	return pr.lastTN
}

func (pr *planetReward) Current() *big.Int {
	return pr.current
}

func (pr *planetReward) Bytes() []byte {
	return codec.MustMarshalToBytes(pr)
}

func (pr *planetReward) equal(other *planetReward) bool {
	return pr.total.Cmp(other.total) == 0 &&
		pr.current.Cmp(other.current) == 0 &&
		pr.lastTN == other.lastTN
}

func (pr *planetReward) clone() *planetReward {
	return &planetReward{
		pr.total,
		pr.lastTN,
		pr.current,
	}
}

func (pr *planetReward) increment(tn int64, amount *big.Int, total *big.Int) error {
	if amount == nil || amount.Sign() < 0 || total == nil || total.Sign() < 0 {
		return scoreresult.New(
			hvhmodule.StatusIllegalArgument, "Invalid amount")
	}
	if tn <= pr.lastTN {
		// tn should be larger than lastTN
		return scoreresult.Errorf(
			hvhmodule.StatusIllegalArgument, "Invalid termNumber: tn=%d lastTN=%d", tn, pr.lastTN)
	}
	// amount is different from total only if the planet is a company planet
	pr.total = new(big.Int).Add(pr.total, total)
	pr.current = new(big.Int).Add(pr.current, amount)
	pr.lastTN = tn
	return nil
}

func (pr *planetReward) claim(amount *big.Int) error {
	if amount == nil || amount.Sign() < 0 {
		return scoreresult.Errorf(
			hvhmodule.StatusIllegalArgument, "Invalid amount: %s", amount)
	}
	if pr.current.Cmp(amount) < 0 {
		return scoreresult.Errorf(
			hvhmodule.StatusIllegalArgument,
			"Not enough reward to claim: cur=%s amount=%s", pr.current, amount)
	}
	pr.current = new(big.Int).Sub(pr.current, amount)
	return nil
}
