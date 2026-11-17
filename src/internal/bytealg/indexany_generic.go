// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !amd64 || !goexperiment.simd

package bytealg

// HaveSIMD is false when the SIMD IndexAny implementation is not built.
const HaveSIMD = false

// IndexAnyASCII4 is unavailable when the specialized search is not built.
func IndexAnyASCII4(s []byte, chars string) int { panic("unreachable") }

// IndexAnyASCII4String is the string form of IndexAnyASCII4.
func IndexAnyASCII4String(s, chars string) int { panic("unreachable") }
