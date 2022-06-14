package havah

import (
	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/havah/hvh"
	"github.com/icon-project/goloop/havah/hvhmodule"
)

type chainConfig struct {
	Platform         string              `json:"platform"`
	BlockInterval    *common.HexInt32    `json:"blockInterval,omitempty"`
	Revision         common.HexInt32     `json:"revision,omitempty"`
	RoundLimitFactor *common.HexInt32    `json:"roundLimitFactor,omitempty"`
	ValidatorList    []*common.Address   `json:"validators"`
	Fee              *FeeConfig          `json:"fee,omitempty"`
	TSThreshold      *common.HexInt32    `json:"timestampThreshold,omitempty"`
	Havah            *hvh.PlatformConfig `json:"havah"`
}

type FeeConfig struct {
	StepPrice common.HexInt              `json:"stepPrice"`
	StepLimit map[string]common.HexInt64 `json:"stepLimit,omitempty"`
	StepCosts map[string]common.HexInt64 `json:"stepCosts,omitempty"`
}

func newChainConfig() *chainConfig {
	return &chainConfig{
		BlockInterval:    &common.HexInt32{Value: hvhmodule.BlockInterval},
		RoundLimitFactor: &common.HexInt32{Value: hvhmodule.RoundLimitFactor},
		TSThreshold:      &common.HexInt32{Value: hvhmodule.TimestampThreshold},
	}
}
