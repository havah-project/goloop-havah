package hvh

import (
	"math/big"

	"github.com/icon-project/goloop/common/errors"
	"github.com/icon-project/goloop/common/intconv"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/havah/hvhutils"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/contract"
	"github.com/icon-project/goloop/service/scoredb"
	"github.com/icon-project/goloop/service/scoreresult"
	"github.com/icon-project/goloop/service/state"
	"github.com/icon-project/goloop/service/txresult"
)

func validateAmount(amount *big.Int) error {
	if amount == nil || amount.Sign() < 0 {
		return errors.Errorf("Invalid amount: %v", amount)
	}
	return nil
}

func setBalance(address module.Address, as state.AccountState, balance *big.Int) error {
	if balance.Sign() < 0 {
		return errors.Errorf(
			"Invalid balance: address=%v balance=%v",
			address, balance,
		)
	}
	as.SetBalance(balance)
	return nil
}

type callContextImpl struct {
	contract.CallContext
	from module.Address
}

func (ctx *callContextImpl) Origin() module.Address {
	return ctx.TransactionInfo().From
}

func (ctx *callContextImpl) GetBalance(address module.Address) *big.Int {
	account := ctx.GetAccountState(address.ID())
	return account.GetBalance()
}

func (ctx *callContextImpl) deposit(address module.Address, amount *big.Int) error {
	if err := validateAmount(amount); err != nil {
		return err
	}
	if amount.Sign() == 0 {
		return nil
	}
	return ctx.addBalance(address, amount)
}

func (ctx *callContextImpl) withdraw(address module.Address, amount *big.Int) error {
	if err := validateAmount(amount); err != nil {
		return err
	}
	if amount.Sign() == 0 {
		return nil
	}
	return ctx.addBalance(address, new(big.Int).Neg(amount))
}

func (ctx *callContextImpl) Issue(address module.Address, amount *big.Int) (*big.Int, error) {
	if address == nil {
		return nil, errors.IllegalArgumentError.New("Invalid address")
	}
	if amount == nil || amount.Sign() < 0 {
		return nil, errors.IllegalArgumentError.Errorf("Invalid issueAmount: %v", amount)
	}

	var err error
	var totalSupply *big.Int
	if amount.Sign() > 0 {
		totalSupply, err = ctx.addTotalSupply(amount)
		if err != nil {
			return nil, err
		}
		if err = ctx.deposit(address, amount); err != nil {
			return nil, err
		}
		ctx.onBalanceChange(module.Issue, nil, address, amount)
	} else {
		totalSupply = ctx.GetTotalSupply()
	}
	return totalSupply, nil
}

func (ctx *callContextImpl) Burn(amount *big.Int) (*big.Int, error) {
	if amount == nil || amount.Sign() < 0 {
		return nil, errors.IllegalArgumentError.Errorf("Invalid issueAmount: %v", amount)
	}

	var err error
	var totalSupply *big.Int
	if amount.Sign() > 0 {
		totalSupply, err = ctx.addTotalSupply(new(big.Int).Neg(amount))
		if err != nil {
			return nil, err
		}
		if err = ctx.withdraw(state.SystemAddress, amount); err != nil {
			return nil, err
		}
		ctx.onBalanceChange(module.Burn, state.SystemAddress, nil, amount)
	} else {
		totalSupply = ctx.GetTotalSupply()
	}
	return totalSupply, nil
}

func (ctx *callContextImpl) Transfer(
	from module.Address, to module.Address, amount *big.Int, opType module.OpType) (err error) {
	if err = validateAmount(amount); err != nil {
		return
	}
	if amount.Sign() == 0 || from.Equal(to) {
		return nil
	}
	// Subtract amount from the balance of "from" address
	if err = ctx.addBalance(from, new(big.Int).Neg(amount)); err != nil {
		return
	}
	// Add amount to "to" address
	if err = ctx.addBalance(to, amount); err != nil {
		return
	}
	ctx.onBalanceChange(opType, from, to, amount)
	ctx.CallContext.OnEvent(
		state.SystemAddress,
		[][]byte{
			[]byte(txresult.EventLogTransfer),
			from.Bytes(),
			to.Bytes(),
		},
		[][]byte{intconv.BigIntToBytes(amount)},
	)
	return
}

func (ctx *callContextImpl) addBalance(address module.Address, amount *big.Int) error {
	as := ctx.GetAccountState(address.ID())
	ob := as.GetBalance()
	return setBalance(address, as, new(big.Int).Add(ob, amount))
}

func (ctx *callContextImpl) GetTotalSupply() *big.Int {
	as := ctx.GetAccountState(state.SystemID)
	tsVar := scoredb.NewVarDB(as, state.VarTotalSupply)
	if ts := tsVar.BigInt(); ts != nil {
		return ts
	}
	return hvhmodule.BigIntZero
}

func (ctx *callContextImpl) addTotalSupply(amount *big.Int) (*big.Int, error) {
	as := ctx.GetAccountState(state.SystemID)
	varDB := scoredb.NewVarDB(as, state.VarTotalSupply)
	ts := varDB.BigInt()
	if ts == nil {
		ts = amount
	} else {
		ts = new(big.Int).Add(ts, amount)
	}
	if ts.Sign() < 0 {
		return nil, errors.Errorf("TotalSupply < 0")
	}
	return ts, varDB.Set(ts)
}

func (ctx *callContextImpl) SetValidators(validators []module.Validator) error {
	return ctx.GetValidatorState().Set(validators)
}

func (ctx *callContextImpl) GetScoreOwner(score module.Address) (module.Address, error) {
	if score == nil || !score.IsContract() {
		return nil, scoreresult.InvalidParameterError.Errorf("Invalid score address")
	}
	as := ctx.GetAccountState(score.ID())
	if hvhutils.IsNil(as) || !as.IsContract() {
		return nil, scoreresult.InvalidParameterError.Errorf("Score not found")
	}
	return as.ContractOwner(), nil
}

func (ctx *callContextImpl) SetScoreOwner(from module.Address, score module.Address, newOwner module.Address) error {
	// Parameter sanity check
	if from == nil {
		return scoreresult.InvalidParameterError.Errorf("Invalid sender")
	}
	if score == nil {
		return scoreresult.InvalidParameterError.Errorf("Invalid score address")
	}
	if !score.IsContract() {
		return errors.IllegalArgumentError.Errorf("Invalid score address")
	}
	if newOwner == nil {
		return scoreresult.InvalidParameterError.Errorf("Invalid owner")
	}

	as := ctx.GetAccountState(score.ID())
	if hvhutils.IsNil(as) || !as.IsContract() {
		return errors.IllegalArgumentError.Errorf("Score not found")
	}

	// Check if s.from is the owner of a given contract
	owner := as.ContractOwner()
	if owner == nil || !owner.Equal(from) {
		return scoreresult.AccessDeniedError.Errorf("Invalid owner")
	}

	// Check if the score is active
	if as.IsBlocked() {
		return scoreresult.AccessDeniedError.Errorf("Blocked score")
	}
	if as.IsDisabled() {
		return scoreresult.AccessDeniedError.Errorf("Disabled score")
	}
	return as.SetContractOwner(newOwner)
}

func (ctx *callContextImpl) From() module.Address {
	return ctx.from
}

func (ctx *callContextImpl) onBalanceChange(opType module.OpType, from, to module.Address, amount *big.Int) {
	if tlog := ctx.FrameLogger(); tlog != nil {
		tlog.OnBalanceChange(opType, from, to, amount)
	}
}

func NewCallContext(cc contract.CallContext, from module.Address) hvhmodule.CallContext {
	return &callContextImpl{
		CallContext: cc,
		from:        from,
	}
}
