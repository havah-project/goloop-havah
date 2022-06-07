package havah

import (
	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/havah/hvh"
	"github.com/icon-project/goloop/havah/hvhmodule"
)

type chainConfig struct {
	BlockInterval    *common.HexInt32    `json:"blockInterval,omitempty"`
	Revision         *common.HexInt32    `json:"revision,omitempty"`
	RoundLimitFactor *common.HexInt32    `json:"roundLimitFactor,omitempty"`
	ValidatorList    []*common.Address   `json:"validators"`
	Fee              *FeeConfig          `json:"fee,omitempty"`
	Platform         *hvh.PlatformConfig `json:"platform"`
}

type FeeConfig struct {
	StepPrice common.HexInt              `json:"stepPrice"`
	StepLimit map[string]common.HexInt64 `json:"stepLimit,omitempty"`
	StepCosts map[string]common.HexInt64 `json:"stepCosts,omitempty"`
}

//type PlatformConfig struct {
//	TermPeriod          *common.HexInt32 `json:"termPeriod"`          // 43200 in block
//	IssueAmount  *common.HexInt   `json:"initialIssueAmount"`  // 5M in hvh
//	IssueReductionCycle *common.HexInt32 `json:"reductionCycle"`      // 360 in term
//	PrivateReleaseCycle *common.HexInt32 `join:"privateReleaseCycle"` // 30 in term (1 month)
//	PrivateLockup       *common.HexInt32 `join:"privateLockup"`       // 360 in term
//	HooverBudget        *common.HexInt   `json:"hooverBudget"`        // unit: hvh
//	USDTPrice           *common.HexInt   `json:"usdtPrice"`           // unit: hvh
//}

func newChainConfig() *chainConfig {
	return &chainConfig{
		BlockInterval:    &common.HexInt32{Value: hvhmodule.BlockInterval},
		Revision:         &common.HexInt32{Value: hvhmodule.Revision0},
		RoundLimitFactor: &common.HexInt32{Value: hvhmodule.RoundLimitFactor},
	}
}

/*
func newDefaultFeeConfig() *FeeConfig {
	cfg := new(FeeConfig)
	cfg.StepPrice.SetInt64(hvhmodule.StepPrice)
	cfg.StepLimit = map[string]common.HexInt64{
		state.StepLimitTypeInvoke: {hvhmodule.MaxStepLimitInvoke},
		state.StepLimitTypeQuery:  {hvhmodule.MaxStepLimitQuery},
	}
	cfg.StepCosts = map[string]common.HexInt64{
		state.StepTypeDefault:        {hvhmodule.StepDefault},
		state.StepTypeContractCall:   {hvhmodule.StepContractCall},
		state.StepTypeContractCreate: {hvhmodule.StepContractCreate},
		state.StepTypeContractUpdate: {hvhmodule.StepContractUpdate},
		state.StepTypeContractSet:    {hvhmodule.StepContractSet},
		state.StepTypeGet:            {hvhmodule.StepGet},
		state.StepTypeGetBase:        {hvhmodule.StepGetBase},
		state.StepTypeSet:            {hvhmodule.StepSet},
		state.StepTypeSetBase:        {hvhmodule.StepSetBase},
		state.StepTypeDelete:         {hvhmodule.StepDelete},
		state.StepTypeDeleteBase:     {hvhmodule.StepDeleteBase},
		state.StepTypeInput:          {hvhmodule.StepInput},
		state.StepTypeLog:            {hvhmodule.StepLog},
		state.StepTypeLogBase:        {hvhmodule.StepLogBase},
		state.StepTypeApiCall:        {hvhmodule.StepApiCall},
	}
	return cfg
}

func newDefaultPlatformConfig() *PlatformConfig {
	return &PlatformConfig{
		TermPeriod:         &common.HexInt32{Value: hvhmodule.TermPeriod},
		IssueReductionCycle:     &common.HexInt32{Value: hvhmodule.IssueReductionCycle},
		IssueAmount: &common.HexInt64{Value: hvhmodule.IssueAmount},
	}
}
*/
