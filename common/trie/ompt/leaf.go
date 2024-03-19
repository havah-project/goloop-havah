package ompt

import (
	"bytes"
	"fmt"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/log"
	"github.com/icon-project/goloop/common/merkle"
	"github.com/icon-project/goloop/common/trie"
)

type leaf struct {
	nodeBase
	keys  []byte
	value trie.Object
}

func newLeaf(hash, serialized []byte, blist [][]byte, state nodeState) (node, error) {
	kbytes, err := rlpParseBytes(blist[0])
	if err != nil {
		return nil, err
	}
	keys := decodeKeys(kbytes)

	vbytes, err := rlpParseBytes(blist[1])
	if err != nil {
		return nil, err
	}
	value := bytesObject(vbytes)

	return &leaf{
		nodeBase: nodeBase{
			hashValue:  hash,
			serialized: serialized,
			state:      state,
		},
		keys:  keys,
		value: value,
	}, nil
}

func (n *leaf) getLink(fh bool) []byte {
	return n.nodeBase.getLink(n, fh)
}

func (n *leaf) toString() string {
	return fmt.Sprintf("L[%p](%v,[%x],%v)", n, n.state, n.keys, n.value)
}

func (n *leaf) dump() {
	log.Println(n.toString())
}

func (n *leaf) freeze() {
	lock := n.rlock()
	defer lock.Unlock()
	if n.state != stateDirty {
		return
	}
	lock.Migrate()
	if n.state == stateDirty {
		n.state = stateFrozen
	}
}

func (n *leaf) flush(m *mpt, nibs []byte) error {
	lock := n.rlock()
	defer lock.Unlock()

	if n.state == stateFlushed {
		return nil
	}
	if n.value == nil {
		return nil
	}
	if err := n.value.Flush(); err != nil {
		return err
	}
	if err := n.nodeBase.flushBaseInLock(m, nil); err != nil {
		return err
	}
	lock.Migrate()
	n.state = stateFlushed
	return nil
}

func (n *leaf) RLPListSize() int {
	return 2
}

func (n *leaf) RLPListEncode(e RLPEncoder) error {
	if err := e.RLPEncode(encodeKeys(0x20, n.keys)); err != nil {
		return err
	}
	if err := e.RLPEncode(n.value.Bytes()); err != nil {
		return err
	}
	return nil
}

func (n *leaf) getChanged(lock *AutoRWUnlock, keys []byte, o trie.Object) *leaf {
	if n.state == stateDirty {
		lock.Migrate()
		n.keys = keys
		n.value = o
		return n
	}
	return &leaf{keys: keys, value: o}
}

func (n *leaf) set(m *mpt, nibs []byte, depth int, o trie.Object) (node, bool, trie.Object, error) {
	keys := nibs[depth:]
	cnt, match := compareKeys(keys, n.keys)

	lock := n.rlock()
	defer lock.Unlock()

	switch {
	case cnt == 0 && !match:
		br := &branch{}
		if len(keys) == 0 {
			br.value = o
		} else {
			br.children[keys[0]] = &leaf{
				keys:  clone(keys[1:]),
				value: o,
			}
		}
		if len(n.keys) == 0 {
			br.value = n.value
		} else {
			idx := n.keys[0]
			br.children[idx] = n.getChanged(&lock, n.keys[1:], n.value)
		}
		return br, true, nil, nil
	case cnt < len(n.keys):
		br := &branch{}
		ext := &extension{keys: clone(keys[:cnt]), next: br}
		if cnt == len(keys) {
			br.value = o
		} else {
			br.children[keys[cnt]] = &leaf{keys: clone(keys[cnt+1:]), value: o}
		}
		idx := n.keys[cnt]
		br.children[idx] = n.getChanged(&lock, n.keys[cnt+1:], n.value)
		return ext, true, nil, nil
	case cnt < len(keys):
		br := &branch{}
		ext := &extension{keys: n.keys, next: br}
		br.value = n.value
		br.children[keys[cnt]] = &leaf{keys: clone(keys[cnt+1:]), value: o}
		return ext, true, nil, nil
	default:
		old := n.value
		if n.value.Equal(o) {
			return n, false, old, nil
		}
		return n.getChanged(&lock, n.keys, o), true, old, nil
	}
}

func (n *leaf) getKeyPrepended(k []byte) *leaf {
	lock := n.rlock()
	defer lock.Unlock()

	nk := make([]byte, len(k)+len(n.keys))
	copy(nk, k)
	copy(nk[len(k):], n.keys)
	return n.getChanged(&lock, nk, n.value)
}

func (n *leaf) delete(m *mpt, nibs []byte, depth int) (node, bool, trie.Object, error) {
	_, match := compareKeys(nibs[depth:], n.keys)
	if match {
		return nil, true, n.value, nil
	}
	return n, false, nil, nil
}

func (n *leaf) get(m *mpt, nibs []byte, depth int) (node, trie.Object, error) {
	lock := n.rlock()
	defer lock.Unlock()

	_, match := compareKeys(nibs[depth:], n.keys)
	if !match {
		return n, nil, nil
	}
	nv, changed, err := m.getObject(n.value)
	if changed {
		lock.Migrate()
		n.value = nv
	}
	return n, nv, err
}

func (n *leaf) realize(m *mpt) (node, error) {
	return n, nil
}

func (n *leaf) traverse(m *mpt, k string, v nodeScheduler) (string, trie.Object, error) {
	lock := n.rlock()
	defer lock.Unlock()

	value, changed, err := m.getObject(n.value)
	if changed {
		lock.Migrate()
		n.value = value
	}
	if err != nil {
		return "", nil, err
	}
	return k + string(n.keys), n.value, nil
}

func (n *leaf) getProof(m *mpt, keys []byte, items [][]byte) (node, [][]byte, error) {
	lock := n.rlock()
	defer lock.Unlock()

	if n.state < stateHashed {
		return n, nil, fmt.Errorf("IllegaState %s", n.toString())
	}
	if _, match := compareKeys(n.keys, keys); !match {
		return n, nil, nil
	}
	if n.hashValue != nil {
		items = append(items, n.serialized)
	}
	return n, items, nil
}

func (n *leaf) prove(m *mpt, keys []byte, proof [][]byte) (node, trie.Object, error) {
	lock := n.rlock()
	defer lock.Unlock()

	if n.hashValue != nil {
		if len(proof) != 1 || !bytes.Equal(proof[0], n.serialized) {
			return n, nil, common.ErrIllegalArgument
		}
	}

	_, match := compareKeys(n.keys, keys)
	if match {
		value, changed, err := m.getObject(n.value)
		if err != nil {
			return n, nil, err
		}
		if changed {
			lock.Migrate()
			n.value = value
		}
		return n, n.value, nil
	}
	return n, nil, common.ErrNotFound

}

func (n *leaf) resolve(m *mpt, bd merkle.Builder) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	nv, changed, err := m.getObject(n.value)
	if err != nil {
		return err
	}
	if changed {
		n.value = nv
	}
	if err := n.value.Resolve(bd); err != nil {
		return err
	}
	return nil
}

func (n *leaf) compact() node {
	lock := n.rlock()
	defer lock.Unlock()

	if n.state < stateFlushed {
		n.value.ClearCache()
		return n
	}
	if n.hashValue == nil {
		return n
	}
	return &hash{
		value: n.hashValue,
	}
}
