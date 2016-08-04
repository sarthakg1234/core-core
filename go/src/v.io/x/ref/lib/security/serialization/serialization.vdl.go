// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file was auto-generated by the vanadium vdl tool.
// Package: serialization

package serialization

import (
	"fmt"
	"v.io/v23/security"
	"v.io/v23/vdl"
)

var _ = __VDLInit() // Must be first; see __VDLInit comments for details.

//////////////////////////////////////////////////
// Type definitions

type SignedHeader struct {
	ChunkSizeBytes int64
}

func (SignedHeader) __VDLReflect(struct {
	Name string `vdl:"v.io/x/ref/lib/security/serialization.SignedHeader"`
}) {
}

func (x SignedHeader) VDLIsZero() bool {
	return x == SignedHeader{}
}

func (x SignedHeader) VDLWrite(enc vdl.Encoder) error {
	if err := enc.StartValue(__VDLType_struct_1); err != nil {
		return err
	}
	if x.ChunkSizeBytes != 0 {
		if err := enc.NextFieldValueInt(0, vdl.Int64Type, x.ChunkSizeBytes); err != nil {
			return err
		}
	}
	if err := enc.NextField(-1); err != nil {
		return err
	}
	return enc.FinishValue()
}

func (x *SignedHeader) VDLRead(dec vdl.Decoder) error {
	*x = SignedHeader{}
	if err := dec.StartValue(__VDLType_struct_1); err != nil {
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
		if decType != __VDLType_struct_1 {
			index = __VDLType_struct_1.FieldIndexByName(decType.Field(index).Name)
			if index == -1 {
				if err := dec.SkipValue(); err != nil {
					return err
				}
				continue
			}
		}
		switch index {
		case 0:
			switch value, err := dec.ReadValueInt(64); {
			case err != nil:
				return err
			default:
				x.ChunkSizeBytes = value
			}
		}
	}
}

type HashCode [32]byte

func (HashCode) __VDLReflect(struct {
	Name string `vdl:"v.io/x/ref/lib/security/serialization.HashCode"`
}) {
}

func (x HashCode) VDLIsZero() bool {
	return x == HashCode{}
}

func (x HashCode) VDLWrite(enc vdl.Encoder) error {
	if err := enc.WriteValueBytes(__VDLType_array_2, x[:]); err != nil {
		return err
	}
	return nil
}

func (x *HashCode) VDLRead(dec vdl.Decoder) error {
	bytes := x[:]
	if err := dec.ReadValueBytes(32, &bytes); err != nil {
		return err
	}
	return nil
}

type (
	// SignedData represents any single field of the SignedData union type.
	//
	// SignedData describes the information sent by a SigningWriter and read by VerifiyingReader.
	SignedData interface {
		// Index returns the field index.
		Index() int
		// Interface returns the field value as an interface.
		Interface() interface{}
		// Name returns the field name.
		Name() string
		// __VDLReflect describes the SignedData union type.
		__VDLReflect(__SignedDataReflect)
		VDLIsZero() bool
		VDLWrite(vdl.Encoder) error
	}
	// SignedDataSignature represents field Signature of the SignedData union type.
	SignedDataSignature struct{ Value security.Signature }
	// SignedDataHash represents field Hash of the SignedData union type.
	SignedDataHash struct{ Value HashCode }
	// __SignedDataReflect describes the SignedData union type.
	__SignedDataReflect struct {
		Name  string `vdl:"v.io/x/ref/lib/security/serialization.SignedData"`
		Type  SignedData
		Union struct {
			Signature SignedDataSignature
			Hash      SignedDataHash
		}
	}
)

func (x SignedDataSignature) Index() int                       { return 0 }
func (x SignedDataSignature) Interface() interface{}           { return x.Value }
func (x SignedDataSignature) Name() string                     { return "Signature" }
func (x SignedDataSignature) __VDLReflect(__SignedDataReflect) {}

func (x SignedDataHash) Index() int                       { return 1 }
func (x SignedDataHash) Interface() interface{}           { return x.Value }
func (x SignedDataHash) Name() string                     { return "Hash" }
func (x SignedDataHash) __VDLReflect(__SignedDataReflect) {}

func (x SignedDataSignature) VDLIsZero() bool {
	return x.Value.VDLIsZero()
}

func (x SignedDataHash) VDLIsZero() bool {
	return false
}

func (x SignedDataSignature) VDLWrite(enc vdl.Encoder) error {
	if err := enc.StartValue(__VDLType_union_4); err != nil {
		return err
	}
	if err := enc.NextField(0); err != nil {
		return err
	}
	if err := x.Value.VDLWrite(enc); err != nil {
		return err
	}
	if err := enc.NextField(-1); err != nil {
		return err
	}
	return enc.FinishValue()
}

func (x SignedDataHash) VDLWrite(enc vdl.Encoder) error {
	if err := enc.StartValue(__VDLType_union_4); err != nil {
		return err
	}
	if err := enc.NextFieldValueBytes(1, __VDLType_array_2, x.Value[:]); err != nil {
		return err
	}
	if err := enc.NextField(-1); err != nil {
		return err
	}
	return enc.FinishValue()
}

func VDLReadSignedData(dec vdl.Decoder, x *SignedData) error {
	if err := dec.StartValue(__VDLType_union_4); err != nil {
		return err
	}
	decType := dec.Type()
	index, err := dec.NextField()
	switch {
	case err != nil:
		return err
	case index == -1:
		return fmt.Errorf("missing field in union %T, from %v", x, decType)
	}
	if decType != __VDLType_union_4 {
		name := decType.Field(index).Name
		index = __VDLType_union_4.FieldIndexByName(name)
		if index == -1 {
			return fmt.Errorf("field %q not in union %T, from %v", name, x, decType)
		}
	}
	switch index {
	case 0:
		var field SignedDataSignature
		if err := field.Value.VDLRead(dec); err != nil {
			return err
		}
		*x = field
	case 1:
		var field SignedDataHash
		bytes := field.Value[:]
		if err := dec.ReadValueBytes(32, &bytes); err != nil {
			return err
		}
		*x = field
	}
	switch index, err := dec.NextField(); {
	case err != nil:
		return err
	case index != -1:
		return fmt.Errorf("extra field %d in union %T, from %v", index, x, dec.Type())
	}
	return dec.FinishValue()
}

// Hold type definitions in package-level variables, for better performance.
var (
	__VDLType_struct_1 *vdl.Type
	__VDLType_array_2  *vdl.Type
	__VDLType_struct_3 *vdl.Type
	__VDLType_union_4  *vdl.Type
)

var __VDLInitCalled bool

// __VDLInit performs vdl initialization.  It is safe to call multiple times.
// If you have an init ordering issue, just insert the following line verbatim
// into your source files in this package, right after the "package foo" clause:
//
//    var _ = __VDLInit()
//
// The purpose of this function is to ensure that vdl initialization occurs in
// the right order, and very early in the init sequence.  In particular, vdl
// registration and package variable initialization needs to occur before
// functions like vdl.TypeOf will work properly.
//
// This function returns a dummy value, so that it can be used to initialize the
// first var in the file, to take advantage of Go's defined init order.
func __VDLInit() struct{} {
	if __VDLInitCalled {
		return struct{}{}
	}
	__VDLInitCalled = true

	// Register types.
	vdl.Register((*SignedHeader)(nil))
	vdl.Register((*HashCode)(nil))
	vdl.Register((*SignedData)(nil))

	// Initialize type definitions.
	__VDLType_struct_1 = vdl.TypeOf((*SignedHeader)(nil)).Elem()
	__VDLType_array_2 = vdl.TypeOf((*HashCode)(nil))
	__VDLType_struct_3 = vdl.TypeOf((*security.Signature)(nil)).Elem()
	__VDLType_union_4 = vdl.TypeOf((*SignedData)(nil))

	return struct{}{}
}