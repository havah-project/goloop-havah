package hvh

import "github.com/icon-project/goloop/common"

type PlatformConfig struct {
	TermPeriod          *common.HexInt32 `json:"termPeriod"`          // 43200 in block
	InitialIssueAmount  *common.HexInt   `json:"initialIssueAmount"`  // 5M in hvh
	IssueReductionCycle *common.HexInt32 `json:"reductionCycle"`      // 360 in term
	PrivateReleaseCycle *common.HexInt32 `join:"privateReleaseCycle"` // 30 in term (1 month)
	PrivateLockup       *common.HexInt32 `join:"privateLockup"`       // 360 in term
	HooverBudget        *common.HexInt   `json:"hooverBudget"`        // unit: hvh
	USDTPrice           *common.HexInt   `json:"usdtPrice"`           // unit: hvh
}
