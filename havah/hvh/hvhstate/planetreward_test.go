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

	err := pr.claim(reward)
	if err != nil {
		t.Errorf(err.Error())
	}
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

func TestPlanetReward_Increment(t *testing.T) {
	tn := int64(10)
	reward := big.NewInt(1000)
	total := new(big.Int)

	pr := newEmpyPlanetReward()

	for i := 0; i < 3; i++ {
		total.Add(total, reward)
		tn++

		if err := pr.increment(tn, reward); err != nil {
			t.Errorf(err.Error())
		}

		if pr.LastTermNumber() != tn {
			t.Errorf("lastTermNumber error")
		}
		if pr.Current().Cmp(reward) != 0 {
			t.Errorf("current error")
		}
		if pr.Total().Cmp(total) != 0 {
			t.Errorf("total error")
		}
		if err := pr.claim(reward); err != nil {
			t.Errorf(err.Error())
		}
	}

	// Ignore multiple rewards during the same term
	pr2 := pr.clone()
	if err := pr.increment(tn, reward); err != nil {
		t.Errorf(err.Error())
	}
	if !pr.equal(pr2) {
		t.Errorf("increment() error")
	}
}

func TestPlanetReward_Claim(t *testing.T) {
	tn := int64(10)
	reward := big.NewInt(1000)
	amount := big.NewInt(700)

	pr := newEmpyPlanetReward()
	if err := pr.increment(tn, reward); err != nil {
		t.Errorf(err.Error())
	}

	if err := pr.claim(amount); err != nil {
		t.Errorf(err.Error())
	}

	pr2 := pr.clone()
	if err := pr.claim(amount); err == nil {
		t.Errorf("Over claimed")
	}
	if !pr.equal(pr2) {
		t.Errorf("claim() error")
	}
}
