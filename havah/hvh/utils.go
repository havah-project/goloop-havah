package hvh

import (
	"math/big"

	"github.com/icon-project/goloop/common/intconv"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/state"
)

const (
	SigRewardOffered = "RewardOffered(int,int,int,int)"
	SigRewardClaimed = "RewardClaimed(int,Address,int)"
)

func onRewardOffered(
	cc hvhmodule.CallContext,
	termSequence, id int64, reward, hoover *big.Int) {
	cc.OnEvent(
		state.SystemAddress,
		[][]byte{[]byte(SigRewardOffered)},
		[][]byte{
			intconv.Int64ToBytes(termSequence),
			intconv.Int64ToBytes(id),
			intconv.BigIntToBytes(reward),
			intconv.BigIntToBytes(hoover),
		},
	)
}

func onRewardClaimedEvent(
	cc hvhmodule.CallContext, id int64, owner module.Address, amount *big.Int) {
	cc.OnEvent(
		state.SystemAddress,
		[][]byte{[]byte(SigRewardClaimed)},
		[][]byte{
			intconv.Int64ToBytes(id),
			owner.Bytes(),
			intconv.BigIntToBytes(amount),
		},
	)
}
