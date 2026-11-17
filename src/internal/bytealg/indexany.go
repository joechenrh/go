// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytealg

// IsASCII4 reports whether a short character set contains only ASCII bytes.
// Callers limit chars to at most four bytes to keep this check inexpensive.
func IsASCII4(chars string) bool {
	for i := 0; i < len(chars); i++ {
		if chars[i] >= 128 {
			return false
		}
	}
	return true
}
