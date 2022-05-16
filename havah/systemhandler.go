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
	"math/big"

	"github.com/icon-project/goloop/common/codec"
	"github.com/icon-project/goloop/common/containerdb"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/contract"
	"github.com/icon-project/goloop/service/eeproxy"
	"github.com/icon-project/goloop/service/trace"
)

func doNotChargeContractCallStep(method string, revision int) bool {
	return true
}

type CallHandler interface {
	contract.AsyncContractHandler
	GetMethodName() string
	AllowExtra()
	DoExecuteAsync(cc contract.CallContext, ch eeproxy.CallContext, store containerdb.BytesStoreState) error
	TLogStart()
	TLogDone(status error, steps *big.Int, result *codec.TypedObj)
	ApplyCallSteps(cc contract.CallContext) error
}

type SystemCallHandler struct {
	CallHandler
	cc       contract.CallContext
	log      *trace.Logger
	revision module.Revision
}

func (h *SystemCallHandler) ExecuteAsync(cc contract.CallContext) (err error) {
	h.cc = cc
	h.revision = cc.Revision()
	h.log = h.TraceLogger()

	h.TLogStart()
	defer func() {
		if err != nil {
			// do not charge contractCall step for some external methods
			if !doNotChargeContractCallStep(h.GetMethodName(), h.revision.Value()) {
				// charge contractCall step if preprocessing is failed
				if err2 := h.ApplyCallSteps(cc); err2 != nil {
					err = err2
				}
			}
			h.TLogDone(err, cc.StepUsed(), nil)
		}
	}()

	return h.CallHandler.DoExecuteAsync(cc, h, nil)
}

func (h *SystemCallHandler) OnResult(status error, steps *big.Int, result *codec.TypedObj) {
	h.CallHandler.OnResult(status, steps, result)
}

func newSystemHandler(ch CallHandler) contract.ContractHandler {
	return &SystemCallHandler{CallHandler: ch}
}
