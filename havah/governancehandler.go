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

package havah

import (
	"github.com/icon-project/goloop/common/codec"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/contract"
	"github.com/icon-project/goloop/service/state"
	"github.com/icon-project/goloop/service/trace"
)

type governanceHandler struct {
	ch    contract.ContractHandler
	ctype int
	call  *contract.DataCallJSON
	log   *trace.Logger
}

func (g *governanceHandler) Prepare(ctx contract.Context) (state.WorldContext, error) {
	lq := []state.LockRequest{
		{state.WorldIDStr, state.AccountWriteLock},
	}
	return ctx.GetFuture(lq), nil
}

func (g *governanceHandler) SetTraceLogger(logger *trace.Logger) {
	g.ch.SetTraceLogger(logger)
	g.log = logger
}

func (g *governanceHandler) TraceLogger() *trace.Logger {
	return g.log
}

func (g *governanceHandler) handleRevisionChange(cc contract.CallContext, r1, r2 int) {
	if r1 >= r2 {
		return
	}
}

func (g *governanceHandler) ExecuteSync(cc contract.CallContext) (error, *codec.TypedObj, module.Address) {
	g.log.TSystem("GOV start")
	defer g.log.TSystem("GOV end")

	status, steps, result, score := cc.Call(g.ch, cc.StepAvailable())
	cc.DeductSteps(steps)
	return status, result, score
}

func newGovernanceHandler(ch contract.ContractHandler) *governanceHandler {
	return &governanceHandler{ch: ch}
}
