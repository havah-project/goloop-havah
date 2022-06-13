package hvh

import "github.com/icon-project/goloop/common"

type PlatformConfig struct {
	TermPeriod          *common.HexInt64 `json:"termPeriod,omitempty"`          // 43200 in block
	IssueReductionCycle *common.HexInt64 `json:"reductionCycle,omitempty"`      // 360 in term
	PrivateReleaseCycle *common.HexInt64 `json:"privateReleaseCycle,omitempty"` // 30 in term (1 month)
	PrivateLockup       *common.HexInt64 `json:"privateLockup,omitempty"`       // 360 in term
	IssueLimit          *common.HexInt64 `json:"issueLimit,omitempty"`

	IssueAmount  *common.HexInt `json:"issueAmount,omitempty"`  // 5M in HVH
	HooverBudget *common.HexInt `json:"hooverBudget,omitempty"` // unit: HVH
	USDTPrice    *common.HexInt `json:"usdtPrice"`              // unit: HVH
}
