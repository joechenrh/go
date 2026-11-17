// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytealg

// maxSIMDChars matches the four comparisons in indexAnyASCII4.
const maxSIMDChars = 4

// IsASCII4 reports whether chars consists of two to four ASCII bytes.
func IsASCII4(chars string) bool {
	if len(chars) < 2 || len(chars) > maxSIMDChars {
		return false
	}
	for i := 0; i < len(chars); i++ {
		if chars[i] >= 128 {
			return false
		}
	}
	return true
}
