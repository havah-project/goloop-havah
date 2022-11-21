/*
 * Copyright 2022 ICON Foundation
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

package ntm

import (
	"bytes"

	"github.com/icon-project/goloop/common/cache"
	"github.com/icon-project/goloop/common/codec"
	"github.com/icon-project/goloop/common/crypto"
	"github.com/icon-project/goloop/common/errors"
	"github.com/icon-project/goloop/common/log"
	"github.com/icon-project/goloop/module"
)

type secp256k1proofContextModule interface {
	UID() string
	AddressFromPubKey(pubKey []byte) ([]byte, error)
}

type secp256k1ProofPart struct {
	Index     int
	Signature *crypto.Signature
}

func (pp *secp256k1ProofPart) Bytes() []byte {
	return codec.MustMarshalToBytes(pp)
}

func (pp *secp256k1ProofPart) recover(mod secp256k1proofContextModule, hash []byte) ([]byte, error) {
	pubKey, err := pp.Signature.RecoverPublicKey(hash)
	if err != nil {
		return nil, err
	}
	return mod.AddressFromPubKey(pubKey.SerializeUncompressed())
}

type secp256k1Proof struct {
	Signatures []*crypto.Signature
	bytes      []byte
}

func newSecp256k1ProofFromBytes(bs []byte) (*secp256k1Proof, error) {
	var p secp256k1Proof
	_, err := codec.UnmarshalFromBytes(bs, &p)
	if err != nil {
		return nil, err
	}
	return &p, err
}

func (p *secp256k1Proof) Bytes() []byte {
	if p.bytes == nil {
		p.bytes = codec.MustMarshalToBytes(p)
	}
	return p.bytes
}

func (p *secp256k1Proof) Add(pp module.BTPProofPart) {
	epp := pp.(*secp256k1ProofPart)
	p.Signatures[epp.Index] = epp.Signature
}

func (p *secp256k1Proof) ValidatorCount() int {
	return len(p.Signatures)
}

func (p *secp256k1Proof) ProofPartAt(i int) module.BTPProofPart {
	if p.Signatures[i] == nil {
		return nil
	}
	return &secp256k1ProofPart{i, p.Signatures[i]}
}

type secp256k1ProofContext struct {
	Validators  [][]byte
	mod         *networkTypeModule
	bytes       cache.ByteSlice
	addrToIndex map[string]int
}

func newSecp256k1ProofContext(
	mod *networkTypeModule,
	keys [][]byte,
) (*secp256k1ProofContext, error) {
	pp := &secp256k1ProofContext{
		Validators:  make([][]byte, 0, len(keys)),
		addrToIndex: make(map[string]int, len(keys)),
		mod:         mod,
	}
	for i, key := range keys {
		var addr []byte
		var err error
		if key != nil {
			addr, err = mod.AddressFromPubKey(key)
			if err != nil {
				return nil, errors.Wrapf(err, "fail to converted key to address index=%d key=%x", i, key)
			}
			pp.addrToIndex[string(addr)] = i
		}
		pp.Validators = append(pp.Validators, addr)
	}
	return pp, nil
}

func (pc *secp256k1ProofContext) indexOf(address []byte) (int, bool) {
	if pc.addrToIndex == nil {
		pc.addrToIndex = make(map[string]int, len(pc.Validators))
		for i, addr := range pc.Validators {
			if addr != nil {
				pc.addrToIndex[string(addr)] = i
			}
		}
	}
	idx, ok := pc.addrToIndex[string(address)]
	return idx, ok
}

func newSecp256k1ProofContextFromBytes(
	mod *networkTypeModule,
	bytes []byte,
) (*secp256k1ProofContext, error) {
	pc := &secp256k1ProofContext{
		mod: mod,
	}
	if bytes != nil {
		_, err := codec.UnmarshalFromBytes(bytes, pc)
		if err != nil {
			return nil, err
		}
	}
	return pc, nil
}

func (pc *secp256k1ProofContext) NetworkTypeModule() module.NetworkTypeModule {
	return pc.mod
}

func (pc *secp256k1ProofContext) Bytes() []byte {
	return pc.bytes.Get(func() []byte {
		if pc.Validators == nil {
			return nil
		}
		return codec.MustMarshalToBytes(pc)
	})
}

// VerifyPart returns validator index and error
func (pc *secp256k1ProofContext) VerifyPart(dHash []byte, pp module.BTPProofPart) (int, error) {
	epp := pp.(*secp256k1ProofPart)
	if epp.Index < 0 || epp.Index >= len(pc.Validators) {
		return -1, errors.Errorf("invalid proof part index=%d numValidators=%d", epp.Index, len(pc.Validators))
	}
	addr, err := epp.recover(pc.mod, dHash)
	if err != nil {
		return -1, err
	}
	if !bytes.Equal(pc.Validators[epp.Index], addr) {
		idx, ok := pc.indexOf(addr)
		if ok {
			return -1, errors.Errorf("invalid proof part. maybe vote index is wrong? vote.index=%d vote.addr=%x pc.validators[%d]=%x pc.validators[%d]=%x", epp.Index, addr, epp.Index, pc.Validators[epp.Index], idx, addr)
		} else {
			return -1, errors.Errorf("invalid proof part. not a validator. vote.index=%d vote.addr=%x pc.validators[%d]=%x", epp.Index, addr, epp.Index, pc.Validators[epp.Index])
		}
	}
	return epp.Index, nil
}

func (pc *secp256k1ProofContext) NewProofPartFromBytes(ppBytes []byte) (module.BTPProofPart, error) {
	var pp secp256k1ProofPart
	_, err := codec.UnmarshalFromBytes(ppBytes, &pp)
	if err != nil {
		return nil, err
	}
	return &pp, err
}

func (pc *secp256k1ProofContext) Verify(dHash []byte, p module.BTPProof) error {
	ep := p.(*secp256k1Proof)
	set := make(map[int]struct{}, len(ep.Signatures))
	valid := 0
	for i, sig := range ep.Signatures {
		if sig == nil {
			continue
		}
		epp := secp256k1ProofPart{
			Index:     i,
			Signature: sig,
		}
		_, err := pc.VerifyPart(dHash, &epp)
		if err != nil {
			return err
		}
		if _, ok := set[epp.Index]; ok {
			addr, _ := epp.recover(pc.mod, dHash)
			return errors.Errorf("duplicated proof parts validator index=%d addr=%x", epp.Index, addr)
		}
		set[epp.Index] = struct{}{}
		valid++
	}
	if valid <= 2*len(pc.Validators)/3 {
		return errors.Errorf("not enough proof parts numValidator=%d numProofParts=%d", len(pc.Validators), len(ep.Signatures))
	}
	return nil
}

func (pc *secp256k1ProofContext) NewProofFromBytes(proofBytes []byte) (module.BTPProof, error) {
	return newSecp256k1ProofFromBytes(proofBytes)
}

func (pc *secp256k1ProofContext) NewProofPart(
	dHash []byte,
	wp module.WalletProvider,
) (module.BTPProofPart, error) {
	w := wp.WalletFor(secp256k1DSA)
	if w == nil {
		return nil, errors.Errorf("no wallet for uid=%s dsa=%s", pc.mod.UID(), secp256k1DSA)
	}
	addr, err := pc.mod.AddressFromPubKey(w.PublicKey())
	if err != nil {
		return nil, err
	}
	sig, err := w.Sign(dHash)
	if err != nil {
		return nil, err
	}
	idx, ok := pc.indexOf(addr)
	if !ok {
		return nil, errors.Errorf("not validator addr=%x", addr)
	}
	pp := secp256k1ProofPart{
		Index: idx,
	}
	pp.Signature, err = crypto.ParseSignature(sig)
	log.Must(err)
	return &pp, nil
}

func (pc *secp256k1ProofContext) DSA() string {
	return secp256k1DSA
}

func (pc *secp256k1ProofContext) NewProof() module.BTPProof {
	return &secp256k1Proof{
		Signatures: make([]*crypto.Signature, len(pc.Validators)),
	}
}
