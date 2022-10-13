package hvhmodule

import (
	"math/big"

	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/state"
	"github.com/icon-project/goloop/service/trace"
)

type WorldContext interface {
	Revision() module.Revision
	BlockHeight() int64
	Origin() module.Address
	Treasury() module.Address
	TransactionID() []byte
	ConsensusInfo() module.ConsensusInfo
	GetBalance(address module.Address) *big.Int
	Issue(address module.Address, amount *big.Int) (*big.Int, error)
	Burn(amount *big.Int) (*big.Int, error)
	Transfer(from module.Address, to module.Address, amount *big.Int, opType module.OpType) error
	GetTotalSupply() *big.Int
	SetValidators(validators []module.Validator) error
	StepPrice() *big.Int
	GetScoreOwner(score module.Address) (module.Address, error)
	SetScoreOwner(from module.Address, score module.Address, owner module.Address) error
	GetAccountState(id []byte) state.AccountState
}

type CallContext interface {
	WorldContext
	From() module.Address
	SumOfStepUsed() *big.Int
	OnEvent(addr module.Address, indexed, data [][]byte)
	Governance() module.Address
	FrameLogger() *trace.Logger
}
