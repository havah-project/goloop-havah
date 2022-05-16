/*
 * Copyright 2020 ICON Foundation
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
	"github.com/icon-project/goloop/common/errors"
	"github.com/icon-project/goloop/havah/hvh"
	"github.com/icon-project/goloop/havah/hvhmodule"
	"github.com/icon-project/goloop/havah/hvhutils"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/contract"
	"github.com/icon-project/goloop/service/scoreresult"
)

func (s *chainScore) getExtensionState() (*hvh.ExtensionStateImpl, error) {
	es := s.cc.GetExtensionState()
	if es == nil {
		err := errors.Errorf("ExtensionState is nil")
		return nil, s.toScoreResultError(scoreresult.UnknownFailureError, err)
	}
	esi := es.(*hvh.ExtensionStateImpl)
	esi.SetLogger(hvhutils.NewLogger(s.cc.Logger()))
	return esi, nil
}

func (s *chainScore) toScoreResultError(code errors.Code, err error) error {
	msg := err.Error()
	if logger := s.cc.Logger(); logger != nil {
		logger = hvhutils.NewLogger(logger)
		logger.Infof(msg)
	}
	return code.Wrap(err, msg)
}

func (s *chainScore) Ex_getScoreOwner(score module.Address) (module.Address, error) {
	if err := s.tryChargeCall(); err != nil {
		return nil, err
	}
	return s.newCallContext(s.cc).GetScoreOwner(score)
}

func (s *chainScore) Ex_setScoreOwner(score module.Address, owner module.Address) error {
	if err := s.tryChargeCall(); err != nil {
		return err
	}
	return s.newCallContext(s.cc).SetScoreOwner(s.from, score, owner)
}

func (s *chainScore) newCallContext(cc contract.CallContext) hvhmodule.CallContext {
	return hvh.NewCallContext(cc, s.from)
}
