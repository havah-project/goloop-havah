package havah

import (
	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/service/state"
)

type ChainConfig struct {
	BlockInterval    *common.HexInt32  `json:"blockInterval"`
	EcoSystem        *common.Address   `json:"ecoSystem"`
	HooverBudget     *common.HexInt64  `json:"hooverBudget"`
	USDTPrice        *common.HexInt64  `json:"usdtPrice"`
	Revision         *common.HexInt32  `json:"revision"`
	RoundLimitFactor *common.HexInt64  `json:"roundLimitFactor"`
	ValidatorList    []*common.Address `json:"validators"`
	Fee              *FeeConfig        `json:"fee"`
	Issue            *IssueConfig      `json:"issue"`
}

type FeeConfig struct {
	StepPrice common.HexInt              `json:"stepPrice"`
	StepLimit map[string]common.HexInt64 `json:"stepLimit,omitempty"`
	StepCosts map[string]common.HexInt64 `json:"stepCosts,omitempty"`
}

type IssueConfig struct {
	TermPeriod         *common.HexInt32 `json:"termPeriod"`         // 43200
	InitialIssueAmount *common.HexInt64 `json:"initialIssueAmount"` // 3920M
	ReductionCycle     *common.HexInt32 `json:"reductionCycle"`     // 365
	TotalPeriod        *common.HexInt32 `json:"totalPeriod"`        // 3650
}

func (s *chainScore) loadChainConfig() *ChainConfig {
	return newChainConfig()
}

func newChainConfig() *ChainConfig {
	cfg := &ChainConfig{
		RoundLimitFactor: &common.HexInt64{Value: hvhmodule.RoundLimitFactor},
		Fee:              newDefaultFeeConfig(),
		Issue:            newDefaultIssueConfig(),
	}
	return cfg
}

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

func newDefaultIssueConfig() *IssueConfig {
	cfg := new(IssueConfig)
	cfg.TermPeriod.Value = hvhmodule.TermPeriod
	cfg.InitialIssueAmount.Value = hvhmodule.IssueInitialAmount
	cfg.ReductionCycle.Value = hvhmodule.IssueReductionCycle
	cfg.TotalPeriod.Value = hvhmodule.IssueTotalPeriod
	return cfg
}
