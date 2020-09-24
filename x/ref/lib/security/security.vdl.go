// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file was auto-generated by the vanadium vdl tool.
// Package: security

//nolint:golint
package security

import (
	"fmt"
	"time"

	"v.io/v23/security"
	"v.io/v23/vdl"
	vdltime "v.io/v23/vdlroot/time"
)

var _ = initializeVDL() // Must be first; see initializeVDL comments for details.

//////////////////////////////////////////////////
// Type definitions

type blessingRootsState map[string][]security.BlessingPattern

func (blessingRootsState) VDLReflect(struct {
	Name string `vdl:"v.io/x/ref/lib/security.blessingRootsState"`
}) {
}

func (x blessingRootsState) VDLIsZero() bool { //nolint:gocyclo
	return len(x) == 0
}

func (x blessingRootsState) VDLWrite(enc vdl.Encoder) error { //nolint:gocyclo
	if err := enc.StartValue(vdlTypeMap1); err != nil {
		return err
	}
	if err := enc.SetLenHint(len(x)); err != nil {
		return err
	}
	for key, elem := range x {
		if err := enc.NextEntryValueString(vdl.StringType, key); err != nil {
			return err
		}
		if err := vdlWriteAnonList1(enc, elem); err != nil {
			return err
		}
	}
	if err := enc.NextEntry(true); err != nil {
		return err
	}
	return enc.FinishValue()
}

func vdlWriteAnonList1(enc vdl.Encoder, x []security.BlessingPattern) error {
	if err := enc.StartValue(vdlTypeList2); err != nil {
		return err
	}
	if err := enc.SetLenHint(len(x)); err != nil {
		return err
	}
	for _, elem := range x {
		if err := enc.NextEntryValueString(vdlTypeString3, string(elem)); err != nil {
			return err
		}
	}
	if err := enc.NextEntry(true); err != nil {
		return err
	}
	return enc.FinishValue()
}

func (x *blessingRootsState) VDLRead(dec vdl.Decoder) error { //nolint:gocyclo
	if err := dec.StartValue(vdlTypeMap1); err != nil {
		return err
	}
	var tmpMap blessingRootsState
	if len := dec.LenHint(); len > 0 {
		tmpMap = make(blessingRootsState, len)
	}
	for {
		switch done, key, err := dec.NextEntryValueString(); {
		case err != nil:
			return err
		case done:
			*x = tmpMap
			return dec.FinishValue()
		default:
			var elem []security.BlessingPattern
			if err := vdlReadAnonList1(dec, &elem); err != nil {
				return err
			}
			if tmpMap == nil {
				tmpMap = make(blessingRootsState)
			}
			tmpMap[key] = elem
		}
	}
}

func vdlReadAnonList1(dec vdl.Decoder, x *[]security.BlessingPattern) error {
	if err := dec.StartValue(vdlTypeList2); err != nil {
		return err
	}
	if len := dec.LenHint(); len > 0 {
		*x = make([]security.BlessingPattern, 0, len)
	} else {
		*x = nil
	}
	for {
		switch done, elem, err := dec.NextEntryValueString(); {
		case err != nil:
			return err
		case done:
			return dec.FinishValue()
		default:
			*x = append(*x, security.BlessingPattern(elem))
		}
	}
}

type dischargeCacheKey [32]byte

func (dischargeCacheKey) VDLReflect(struct {
	Name string `vdl:"v.io/x/ref/lib/security.dischargeCacheKey"`
}) {
}

func (x dischargeCacheKey) VDLIsZero() bool { //nolint:gocyclo
	return x == dischargeCacheKey{}
}

func (x dischargeCacheKey) VDLWrite(enc vdl.Encoder) error { //nolint:gocyclo
	if err := enc.WriteValueBytes(vdlTypeArray4, x[:]); err != nil {
		return err
	}
	return nil
}

func (x *dischargeCacheKey) VDLRead(dec vdl.Decoder) error { //nolint:gocyclo
	bytes := x[:]
	if err := dec.ReadValueBytes(32, &bytes); err != nil {
		return err
	}
	return nil
}

type CachedDischarge struct {
	Discharge security.Discharge
	// CacheTime is the time at which the discharge was first cached.
	CacheTime time.Time
}

func (CachedDischarge) VDLReflect(struct {
	Name string `vdl:"v.io/x/ref/lib/security.CachedDischarge"`
}) {
}

func (x CachedDischarge) VDLIsZero() bool { //nolint:gocyclo
	if !x.Discharge.VDLIsZero() {
		return false
	}
	if !x.CacheTime.IsZero() {
		return false
	}
	return true
}

func (x CachedDischarge) VDLWrite(enc vdl.Encoder) error { //nolint:gocyclo
	if err := enc.StartValue(vdlTypeStruct5); err != nil {
		return err
	}
	if !x.Discharge.VDLIsZero() {
		if err := enc.NextField(0); err != nil {
			return err
		}
		var wire security.WireDischarge
		if err := security.WireDischargeFromNative(&wire, x.Discharge); err != nil {
			return err
		}
		if err := wire.VDLWrite(enc); err != nil {
			return err
		}
	}
	if !x.CacheTime.IsZero() {
		if err := enc.NextField(1); err != nil {
			return err
		}
		var wire vdltime.Time
		if err := vdltime.TimeFromNative(&wire, x.CacheTime); err != nil {
			return err
		}
		if err := wire.VDLWrite(enc); err != nil {
			return err
		}
	}
	if err := enc.NextField(-1); err != nil {
		return err
	}
	return enc.FinishValue()
}

func (x *CachedDischarge) VDLRead(dec vdl.Decoder) error { //nolint:gocyclo
	*x = CachedDischarge{}
	if err := dec.StartValue(vdlTypeStruct5); err != nil {
		return err
	}
	decType := dec.Type()
	for {
		index, err := dec.NextField()
		switch {
		case err != nil:
			return err
		case index == -1:
			return dec.FinishValue()
		}
		if decType != vdlTypeStruct5 {
			index = vdlTypeStruct5.FieldIndexByName(decType.Field(index).Name)
			if index == -1 {
				if err := dec.SkipValue(); err != nil {
					return err
				}
				continue
			}
		}
		switch index {
		case 0:
			var wire security.WireDischarge
			if err := security.VDLReadWireDischarge(dec, &wire); err != nil {
				return err
			}
			if err := security.WireDischargeToNative(wire, &x.Discharge); err != nil {
				return err
			}
		case 1:
			var wire vdltime.Time
			if err := wire.VDLRead(dec); err != nil {
				return err
			}
			if err := vdltime.TimeToNative(wire, &x.CacheTime); err != nil {
				return err
			}
		}
	}
}

type blessingStoreState struct {
	// PeerBlessings maps BlessingPatterns to the Blessings object that is to
	// be shared with peers which present blessings of their own that match the
	// pattern.
	//
	// All blessings bind to the same public key.
	PeerBlessings map[security.BlessingPattern]security.Blessings
	// DefaultBlessings is the default Blessings to be shared with peers for which
	// no other information is available to select blessings.
	DefaultBlessings security.Blessings
	// DischargeCache is the cache of discharges.
	// TODO(mattr): This map is deprecated in favor of the Discharges map below.
	DischargeCache map[dischargeCacheKey]security.Discharge
	// DischargeCache is the cache of discharges.
	Discharges map[dischargeCacheKey]CachedDischarge
	// CacheKeyFormat is the dischargeCacheKey format version. It should incremented
	// any time the format of the dischargeCacheKey is changed.
	CacheKeyFormat uint32
}

func (blessingStoreState) VDLReflect(struct {
	Name string `vdl:"v.io/x/ref/lib/security.blessingStoreState"`
}) {
}

func (x blessingStoreState) VDLIsZero() bool { //nolint:gocyclo
	if len(x.PeerBlessings) != 0 {
		return false
	}
	if !x.DefaultBlessings.IsZero() {
		return false
	}
	if len(x.DischargeCache) != 0 {
		return false
	}
	if len(x.Discharges) != 0 {
		return false
	}
	if x.CacheKeyFormat != 0 {
		return false
	}
	return true
}

func (x blessingStoreState) VDLWrite(enc vdl.Encoder) error { //nolint:gocyclo
	if err := enc.StartValue(vdlTypeStruct8); err != nil {
		return err
	}
	if len(x.PeerBlessings) != 0 {
		if err := enc.NextField(0); err != nil {
			return err
		}
		if err := vdlWriteAnonMap2(enc, x.PeerBlessings); err != nil {
			return err
		}
	}
	if !x.DefaultBlessings.IsZero() {
		if err := enc.NextField(1); err != nil {
			return err
		}
		var wire security.WireBlessings
		if err := security.WireBlessingsFromNative(&wire, x.DefaultBlessings); err != nil {
			return err
		}
		if err := wire.VDLWrite(enc); err != nil {
			return err
		}
	}
	if len(x.DischargeCache) != 0 {
		if err := enc.NextField(2); err != nil {
			return err
		}
		if err := vdlWriteAnonMap3(enc, x.DischargeCache); err != nil {
			return err
		}
	}
	if len(x.Discharges) != 0 {
		if err := enc.NextField(3); err != nil {
			return err
		}
		if err := vdlWriteAnonMap4(enc, x.Discharges); err != nil {
			return err
		}
	}
	if x.CacheKeyFormat != 0 {
		if err := enc.NextFieldValueUint(4, vdl.Uint32Type, uint64(x.CacheKeyFormat)); err != nil {
			return err
		}
	}
	if err := enc.NextField(-1); err != nil {
		return err
	}
	return enc.FinishValue()
}

func vdlWriteAnonMap2(enc vdl.Encoder, x map[security.BlessingPattern]security.Blessings) error {
	if err := enc.StartValue(vdlTypeMap9); err != nil {
		return err
	}
	if err := enc.SetLenHint(len(x)); err != nil {
		return err
	}
	for key, elem := range x {
		if err := enc.NextEntryValueString(vdlTypeString3, string(key)); err != nil {
			return err
		}
		var wire security.WireBlessings
		if err := security.WireBlessingsFromNative(&wire, elem); err != nil {
			return err
		}
		if err := wire.VDLWrite(enc); err != nil {
			return err
		}
	}
	if err := enc.NextEntry(true); err != nil {
		return err
	}
	return enc.FinishValue()
}

func vdlWriteAnonMap3(enc vdl.Encoder, x map[dischargeCacheKey]security.Discharge) error {
	if err := enc.StartValue(vdlTypeMap11); err != nil {
		return err
	}
	if err := enc.SetLenHint(len(x)); err != nil {
		return err
	}
	for key, elem := range x {
		if err := enc.NextEntryValueBytes(vdlTypeArray4, key[:]); err != nil {
			return err
		}
		var wire security.WireDischarge
		if err := security.WireDischargeFromNative(&wire, elem); err != nil {
			return err
		}
		switch {
		case wire == nil:
			// Write the zero value of the union type.
			if err := vdl.ZeroValue(vdlTypeUnion6).VDLWrite(enc); err != nil {
				return err
			}
		default:
			if err := wire.VDLWrite(enc); err != nil {
				return err
			}
		}
	}
	if err := enc.NextEntry(true); err != nil {
		return err
	}
	return enc.FinishValue()
}

func vdlWriteAnonMap4(enc vdl.Encoder, x map[dischargeCacheKey]CachedDischarge) error {
	if err := enc.StartValue(vdlTypeMap12); err != nil {
		return err
	}
	if err := enc.SetLenHint(len(x)); err != nil {
		return err
	}
	for key, elem := range x {
		if err := enc.NextEntryValueBytes(vdlTypeArray4, key[:]); err != nil {
			return err
		}
		if err := elem.VDLWrite(enc); err != nil {
			return err
		}
	}
	if err := enc.NextEntry(true); err != nil {
		return err
	}
	return enc.FinishValue()
}

func (x *blessingStoreState) VDLRead(dec vdl.Decoder) error { //nolint:gocyclo
	*x = blessingStoreState{}
	if err := dec.StartValue(vdlTypeStruct8); err != nil {
		return err
	}
	decType := dec.Type()
	for {
		index, err := dec.NextField()
		switch {
		case err != nil:
			return err
		case index == -1:
			return dec.FinishValue()
		}
		if decType != vdlTypeStruct8 {
			index = vdlTypeStruct8.FieldIndexByName(decType.Field(index).Name)
			if index == -1 {
				if err := dec.SkipValue(); err != nil {
					return err
				}
				continue
			}
		}
		switch index {
		case 0:
			if err := vdlReadAnonMap2(dec, &x.PeerBlessings); err != nil {
				return err
			}
		case 1:
			var wire security.WireBlessings
			if err := wire.VDLRead(dec); err != nil {
				return err
			}
			if err := security.WireBlessingsToNative(wire, &x.DefaultBlessings); err != nil {
				return err
			}
		case 2:
			if err := vdlReadAnonMap3(dec, &x.DischargeCache); err != nil {
				return err
			}
		case 3:
			if err := vdlReadAnonMap4(dec, &x.Discharges); err != nil {
				return err
			}
		case 4:
			switch value, err := dec.ReadValueUint(32); {
			case err != nil:
				return err
			default:
				x.CacheKeyFormat = uint32(value)
			}
		}
	}
}

func vdlReadAnonMap2(dec vdl.Decoder, x *map[security.BlessingPattern]security.Blessings) error {
	if err := dec.StartValue(vdlTypeMap9); err != nil {
		return err
	}
	var tmpMap map[security.BlessingPattern]security.Blessings
	if len := dec.LenHint(); len > 0 {
		tmpMap = make(map[security.BlessingPattern]security.Blessings, len)
	}
	for {
		switch done, key, err := dec.NextEntryValueString(); {
		case err != nil:
			return err
		case done:
			*x = tmpMap
			return dec.FinishValue()
		default:
			var elem security.Blessings
			var wire security.WireBlessings
			if err := wire.VDLRead(dec); err != nil {
				return err
			}
			if err := security.WireBlessingsToNative(wire, &elem); err != nil {
				return err
			}
			if tmpMap == nil {
				tmpMap = make(map[security.BlessingPattern]security.Blessings)
			}
			tmpMap[security.BlessingPattern(key)] = elem
		}
	}
}

func vdlReadAnonMap3(dec vdl.Decoder, x *map[dischargeCacheKey]security.Discharge) error {
	if err := dec.StartValue(vdlTypeMap11); err != nil {
		return err
	}
	var tmpMap map[dischargeCacheKey]security.Discharge
	if len := dec.LenHint(); len > 0 {
		tmpMap = make(map[dischargeCacheKey]security.Discharge, len)
	}
	for {
		switch done, err := dec.NextEntry(); {
		case err != nil:
			return err
		case done:
			*x = tmpMap
			return dec.FinishValue()
		default:
			var key dischargeCacheKey
			bytes := key[:]
			if err := dec.ReadValueBytes(32, &bytes); err != nil {
				return err
			}
			var elem security.Discharge
			var wire security.WireDischarge
			if err := security.VDLReadWireDischarge(dec, &wire); err != nil {
				return err
			}
			if err := security.WireDischargeToNative(wire, &elem); err != nil {
				return err
			}
			if tmpMap == nil {
				tmpMap = make(map[dischargeCacheKey]security.Discharge)
			}
			tmpMap[key] = elem
		}
	}
}

func vdlReadAnonMap4(dec vdl.Decoder, x *map[dischargeCacheKey]CachedDischarge) error {
	if err := dec.StartValue(vdlTypeMap12); err != nil {
		return err
	}
	var tmpMap map[dischargeCacheKey]CachedDischarge
	if len := dec.LenHint(); len > 0 {
		tmpMap = make(map[dischargeCacheKey]CachedDischarge, len)
	}
	for {
		switch done, err := dec.NextEntry(); {
		case err != nil:
			return err
		case done:
			*x = tmpMap
			return dec.FinishValue()
		default:
			var key dischargeCacheKey
			bytes := key[:]
			if err := dec.ReadValueBytes(32, &bytes); err != nil {
				return err
			}
			var elem CachedDischarge
			if err := elem.VDLRead(dec); err != nil {
				return err
			}
			if tmpMap == nil {
				tmpMap = make(map[dischargeCacheKey]CachedDischarge)
			}
			tmpMap[key] = elem
		}
	}
}

type paramListIterator struct {
	err      error
	idx, max int
	params   []interface{}
}

func (pl *paramListIterator) next() (interface{}, error) {
	if pl.err != nil {
		return nil, pl.err
	}
	if pl.idx+1 > pl.max {
		pl.err = fmt.Errorf("too few parameters: have %v", pl.max)
		return nil, pl.err
	}
	pl.idx++
	return pl.params[pl.idx-1], nil
}

func (pl *paramListIterator) preamble() (component, operation string, err error) {
	var tmp interface{}
	if tmp, err = pl.next(); err != nil {
		return
	}
	var ok bool
	if component, ok = tmp.(string); !ok {
		return "", "", fmt.Errorf("ParamList[0]: component name is not a string: %T", tmp)
	}
	if tmp, err = pl.next(); err != nil {
		return
	}
	if operation, ok = tmp.(string); !ok {
		return "", "", fmt.Errorf("ParamList[1]: operation name is not a string: %T", tmp)
	}
	return
}

// Hold type definitions in package-level variables, for better performance.
//nolint:unused
var (
	vdlTypeMap1     *vdl.Type
	vdlTypeList2    *vdl.Type
	vdlTypeString3  *vdl.Type
	vdlTypeArray4   *vdl.Type
	vdlTypeStruct5  *vdl.Type
	vdlTypeUnion6   *vdl.Type
	vdlTypeStruct7  *vdl.Type
	vdlTypeStruct8  *vdl.Type
	vdlTypeMap9     *vdl.Type
	vdlTypeStruct10 *vdl.Type
	vdlTypeMap11    *vdl.Type
	vdlTypeMap12    *vdl.Type
)

var initializeVDLCalled bool

// initializeVDL performs vdl initialization.  It is safe to call multiple times.
// If you have an init ordering issue, just insert the following line verbatim
// into your source files in this package, right after the "package foo" clause:
//
//    var _ = initializeVDL()
//
// The purpose of this function is to ensure that vdl initialization occurs in
// the right order, and very early in the init sequence.  In particular, vdl
// registration and package variable initialization needs to occur before
// functions like vdl.TypeOf will work properly.
//
// This function returns a dummy value, so that it can be used to initialize the
// first var in the file, to take advantage of Go's defined init order.
func initializeVDL() struct{} {
	if initializeVDLCalled {
		return struct{}{}
	}
	initializeVDLCalled = true

	// Register types.
	vdl.Register((*blessingRootsState)(nil))
	vdl.Register((*dischargeCacheKey)(nil))
	vdl.Register((*CachedDischarge)(nil))
	vdl.Register((*blessingStoreState)(nil))

	// Initialize type definitions.
	vdlTypeMap1 = vdl.TypeOf((*blessingRootsState)(nil))
	vdlTypeList2 = vdl.TypeOf((*[]security.BlessingPattern)(nil))
	vdlTypeString3 = vdl.TypeOf((*security.BlessingPattern)(nil))
	vdlTypeArray4 = vdl.TypeOf((*dischargeCacheKey)(nil))
	vdlTypeStruct5 = vdl.TypeOf((*CachedDischarge)(nil)).Elem()
	vdlTypeUnion6 = vdl.TypeOf((*security.WireDischarge)(nil))
	vdlTypeStruct7 = vdl.TypeOf((*vdltime.Time)(nil)).Elem()
	vdlTypeStruct8 = vdl.TypeOf((*blessingStoreState)(nil)).Elem()
	vdlTypeMap9 = vdl.TypeOf((*map[security.BlessingPattern]security.Blessings)(nil))
	vdlTypeStruct10 = vdl.TypeOf((*security.WireBlessings)(nil)).Elem()
	vdlTypeMap11 = vdl.TypeOf((*map[dischargeCacheKey]security.Discharge)(nil))
	vdlTypeMap12 = vdl.TypeOf((*map[dischargeCacheKey]CachedDischarge)(nil))

	return struct{}{}
}
