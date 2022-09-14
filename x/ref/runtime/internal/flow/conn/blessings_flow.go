// Copyright 2022 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package conn

import (
	"sync"

	"v.io/v23/context"
	"v.io/v23/security"
	"v.io/v23/verror"
	"v.io/v23/vom"
	iflow "v.io/x/ref/runtime/internal/flow"
)

type blessingsFlow struct {
	enc *vom.Encoder
	dec *vom.Decoder
	f   *flw

	mu       sync.Mutex
	nextKey  uint64
	incoming *inCache
	outgoing *outCache
}

// inCache keeps track of incoming blessings, discharges, and keys.
type inCache struct {
	dkeys      map[uint64]uint64               // bkey -> dkey of the latest discharges.
	blessings  map[uint64]security.Blessings   // keyed by bkey
	discharges map[uint64][]security.Discharge // keyed by dkey
}

// outCache keeps track of outgoing blessings, discharges, and keys.
type outCache struct {
	bkeys map[string]uint64 // blessings uid -> bkey

	dkeys      map[uint64]uint64               // blessings bkey -> dkey of latest discharges
	blessings  map[uint64]security.Blessings   // keyed by bkey
	discharges map[uint64][]security.Discharge // keyed by dkey
}

func newBlessingsFlow(f *flw) *blessingsFlow {
	b := &blessingsFlow{
		f:       f,
		enc:     vom.NewEncoder(f),
		dec:     vom.NewDecoder(f),
		nextKey: 1,
		incoming: &inCache{
			blessings:  make(map[uint64]security.Blessings),
			dkeys:      make(map[uint64]uint64),
			discharges: make(map[uint64][]security.Discharge),
		},
		outgoing: &outCache{
			bkeys:      make(map[string]uint64),
			dkeys:      make(map[uint64]uint64),
			blessings:  make(map[uint64]security.Blessings),
			discharges: make(map[uint64][]security.Discharge),
		},
	}
	return b
}

func (b *blessingsFlow) receiveBlessingsLocked(ctx *context.T, bkey uint64, blessings security.Blessings) error {
	b.f.useCurrentContext(ctx)
	// When accepting, make sure the blessings received are bound to the conn's
	// remote public key.
	if err := b.f.validateReceivedBlessings(ctx, blessings); err != nil {
		return err
	}
	b.incoming.blessings[bkey] = blessings
	return nil
}

func (b *blessingsFlow) receiveDischargesLocked(ctx *context.T, bkey, dkey uint64, discharges []security.Discharge) {
	b.incoming.discharges[dkey] = discharges
	b.incoming.dkeys[bkey] = dkey
}

func (b *blessingsFlow) receiveLocked(ctx *context.T, bd BlessingsFlowMessage) error {
	b.f.useCurrentContext(ctx)
	switch bd := bd.(type) {
	case BlessingsFlowMessageBlessings:
		bkey, blessings := bd.Value.BKey, bd.Value.Blessings
		if err := b.receiveBlessingsLocked(ctx, bkey, blessings); err != nil {
			return err
		}
	case BlessingsFlowMessageEncryptedBlessings:
		bkey, ciphertexts := bd.Value.BKey, bd.Value.Ciphertexts
		var blessings security.Blessings
		if err := decrypt(ctx, ciphertexts, &blessings); err != nil {
			// TODO(ataly): This error should not be returned if the
			// client has explicitly set the peer authorizer to nil.
			// In that case, the client does not care whether the server's
			// blessings can be decrypted or not. Ideally we should just
			// pass this error to the peer authorizer and handle it there.
			return iflow.MaybeWrapError(verror.ErrNotTrusted, ctx, ErrCannotDecryptBlessings.Errorf(ctx, "cannot decrypt the encrypted blessings sent by peer: %v", err))
		}
		if err := b.receiveBlessingsLocked(ctx, bkey, blessings); err != nil {
			return err
		}
	case BlessingsFlowMessageDischarges:
		bkey, dkey, discharges := bd.Value.BKey, bd.Value.DKey, bd.Value.Discharges
		b.receiveDischargesLocked(ctx, bkey, dkey, discharges)
	case BlessingsFlowMessageEncryptedDischarges:
		bkey, dkey, ciphertexts := bd.Value.BKey, bd.Value.DKey, bd.Value.Ciphertexts
		var discharges []security.Discharge
		if err := decrypt(ctx, ciphertexts, &discharges); err != nil {
			return iflow.MaybeWrapError(verror.ErrNotTrusted, ctx, ErrCannotDecryptDischarges.Errorf(ctx, "cannot decrypt the encrypted discharges sent by peer: %v", err))
		}
		b.receiveDischargesLocked(ctx, bkey, dkey, discharges)
	}
	return nil
}

// getRemote gets the remote blessings and discharges associated with the given
// bkey and dkey. We will read messages from the wire until we receive the
// looked for blessings. This method is normally called from the read loop of
// the conn, so all the packets for the desired blessing must have been received
// and buffered before calling this function.  This property is guaranteed since
// we always send blessings and discharges before sending their bkey/dkey
// references in the Auth message that terminates the auth handshake.
func (b *blessingsFlow) getRemote(ctx *context.T, bkey, dkey uint64) (security.Blessings, map[string]security.Discharge, error) {
	defer b.mu.Unlock()
	b.mu.Lock()
	b.f.useCurrentContext(ctx)
	for {
		blessings, hasB := b.incoming.blessings[bkey]
		if hasB {
			if dkey == 0 {
				return blessings, nil, nil
			}
			discharges, hasD := b.incoming.discharges[dkey]
			if hasD {
				return blessings, dischargeMap(discharges), nil
			}
		}

		var received BlessingsFlowMessage
		err := b.dec.Decode(&received)
		if err != nil {
			return security.Blessings{}, nil, err
		}
		if err := b.receiveLocked(ctx, received); err != nil {
			b.f.internalClose(ctx, false, false, err)
			return security.Blessings{}, nil, err
		}
	}
}

func (b *blessingsFlow) encodeBlessingsLocked(ctx *context.T, blessings security.Blessings, bkey uint64, peers []security.BlessingPattern) error {
	b.f.useCurrentContext(ctx)
	if len(peers) == 0 {
		// blessings can be encoded in plaintext
		return b.enc.Encode(BlessingsFlowMessageBlessings{Blessings{
			BKey:      bkey,
			Blessings: blessings,
		}})
	}
	ciphertexts, err := encrypt(ctx, peers, blessings)
	if err != nil {
		return ErrCannotEncryptBlessings.Errorf(ctx, "cannot encrypt blessings for peer: %v: %v", peers, err)
	}
	return b.enc.Encode(BlessingsFlowMessageEncryptedBlessings{EncryptedBlessings{
		BKey:        bkey,
		Ciphertexts: ciphertexts,
	}})
}

func (b *blessingsFlow) encodeDischargesLocked(ctx *context.T, discharges []security.Discharge, bkey, dkey uint64, peers []security.BlessingPattern) error {
	b.f.useCurrentContext(ctx)
	if len(peers) == 0 {
		// discharges can be encoded in plaintext
		return b.enc.Encode(BlessingsFlowMessageDischarges{Discharges{
			Discharges: discharges,
			DKey:       dkey,
			BKey:       bkey,
		}})
	}
	ciphertexts, err := encrypt(ctx, peers, discharges)
	if err != nil {
		return ErrCannotEncryptDischarges.Errorf(ctx, "cannot encrypt discharges for peers: %v: %v", peers, err)
	}
	return b.enc.Encode(BlessingsFlowMessageEncryptedDischarges{EncryptedDischarges{
		DKey:        dkey,
		BKey:        bkey,
		Ciphertexts: ciphertexts,
	}})
}

func (b *blessingsFlow) send(
	ctx *context.T,
	blessings security.Blessings,
	discharges map[string]security.Discharge,
	peers []security.BlessingPattern) (bkey, dkey uint64, err error) {
	if blessings.IsZero() {
		return 0, 0, nil
	}
	defer b.mu.Unlock()
	b.mu.Lock()
	b.f.useCurrentContext(ctx)

	buid := string(blessings.UniqueID())
	bkey, hasB := b.outgoing.bkeys[buid]
	if !hasB {
		bkey = b.nextKey
		b.nextKey++
		b.outgoing.bkeys[buid] = bkey
		b.outgoing.blessings[bkey] = blessings
		if err := b.encodeBlessingsLocked(ctx, blessings, bkey, peers); err != nil {
			return 0, 0, err
		}
	}
	if len(discharges) == 0 {
		return bkey, 0, nil
	}
	dkey, hasD := b.outgoing.dkeys[bkey]
	if hasD && equalDischarges(discharges, b.outgoing.discharges[dkey]) {
		return bkey, dkey, nil
	}
	dlist := dischargeList(discharges)
	dkey = b.nextKey
	b.nextKey++
	b.outgoing.dkeys[bkey] = dkey
	b.outgoing.discharges[dkey] = dlist
	return bkey, dkey, b.encodeDischargesLocked(ctx, dlist, bkey, dkey, peers)
}

func (b *blessingsFlow) close(ctx *context.T, err error) {
	b.f.useCurrentContext(ctx)
	b.f.close(ctx, false, err)
}

func dischargeList(in map[string]security.Discharge) []security.Discharge {
	out := make([]security.Discharge, 0, len(in))
	for _, d := range in {
		out = append(out, d)
	}
	return out
}
func dischargeMap(in []security.Discharge) map[string]security.Discharge {
	out := make(map[string]security.Discharge, len(in))
	for _, d := range in {
		out[d.ID()] = d
	}
	return out
}

func equalDischarges(m map[string]security.Discharge, s []security.Discharge) bool {
	if len(m) != len(s) {
		return false
	}
	for _, d := range s {
		inm, ok := m[d.ID()]
		if !ok || !d.Equivalent(inm) {
			return false
		}
	}
	return true
}