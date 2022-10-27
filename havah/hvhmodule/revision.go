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

import "github.com/icon-project/goloop/module"

const (
	Revision0 = iota
	Revision1
	RevisionReserved
)

const (
	DefaultRevision = Revision0
	MaxRevision     = RevisionReserved - 1
	LatestRevision  = Revision0
)

var revisionFlags = []module.Revision{
	// Revision 0
	module.FixLostFeeByDeposit | module.InputCostingWithJSON | module.ExpandErrorCode | module.UseChainID |
		module.UseMPTOnEvents | module.UseCompactAPIInfo | module.PurgeEnumCache | module.FixMapValues,
	// Revision 1
	module.ContractSetEvent,
}

func init() {
	var revSum module.Revision
	for idx, rev := range revisionFlags {
		revSum ^= rev
		revisionFlags[idx] = revSum
	}
}

func ValueToRevision(v int) module.Revision {
	idx := v
	if v < Revision0 {
		v, idx = Revision0, 0
	}
	if idx >= len(revisionFlags) {
		idx = len(revisionFlags) - 1
	}
	return module.Revision(v) + revisionFlags[idx]
}
