// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !amd64 || !goexperiment.simd

package bytealg

// IndexAnyASCII4Supported reports whether the experimental implementation is built.
const IndexAnyASCII4Supported = false

// IndexAnyASCII4Enabled is false when the SIMD implementation is not built.
const IndexAnyASCII4Enabled = false

// IndexAnyASCII4 is unavailable when the specialized search is not built.
func IndexAnyASCII4(s []byte, chars string) int { panic("unreachable") }

// IndexAnyASCII4String is the string form of IndexAnyASCII4.
func IndexAnyASCII4String(s, chars string) int { panic("unreachable") }
