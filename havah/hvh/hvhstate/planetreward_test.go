package hvhstate

import (
	"math/big"
	"testing"
)

func TestPlanetReward_newEmptyPlanetReward(t *testing.T) {
	pr := newEmpyPlanetReward()
	if pr.total == nil || pr.total.Sign() != 0 {
		t.Errorf("Incorrect initial total value")
	}
	if pr.lastTN != 0 {
		t.Errorf("Incorrect initial lastTN")
	}
	if pr.current == nil || pr.current.Sign() != 0 {
		t.Errorf("Incorrect initial current")
	}
}

func TestPlanetReward_RLPEncodeSelf(t *testing.T) {
	tn := int64(1)
	reward := big.NewInt(1000)

	pr := newEmpyPlanetReward()
	if err := pr.increment(tn, reward); err != nil {
		t.Errorf(err.Error())
	}

	pr.claim()
	if pr.Total().Cmp(reward) != 0 {
		t.Errorf("Incorrect total")
	}
	if pr.Current().Sign() != 0 {
		t.Errorf("Incorrect current")
	}
	if pr.lastTN != tn {
		t.Errorf(
			"Incorrect termNumber: lastTN(%d) != tn(%d",
			pr.LastTermNumber(), tn)
	}

	b := pr.Bytes()
	pr2, err := newPlanetRewardFromBytes(b)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !pr.equal(pr2) {
		t.Errorf("planetReward.Bytes() or newPlanetRewardFromBytes() error")
	}
}
