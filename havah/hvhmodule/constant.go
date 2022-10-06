/*
 * Copyright 2021 ICON Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package hvhmodule

import (
	"math/big"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/module"
)

// From iiss.calculator.go
const (
	DayBlock     = 24 * 60 * 60 / 2
	DayPerMonth  = 30
	MonthBlock   = DayBlock * DayPerMonth
	MonthPerYear = 12
	YearBlock    = MonthBlock * MonthPerYear
	DayPerYear   = DayPerMonth * MonthPerYear
)

const (
	BlockInterval    = 2000 // unit: ms
	RoundLimitFactor = 3
	MaxPlanetCount   = 50000
	MaxCountToClaim  = 50

	StepPrice          = 12500000000
	MaxStepLimitInvoke = 2500000000
	MaxStepLimitQuery  = 50000000
	StepSchema         = 1
	StepApiCall        = 10000
	StepContractCall   = 25000
	StepContractCreate = 1000000000
	StepContractUpdate = 1000000000
	StepContractSet    = 15000
	StepDefault        = 100000
	StepDelete         = -240
	StepDeleteBase     = 200
	StepGet            = 25
	StepGetBase        = 3000
	StepInput          = 200
	StepLog            = 100
	StepLogBase        = 5000
	StepSet            = 320
	StepSetBase        = 10000

	// TimestampThreshold indicates threshold of timestamp of transactions
	// to be included in the block in millisecond.
	TimestampThreshold = 30_000
)

// PlatformConfig default values
const (
	TermPeriod           = DayBlock                 // unit: block
	IssueAmount          = 4_300_000                // unit: HVH, IssueAmount per term
	IssueReductionCycle  = DayPerYear               // 1 year (360) in term
	HooverBudget         = 4_300_000                // unit: HVH
	IssueLimit           = 50 * IssueReductionCycle // unit: term
	PrivateClaimableRate = MonthPerYear * 2         // 0 / 24
)

// VarDB, DictDB, ArrayDB keys
const (
	VarIssueAmount          = "issue_amount"          // unit: loop
	VarIssueStart           = "issue_start"           // block height
	VarIssueLimit           = "issue_limit"           // unit: term
	VarTermPeriod           = "term_period"           // unit: block
	VarIssueReductionCycle  = "issue_reduction_cycle" // unit: term
	VarHooverBudget         = "hoover_budget"         // unit: hvh
	VarUSDTPrice            = "usdt_price"            // unit: hvh
	VarActiveUSDTPrice      = "active_usdt_price"     // unit: hvh
	DictPlanet              = "planet"
	ArrayPlanetManager      = "planet_manager"
	DictPlanetReward        = "planet_reward"
	VarAllPlanet            = "all_planet"
	VarActivePlanet         = "active_planet"
	VarWorkingPlanet        = "working_planet"
	VarRewardTotal          = "reward_total"  // unit: hvh
	VarRewardRemain         = "reward_remain" // unit: hvh
	VarEcoReward            = "eco_reward"    // unit: hvh
	VarPrivateClaimableRate = "private_claimable_rate"
)

// VarDBs in SustainableFund Score
const (
	VarServiceFee    = "service_fee"
	VarTxFee         = "tx_fee"
	VarMissingReward = "missing_reward"
	VarHooverRefill  = "hoover_refill"
)

// ErrorCodes for havah chainscore
const (
	StatusIllegalArgument = module.StatusReverted + iota
	StatusNotFound
	StatusCriticalError
	StatusRewardError
	StatusNotReady
)

// The following variables are read-only
var (
	BigIntZero               = new(big.Int)
	BigIntCoinDecimal        = big.NewInt(1_000_000_000_000_000_000)
	BigIntUSDTDecimal        = big.NewInt(1_000_000)
	BigIntInitIssueAmount    = new(big.Int).Mul(big.NewInt(IssueAmount), BigIntCoinDecimal)
	BigIntHooverBudget       = new(big.Int).Mul(big.NewInt(HooverBudget), BigIntCoinDecimal)
	BigIntDayPerYear         = big.NewInt(DayPerYear)
	BigRatIssueReductionRate = big.NewRat(3, 10) // 30%
	BigRatEcoSystemToFee     = big.NewRat(1, 5)  // 20%
	// BigRatEcoSystemToCompanyReward is the EcoSystem proportion to company planet reward
	BigRatEcoSystemToCompanyReward = big.NewRat(3, 5) // 60%
)

var (
	PublicTreasury  = common.MustNewAddressFromString("hx3000000000000000000000000000000000000000")
	SustainableFund = common.MustNewAddressFromString("cx4000000000000000000000000000000000000000")
	CompanyTreasury = common.MustNewAddressFromString("cx5000000000000000000000000000000000000000")
	HooverFund      = common.MustNewAddressFromString("hx6000000000000000000000000000000000000000")
	EcoSystem       = common.MustNewAddressFromString("cx7000000000000000000000000000000000000000")
	PlanetNFT       = common.MustNewAddressFromString("cx8000000000000000000000000000000000000000")
	ServiceTreasury = common.MustNewAddressFromString("hx9000000000000000000000000000000000000000")
)
