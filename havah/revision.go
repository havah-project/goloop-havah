package havah

import (
	"github.com/icon-project/goloop/havah/hvhmodule"
)

type handleRevFunc func(s *chainScore) error
type revHandlerItem struct {
	rev int
	fn  handleRevFunc
}

var revHandlerTable = []revHandlerItem{
	{hvhmodule.RevisionBTP2, handleRevBTP2},
	{hvhmodule.RevisionFixMissingBTPPublicKey, handleRevFixMissingBTPPublicKey},
	{hvhmodule.RevisionFixVegaNetProblem, handleRevFixVegaNetProblem},
}

// DO NOT update revHandlerMap manually
var revHandlerMap = make(map[int][]revHandlerItem)

func init() {
	for _, item := range revHandlerTable {
		rev := item.rev
		items, _ := revHandlerMap[rev]
		revHandlerMap[rev] = append(items, item)
	}
	revHandlerTable = nil
}

func (s *chainScore) handleRevisionChange(oldRev, newRev int) error {
	s.log.Infof("handleRevisionChange %d->%d", oldRev, newRev)
	if oldRev >= newRev {
		return nil
	}

	for rev := oldRev + 1; rev <= newRev; rev++ {
		if items, ok := revHandlerMap[rev]; ok {
			for _, item := range items {
				if err := item.fn(s); err != nil {
					s.log.Errorf("handleRevFunc() error: rev=%d err=%v", rev, err)
					return err
				}
			}
		}
	}
	return nil
}

func initBTPPublicKeysFromValidators(s *chainScore) error {
	es, cc, err := s.getExtensionStateAndContext()
	if err != nil {
		return err
	}
	return es.InitBTPPublicKeys(cc)
}

func handleRevBTP2(s *chainScore) error {
	return initBTPPublicKeysFromValidators(s)
}

func handleRevFixMissingBTPPublicKey(s *chainScore) error {
	return initBTPPublicKeysFromValidators(s)
}

func handleRevFixVegaNetProblem(s *chainScore) error {
	return initBTPPublicKeysFromValidators(s)
}
