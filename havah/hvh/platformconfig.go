package hvh

import "github.com/icon-project/goloop/common"

type PlatformConfig struct {
	TermPeriod          *common.HexInt64 `json:"termPeriod"`          // 43200 in block
	IssueAmount         *common.HexInt   `json:"issueAmount"`         // 5M in hvh
	IssueReductionCycle *common.HexInt64 `json:"reductionCycle"`      // 360 in term
	PrivateReleaseCycle *common.HexInt64 `join:"privateReleaseCycle"` // 30 in term (1 month)
	PrivateLockup       *common.HexInt64 `join:"privateLockup"`       // 360 in term
	HooverBudget        *common.HexInt   `json:"hooverBudget"`        // unit: hvh
	USDTPrice           *common.HexInt   `json:"usdtPrice"`           // unit: hvh
}
