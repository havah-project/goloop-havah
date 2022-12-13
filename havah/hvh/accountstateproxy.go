package hvh

import (
	"math/big"
	"strings"

	"github.com/icon-project/goloop/service/state"
)

type accountStateProxy struct {
	state.AccountState
	contract bool
	balance  *big.Int
	store    map[string][]byte
}

func (as *accountStateProxy) GetBalance() *big.Int {
	if as.balance == nil {
		return as.AccountState.GetBalance()
	}
	return as.balance
}

func (as *accountStateProxy) SetBalance(balance *big.Int) {
	as.balance = balance
}

func (as *accountStateProxy) lazyInitStore() {
	if as.store == nil {
		as.store = make(map[string][]byte)
	}
}

func (as *accountStateProxy) GetValue(k []byte) ([]byte, error) {
	if as.store != nil {
		if v, ok := as.store[string(k)]; ok {
			return v, nil
		}
	}
	return as.AccountState.GetValue(k)
}

func (as *accountStateProxy) SetValue(k, v []byte) ([]byte, error) {
	as.lazyInitStore()

	var err error
	ks := string(k)
	old, ok := as.store[ks]
	if !ok {
		old, err = as.AccountState.GetValue(k)
	}
	as.store[ks] = v
	return old, err
}

func (as *accountStateProxy) DeleteValue(k []byte) ([]byte, error) {
	as.lazyInitStore()

	ks := string(k)
	old := as.store[ks]
	as.store[ks] = nil
	return old, nil
}

func newAccountStateProxy(id []byte, as state.AccountState) *accountStateProxy {
	return &accountStateProxy{
		AccountState: as,
		contract:     strings.HasPrefix(string(id), "cx"),
	}
}
