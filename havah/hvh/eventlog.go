package hvh

import (
	"math/big"

	"github.com/icon-project/goloop/common/intconv"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/state"
)

const (
	SigRewardOffered  = "RewardOffered(int,int,int,int)"
	SigRewardClaimed  = "RewardClaimed(int,int,Address,int)"
	SigTermStarted    = "TermStarted(int)"
	SigICXIssued      = "ICXIssued(int,int,int)"
	SigHooverRefilled = "HooverRefilled(int,int,int)"
)

func onRewardOfferedEvent(
	cc hvhmodule.CallContext,
	termSeq, id int64, reward, hoover *big.Int) {
	cc.OnEvent(
		state.SystemAddress,
		[][]byte{[]byte(SigRewardOffered)},
		[][]byte{
			intconv.Int64ToBytes(termSeq),
			intconv.Int64ToBytes(id),
			intconv.BigIntToBytes(reward),
			intconv.BigIntToBytes(hoover),
		},
	)
}

func onRewardClaimedEvent(
	cc hvhmodule.CallContext, termSeq, id int64, owner module.Address, amount *big.Int) {
	cc.OnEvent(
		state.SystemAddress,
		[][]byte{[]byte(SigRewardClaimed)},
		[][]byte{
			intconv.Int64ToBytes(termSeq),
			intconv.Int64ToBytes(id),
			owner.Bytes(),
			intconv.BigIntToBytes(amount),
		},
	)
}

func onTermStartedEvent(cc hvhmodule.CallContext, termSeq int64) {
	cc.OnEvent(
		state.SystemAddress,
		[][]byte{[]byte(SigTermStarted)},
		[][]byte{intconv.Int64ToBytes(termSeq)},
	)
}

func onICXIssuedEvent(cc hvhmodule.CallContext, termSeq int64, amount, totalSupply *big.Int) {
	cc.OnEvent(state.SystemAddress,
		[][]byte{[]byte(SigICXIssued)},
		[][]byte{
			intconv.Int64ToBytes(termSeq),
			intconv.BigIntToBytes(amount),
			intconv.BigIntToBytes(totalSupply),
		},
	)
}

func onHooverRefilledEvent(cc hvhmodule.CallContext, amount, hooverBalance, sustainableFundBalance *big.Int) {
	cc.OnEvent(state.SystemAddress,
		[][]byte{[]byte(SigHooverRefilled)},
		[][]byte{
			intconv.BigIntToBytes(amount),
			intconv.BigIntToBytes(hooverBalance),
			intconv.BigIntToBytes(sustainableFundBalance),
		},
	)
}
