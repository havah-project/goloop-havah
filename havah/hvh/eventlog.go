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
	SigIssued         = "Issued(int,int,int)"
	SigBurned         = "Burned(Address,int,int)"
	SigHooverRefilled = "HooverRefilled(int,int,int)"
	// LostDepositied(lostDelta int, lost int, reason str)
	SigLostDeposited = "LostDeposited(int,int,str)"
	// LostWithdrawn(to Address, amount int)
	SigLostWithdrawn = "LostWithdrawn(Address,int)"
	// TermStarted(termSeq int, planetCount int, rewardPerActivePlanet int)
	SigTermStarted = "TermStarted(int,int,int)"
	// Decentralized(activeValidatorCount int)
	SigDecentralized = "Decentralized(int)"
	// ValidatorLeaved(owner Address, node Address, reason string)
	SigValidatorLeaved = "ValidatorLeaved(Address,Address,str)"
	// ValidatorEntered(owner Address, node Address)
	SigValidatorEntered = "ValidatorEntered(Address,Address)"
	// ValidatorPenalized(owner Address, node Address)
	SigValidatorPenalized = "ValidatorPenalized(Address,Address)"
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

func onLostDepositedEvent(cc hvhmodule.CallContext, lostDelta, lost *big.Int, reason string) {
	cc.OnEvent(
		state.SystemAddress,
		[][]byte{[]byte(SigLostDeposited)},
		[][]byte{
			intconv.BigIntToBytes(lostDelta),
			intconv.BigIntToBytes(lost),
			[]byte(reason),
		},
	)
}

func onLostWithdrawnEvent(cc hvhmodule.CallContext, to module.Address, amount *big.Int) {
	cc.OnEvent(
		state.SystemAddress,
		[][]byte{[]byte(SigLostWithdrawn)},
		[][]byte{
			to.Bytes(),
			intconv.BigIntToBytes(amount),
		},
	)
}

func onTermStartedEvent(cc hvhmodule.CallContext, termSeq int64, planetCount, rewardPerActivePlanet *big.Int) {
	cc.OnEvent(
		state.SystemAddress,
		[][]byte{[]byte(SigTermStarted)},
		[][]byte{
			intconv.Int64ToBytes(termSeq),
			intconv.BigIntToBytes(planetCount),
			intconv.BigIntToBytes(rewardPerActivePlanet),
		},
	)
}

func onDecentralizedEvent(cc hvhmodule.CallContext, activeValidatorCount int64) {
	cc.OnEvent(
		state.SystemAddress,
		[][]byte{[]byte(SigDecentralized)},
		[][]byte{
			intconv.Int64ToBytes(activeValidatorCount),
		},
	)
}

// onValidatorLeaved is called when a validator leaved active validator set
func onValidatorLeaved(cc hvhmodule.CallContext, owner, node module.Address, reason string) {
	cc.OnEvent(
		state.SystemAddress,
		[][]byte{
			[]byte(SigValidatorLeaved),
			owner.Bytes(),
		},
		[][]byte{
			node.Bytes(),
			[]byte(reason),
		},
	)
}

// onValidatorEntered is called when a validator entered active validator set
func onValidatorEntered(cc hvhmodule.CallContext, owner, node module.Address) {
	cc.OnEvent(
		state.SystemAddress,
		[][]byte{
			[]byte(SigValidatorEntered),
			owner.Bytes(),
		},
		[][]byte{
			node.Bytes(),
		},
	)
}

// onValidatorPenalized is called when a validator leaved active validator set
func onValidatorPenalized(cc hvhmodule.CallContext, owner, node module.Address) {
	cc.OnEvent(
		state.SystemAddress,
		[][]byte{
			[]byte(SigValidatorPenalized),
			owner.Bytes(),
		},
		[][]byte{
			node.Bytes(),
		},
	)
}
