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
	SigRewardClaimed  = "RewardClaimed(Address,int,int,int)"
	SigTermStarted    = "TermStarted(int)"
	SigIssued         = "Issued(int,int,int)"
	SigBurned         = "Burned(Address,int,int)"
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
	cc hvhmodule.CallContext, owner module.Address, termSeq, id int64, amount *big.Int) {
	cc.OnEvent(
		state.SystemAddress,
		[][]byte{[]byte(SigRewardClaimed), owner.Bytes()},
		[][]byte{
			intconv.Int64ToBytes(termSeq),
			intconv.Int64ToBytes(id),
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

func onIssuedEvent(cc hvhmodule.CallContext, termSeq int64, amount, totalSupply *big.Int) {
	cc.OnEvent(
		state.SystemAddress,
		[][]byte{[]byte(SigIssued)},
		[][]byte{
			intconv.Int64ToBytes(termSeq),
			intconv.BigIntToBytes(amount),
			intconv.BigIntToBytes(totalSupply),
		},
	)
}

func onBurnedEvent(cc hvhmodule.CallContext, owner module.Address, amount, totalSupply *big.Int) {
	cc.OnEvent(
		state.SystemAddress,
		[][]byte{[]byte(SigBurned), owner.Bytes()},
		[][]byte{
			intconv.BigIntToBytes(amount),
			intconv.BigIntToBytes(totalSupply),
		},
	)
}

func onHooverRefilledEvent(cc hvhmodule.CallContext, amount, hooverBalance, sustainableFundBalance *big.Int) {
	cc.OnEvent(
		state.SystemAddress,
		[][]byte{[]byte(SigHooverRefilled)},
		[][]byte{
			intconv.BigIntToBytes(amount),
			intconv.BigIntToBytes(hooverBalance),
			intconv.BigIntToBytes(sustainableFundBalance),
		},
	)
}
