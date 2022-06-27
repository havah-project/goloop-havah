package hvhstate

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlanetReward_newEmptyPlanetReward(t *testing.T) {
	pr := newEmpyPlanetReward()
	assert.Zero(t, pr.Total().Sign())
	assert.Zero(t, pr.LastTermNumber())
	assert.Zero(t, pr.Current().Sign())
}

func TestPlanetReward_RLPEncodeSelf(t *testing.T) {
	tn := int64(1)
	reward := big.NewInt(1000)

	pr := newEmpyPlanetReward()
	assert.NoError(t, pr.increment(tn, reward, reward))
	assert.Zero(t, pr.Total().Cmp(reward))
	assert.Zero(t, pr.Current().Cmp(reward))
	assert.Equal(t, tn, pr.LastTermNumber())

	assert.NoError(t, pr.claim(reward))
	assert.Zero(t, pr.Total().Cmp(reward))
	assert.Zero(t, pr.Current().Sign())
	assert.Equal(t, tn, pr.LastTermNumber())

	b := pr.Bytes()
	pr2, err := newPlanetRewardFromBytes(b)
	assert.NoError(t, err)
	assert.True(t, pr.equal(pr2))
}

func TestPlanetReward_Increment(t *testing.T) {
	tn := int64(10)
	reward := big.NewInt(1000)
	total := new(big.Int)

	pr := newEmpyPlanetReward()

	for i := 0; i < 3; i++ {
		total.Add(total, reward)
		tn++

		assert.NoError(t, pr.increment(tn, reward, reward))
		assert.Equal(t, tn, pr.LastTermNumber())
		assert.Zero(t, pr.Current().Cmp(reward))
		assert.Zero(t, pr.Total().Cmp(total))
		assert.NoError(t, pr.claim(reward))
	}

	// Check for multiple rewards during the same term
	pr2 := pr.clone()
	assert.Error(t, pr.increment(tn, reward, reward))
	assert.True(t, pr.equal(pr2))
}

func TestPlanetReward_Claim(t *testing.T) {
	tn := int64(10)
	reward := big.NewInt(1000)
	amount := big.NewInt(700)

	pr := newEmpyPlanetReward()
	assert.NoError(t, pr.increment(tn, reward, reward))

	assert.NoError(t, pr.claim(amount))

	pr2 := pr.clone()
	assert.Error(t, pr.claim(amount))
	assert.True(t, pr.equal(pr2))
}
