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
	TermPeriod       = DayBlock

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

const (
	IssueInitialAmount  = 3_920_000_000
	IssueReductionCycle = MonthPerYear * DayPerMonth // 1 year (360)
	IssueTotalPeriod    = IssueReductionCycle * 10   // 10 years
)

// The following variables are read-only
var (
	BigIntZero      = new(big.Int)
	BigIntDecimal   = big.NewInt(1_000_000_000_000_000_000)
	BigIntDayBlocks = big.NewInt(DayBlock)
)
