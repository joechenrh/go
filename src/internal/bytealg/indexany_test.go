// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytealg_test

import (
	"bytes"
	"strings"
	"testing"
	"unsafe"
)

func TestIndexAnySmallASCII(t *testing.T) {
	reference := func(s, chars string) int {
		for i, r := range s {
			for _, c := range chars {
				if r == c {
					return i
				}
			}
		}
		return -1
	}
	for _, chars := range []string{"", "a", "ab", "abc", "abcd", "aaaa", "abcde", "\x00\x7f", "é", "éx", "xé", "\xffx", "世", "�"} {
		for _, n := range []int{0, 1, 2, 7, 8, 9, 15, 16, 17, 31, 32, 33, 63, 64, 65, 127, 128, 129, 255, 256, 257} {
			for align := 0; align < 32; align++ {
				buf := make([]byte, n+31)
				s := buf[align : align+n]
				for i := range s {
					s[i] = '#'
				}
				for _, pos := range []int{-1, 0, 1, 15, 16, 31, 32, n - 1} {
					if pos >= n {
						continue
					}
					for _, value := range []byte{0, 'a', 'b', 'c', 'd', 127, 128, 195, 255} {
						if pos >= 0 {
							s[pos] = value
						}
						text := unsafe.String(unsafe.SliceData(s), len(s))
						want := reference(text, chars)
						if got := bytes.IndexAny(s, chars); got != want {
							t.Fatalf("bytes chars=%q n=%d align=%d pos=%d value=%x got=%d want=%d", chars, n, align, pos, value, got, want)
						}
						if got := strings.IndexAny(text, chars); got != want {
							t.Fatalf("strings chars=%q n=%d align=%d pos=%d value=%x got=%d want=%d", chars, n, align, pos, value, got, want)
						}
						if pos >= 0 {
							s[pos] = '#'
						}
					}
				}
			}
		}
	}
}

func FuzzIndexAnySmallASCII(f *testing.F) {
	for _, chars := range []string{"ab", "abc", "abcd", "aa", "\x00?", "éx", "\xffx"} {
		for _, n := range []int{0, 9, 15, 16, 31, 32, 33, 65} {
			f.Add(strings.Repeat("#", n)+chars, chars)
		}
	}
	f.Fuzz(func(t *testing.T, s, chars string) {
		if len(s) > 4096 || len(chars) > 64 {
			t.Skip()
		}
		want := -1
	outer:
		for i, r := range s {
			for _, c := range chars {
				if r == c {
					want = i
					break outer
				}
			}
		}
		if got := strings.IndexAny(s, chars); got != want {
			t.Fatalf("strings got=%d want=%d", got, want)
		}
		if got := bytes.IndexAny([]byte(s), chars); got != want {
			t.Fatalf("bytes got=%d want=%d", got, want)
		}
	})
}
