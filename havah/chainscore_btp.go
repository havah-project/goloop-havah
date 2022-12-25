package havah

import (
	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/intconv"
	"github.com/icon-project/goloop/module"
	"github.com/icon-project/goloop/service/scoreresult"
	"github.com/icon-project/goloop/service/state"
)

func (s *chainScore) Ex_getBTPNetworkTypeID(name string) (int64, error) {
	if err := s.tryChargeCall(); err != nil {
		return 0, err
	}
	return s.newBTPContext().GetNetworkTypeIDByName(name), nil
}

func (s *chainScore) Ex_getBTPPublicKey(address module.Address, name string) ([]byte, error) {
	if err := s.tryChargeCall(); err != nil {
		return nil, err
	}
	return s.newBTPContext().GetPublicKey(address, name), nil
}

func (s *chainScore) Ex_openBTPNetwork(networkTypeName string, name string, owner module.Address) (int64, error) {
	if err := s.checkGovernance(true); err != nil {
		return 0, err
	}
	if bs, err := s.getBTPState(); err != nil {
		return 0, err
	} else {
		bc := s.newBTPContext()
		ntActivated := false
		if bc.GetNetworkTypeIDByName(networkTypeName) <= 0 {
			ntActivated = true
		}
		ntid, nid, err := bs.OpenNetwork(bc, networkTypeName, name, owner)
		if err != nil {
			return 0, err
		}
		if ntActivated {
			s.cc.OnEvent(state.SystemAddress,
				[][]byte{
					[]byte("BTPNetworkTypeActivated(str,int)"),
					[]byte(networkTypeName),
					intconv.Int64ToBytes(ntid),
				},
				nil,
			)
		}
		s.cc.OnEvent(state.SystemAddress,
			[][]byte{
				[]byte("BTPNetworkOpened(int,int)"),
				intconv.Int64ToBytes(ntid),
				intconv.Int64ToBytes(nid),
			},
			nil,
		)
		return nid, nil
	}
}

func (s *chainScore) Ex_closeBTPNetwork(id *common.HexInt) error {
	if err := s.checkGovernance(true); err != nil {
		return err
	}
	nid := id.Int64()
	if bs, err := s.getBTPState(); err != nil {
		return err
	} else {
		if ntid, err := bs.CloseNetwork(s.newBTPContext(), nid); err != nil {
			return err
		} else {
			s.cc.OnEvent(state.SystemAddress,
				[][]byte{
					[]byte("BTPNetworkClosed(int,int)"),
					intconv.Int64ToBytes(ntid),
					intconv.Int64ToBytes(nid),
				},
				nil,
			)
		}
	}
	return nil
}

func (s *chainScore) Ex_sendBTPMessage(networkId *common.HexInt, message []byte) error {
	s.log.Tracef("%s send BTP Message(%v) to Network(%d)", s.from, common.HexPre(message), networkId)
	if err := s.tryChargeCall(); err != nil {
		return err
	}
	if len(message) == 0 {
		return scoreresult.ErrInvalidParameter
	}
	if bs, err := s.getBTPState(); err != nil {
		return err
	} else {
		nid := networkId.Int64()
		sn, err := bs.HandleMessage(s.newBTPContext(), s.from, nid)
		if err != nil {
			return err
		}
		s.cc.OnBTPMessage(nid, message)
		s.cc.OnEvent(state.SystemAddress,
			[][]byte{
				[]byte("BTPMessage(int,int)"),
				intconv.Int64ToBytes(nid),
				intconv.Int64ToBytes(sn),
			},
			nil,
		)
		return nil
	}
}

func (s *chainScore) Ex_setBTPPublicKey(name string, pubKey []byte) error {
	if err := s.tryChargeCall(); err != nil {
		return err
	}
	if s.from.IsContract() {
		return scoreresult.New(module.StatusAccessDenied, "NoPermission")
	}
	if bs, err := s.getBTPState(); err != nil {
		return err
	} else {
		if err = bs.SetPublicKey(s.newBTPContext(), s.from, name, pubKey); err != nil {
			return err
		}
	}
	return nil
}

func (s *chainScore) getBTPState() (*state.BTPStateImpl, error) {
	btpState := s.cc.GetBTPState()
	if btpState == nil {
		return nil, scoreresult.UnknownFailureError.Errorf("BTP state is nil")
	}
	return btpState.(*state.BTPStateImpl), nil
}

func (s *chainScore) newBTPContext() state.BTPContext {
	store := s.cc.GetAccountState(state.SystemID)
	return state.NewBTPContext(s.cc, store)
}
