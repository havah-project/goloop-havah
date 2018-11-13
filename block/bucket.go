package block

import (
	"bytes"

	"github.com/icon-project/goloop/common"
	"github.com/icon-project/goloop/common/codec"
	"github.com/icon-project/goloop/common/crypto"
	"github.com/icon-project/goloop/common/db"
)

type bucket struct {
	dbBucket db.Bucket
	codec    codec.Codec
}

type raw []byte

func (b *bucket) _marshal(obj interface{}) ([]byte, error) {
	if bs, ok := obj.(raw); ok {
		return []byte(bs), nil
	}
	buf := bytes.NewBuffer(nil)
	err := b.codec.Marshal(buf, obj)
	return buf.Bytes(), err
}

func (b *bucket) get(key interface{}, value interface{}) error {
	bs, err := b.getBytes(key)
	if err != nil {
		return err
	}
	return b.codec.Unmarshal(bytes.NewBuffer(bs), value)
}

func (b *bucket) getBytes(key interface{}) ([]byte, error) {
	keyBS, err := b._marshal(key)
	if err != nil {
		return nil, err
	}
	bs, err := b.dbBucket.Get(keyBS)
	if bs == nil && err == nil {
		err = common.ErrNotFound
	}
	return bs, err
}

func (b *bucket) set(key interface{}, value interface{}) error {
	keyBS, err := b._marshal(key)
	if err != nil {
		return err
	}
	valueBS, err := b._marshal(value)
	if err != nil {
		return err
	}
	return b.dbBucket.Set(keyBS, valueBS)
}

func (b *bucket) put(value interface{}) error {
	valueBS, err := b._marshal(value)
	if err != nil {
		return err
	}
	keyBS := crypto.SHA3Sum256(valueBS)
	return b.dbBucket.Set(keyBS, valueBS)
}
