// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.simd

package bytealg

import (
	"internal/byteorder"
	"math/bits"
	"simd/archsimd"
	"unsafe"
)

// IndexAnyASCII4Supported reports whether the experimental implementation is built.
const IndexAnyASCII4Supported = true

// IndexAnyASCII4Enabled reports whether the CPU supports the SIMD search.
var IndexAnyASCII4Enabled = archsimd.X86.AVX2()

// IndexAnyASCII4 searches for two to four ASCII bytes. The caller must ensure
// AVX2 is enabled, len(s) > 8, 2 <= len(chars) <= 4, and chars is ASCII.
func IndexAnyASCII4(s []byte, chars string) int {
	return indexAnyASCII4(unsafe.SliceData(s), len(s), chars)
}

// indexAnyASCII4 keeps the pointer-to-slice conversion in the SIMD path.
func indexAnyASCII4(data *byte, n int, chars string) int {
	s := unsafe.Slice(data, n)
	c0, c1 := chars[0], chars[1]
	c2, c3 := c1, c1
	if len(chars) > 2 {
		c2 = chars[2]
	}
	if len(chars) > 3 {
		c3 = chars[3]
	}
	// Preserve cheap first-byte matches before setting up vector registers.
	if s[0] == c0 || s[0] == c1 || s[0] == c2 || s[0] == c3 {
		return 0
	}
	v0 := archsimd.BroadcastUint8x32(c0)
	v1 := archsimd.BroadcastUint8x32(c1)
	v2 := archsimd.BroadcastUint8x32(c2)
	v3 := archsimd.BroadcastUint8x32(c3)
	index := -1
	// Short inputs use overlapping head and tail loads that stay inside s.
	// Move the tail match bits back to their positions in the original input.
	if n < 32 {
		var v archsimd.Uint8x32
		var mask uint32
		if n < 16 {
			var pair archsimd.Uint64x2
			pair = pair.SetElem(0, byteorder.LEUint64(s))
			pair = pair.SetElem(1, byteorder.LEUint64(s[n-8:]))
			v = v.SetLo(pair.ReshapeToUint8s())
			mask = indexAnyASCII4Mask(v, v0, v1, v2, v3) & 0xffff
			mask = mask&255 | (mask>>8)<<uint(n-8)
		} else {
			v = v.SetLo(archsimd.LoadUint8x16(s[:16])).SetHi(archsimd.LoadUint8x16(s[n-16:]))
			mask = indexAnyASCII4Mask(v, v0, v1, v2, v3)
			mask = mask&0xffff | (mask>>16)<<uint(n-16)
		}
		if mask != 0 {
			index = bits.TrailingZeros32(mask)
		}
	} else {
		for i := 0; ; i += 32 {
			last := i >= n-32
			if last {
				i = n - 32
			}
			v := archsimd.LoadUint8x32(s[i : i+32])
			mask := indexAnyASCII4Mask(v, v0, v1, v2, v3)
			if mask != 0 {
				index = i + bits.TrailingZeros32(mask)
				break
			}
			if last {
				break
			}
		}
	}
	archsimd.ClearAVXUpperBits()
	return index
}

// IndexAnyASCII4String is the string form of IndexAnyASCII4.
func IndexAnyASCII4String(s, chars string) int {
	return indexAnyASCII4(unsafe.StringData(s), len(s), chars)
}

func indexAnyASCII4Mask(v, c0, c1, c2, c3 archsimd.Uint8x32) uint32 {
	return v.Equal(c0).Or(v.Equal(c1)).Or(v.Equal(c2)).Or(v.Equal(c3)).ToBits()
}
