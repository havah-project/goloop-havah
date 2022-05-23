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
)

// From iiss.calculator.go
const (
	DayBlock     = 24 * 60 * 60 / 2
	DayPerMonth  = 30
	MonthBlock   = DayBlock * DayPerMonth
	MonthPerYear = 12
	YearBlock    = MonthBlock * MonthPerYear
)

const (
	BlockInterval    = 2000 // unit: ms
	RoundLimitFactor = 3

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
)

// PlatformConfig default values
const (
	TermPeriod          = DayBlock
	InitialIssueAmount  = 5_000_000                  // unit: hvh
	ReductionCycle      = MonthPerYear * DayPerMonth // 1 year (360) in term
	PrivateReleaseCycle = DayPerMonth
	PrivateLockup       = MonthPerYear * DayPerMonth // 1 year (360) in term
	HooverBudget        = 5_000_000                  // unit: hvh
)

// Addresses
const (
	SystemTreasury  = "hx1000000000000000000000000000000000000000"
	Governance      = "cx0000000000000000000000000000000000000001"
	PublicTreasury  = "hx3000000000000000000000000000000000000000"
	SustainableFund = "cx4000000000000000000000000000000000000000"
	CompanyTreasury = "cx5000000000000000000000000000000000000000"
	HooverFund      = "cx6000000000000000000000000000000000000000"
	EcoSystem       = "cx7000000000000000000000000000000000000000"
	PlanetNFT       = "cx8000000000000000000000000000000000000000"
)

// VarDB, DictDB, ArrayDB keys
const (
	VarIssueAmount         = "issue_amount"
	VarIssueStart          = "issue_start"
	VarTermPeriod          = "term_period"
	VarInitialIssueAmount  = "initial_issue_amount"
	VarIssueReductionCycle = "issue_reduction_cycle"
	VarPrivateReleaseCycle = "private_release_cycle"
	VarPrivateLockup       = "private_lockup"
	VarHooverBudget        = "hoover_budget"
	VarUSDTPrice           = "usdt_price"
	VarActiveUSDTPrice     = "active_usdt_price"
	DictPlanet             = "planet"
	ArrayPlanetManager     = "planet_manager"
	DictPlanetReward       = "planet_reward"
	VarAllPlanet           = "all_planet"
	VarActivePlanet        = "active_planet"
	VarRewardTotal         = "reward_total"
	VarRewardRemain        = "reward_remain"
	VarCompanyReward       = "company_reward"
)

// The following variables are read-only
var (
	BigIntZero      = new(big.Int)
	BigIntDecimal   = big.NewInt(1_000_000_000_000_000_000)
	BigIntDayBlocks = big.NewInt(DayBlock)
)
