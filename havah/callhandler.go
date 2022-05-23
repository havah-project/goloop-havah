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
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/contract"

	"github.com/icon-project/goloop/common/containerdb"
)

type callHandler struct {
	CallHandler
	to module.Address

	ext bool
}

func (h *callHandler) ExecuteAsync(cc contract.CallContext) (err error) {
	h.TLogStart()
	defer func() {
		if err != nil {
			if err2 := h.ApplyCallSteps(cc); err2 != nil {
				err = err2
			}
			h.TLogDone(err, cc.StepUsed(), nil)
		}
	}()

	var store containerdb.BytesStoreState
	return h.DoExecuteAsync(cc, h, store)
}

func newCallHandler(ch CallHandler, to module.Address, external bool) contract.ContractHandler {
	return &callHandler{CallHandler: ch, to: to, ext: external}
}
