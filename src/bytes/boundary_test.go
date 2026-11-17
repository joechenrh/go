// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
//go:build linux

package bytes_test

import (
	. "bytes"
	"strings"
	"syscall"
	"testing"
	"unsafe"
)

// This file tests the situation where byte operations are checking
// data very near to a page boundary. We want to make sure those
// operations do not read across the boundary and cause a page
// fault where they shouldn't.

// These tests run only on linux. The code being tested is
// not OS-specific, so it does not need to be tested on all
// operating systems.

// dangerousSlice returns a slice which is immediately
// preceded and followed by a faulting page.
func dangerousSlice(t *testing.T) []byte {
	pagesize := syscall.Getpagesize()
	b, err := syscall.Mmap(0, 0, 3*pagesize, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_ANONYMOUS|syscall.MAP_PRIVATE)
	if err != nil {
		t.Fatalf("mmap failed %s", err)
	}
	err = syscall.Mprotect(b[:pagesize], syscall.PROT_NONE)
	if err != nil {
		t.Fatalf("mprotect low failed %s\n", err)
	}
	err = syscall.Mprotect(b[2*pagesize:], syscall.PROT_NONE)
	if err != nil {
		t.Fatalf("mprotect high failed %s\n", err)
	}
	return b[pagesize : 2*pagesize]
}

func TestEqualNearPageBoundary(t *testing.T) {
	t.Parallel()
	b := dangerousSlice(t)
	for i := range b {
		b[i] = 'A'
	}
	for i := 0; i <= len(b); i++ {
		Equal(b[:i], b[len(b)-i:])
		Equal(b[len(b)-i:], b[:i])
	}
}

func TestIndexByteNearPageBoundary(t *testing.T) {
	t.Parallel()
	b := dangerousSlice(t)
	for i := range b {
		idx := IndexByte(b[i:], 1)
		if idx != -1 {
			t.Fatalf("IndexByte(b[%d:])=%d, want -1\n", i, idx)
		}
	}
}

func TestLastIndexByteNearPageBoundary(t *testing.T) {
	t.Parallel()
	b := dangerousSlice(t)
	for i := range b {
		idx := LastIndexByte(b[i:], 1)
		if idx != -1 {
			t.Fatalf("LastIndexByte(b[%d:])=%d, want -1\n", i, idx)
		}
		idx = LastIndexByte(b[:i], 1)
		if idx != -1 {
			t.Fatalf("LastIndexByte(b[:%d])=%d, want -1\n", i, idx)
		}
	}
}

func TestIndexNearPageBoundary(t *testing.T) {
	t.Parallel()
	q := dangerousSlice(t)
	if len(q) > 64 {
		// Only worry about when we're near the end of a page.
		q = q[len(q)-64:]
	}
	b := dangerousSlice(t)
	if len(b) > 256 {
		// Only worry about when we're near the end of a page.
		b = b[len(b)-256:]
	}
	for j := 1; j < len(q); j++ {
		q[j-1] = 1 // difference is only found on the last byte
		for i := range b {
			idx := Index(b[i:], q[:j])
			if idx != -1 {
				t.Fatalf("Index(b[%d:], q[:%d])=%d, want -1\n", i, j, idx)
			}
		}
		q[j-1] = 0
	}

	// Test differing alignments and sizes of q which always end on a page boundary.
	q[len(q)-1] = 1 // difference is only found on the last byte
	for j := 0; j < len(q); j++ {
		for i := range b {
			idx := Index(b[i:], q[j:])
			if idx != -1 {
				t.Fatalf("Index(b[%d:], q[%d:])=%d, want -1\n", i, j, idx)
			}
		}
	}
	q[len(q)-1] = 0
}

func TestCountNearPageBoundary(t *testing.T) {
	t.Parallel()
	b := dangerousSlice(t)
	for i := range b {
		c := Count(b[i:], []byte{1})
		if c != 0 {
			t.Fatalf("Count(b[%d:], {1})=%d, want 0\n", i, c)
		}
		c = Count(b[:i], []byte{0})
		if c != i {
			t.Fatalf("Count(b[:%d], {0})=%d, want %d\n", i, c, i)
		}
	}
}

// TestIndexAnyNearPageBoundary checks both APIs without copying mapped inputs
// or delimiter sets, at the low and high guard pages.
func TestIndexAnyNearPageBoundary(t *testing.T) {
	t.Parallel()
	page := dangerousSlice(t)
	charsPage := dangerousSlice(t)
	for i := range page {
		page[i] = '#'
	}
	for i := range charsPage {
		charsPage[i] = '#'
	}
	sizes := make([]int, 257)
	for i := range sizes {
		sizes[i] = i
	}
	sizes = append(sizes, len(page)-1, len(page))
	check := func(input []byte, chars string, want int) {
		t.Helper()
		if got := IndexAny(input, chars); got != want {
			t.Fatalf("bytes n=%d chars=%q got=%d want=%d", len(input), chars, got, want)
		}
		view := unsafe.String(unsafe.SliceData(input), len(input))
		if got := strings.IndexAny(view, chars); got != want {
			t.Fatalf("strings n=%d chars=%q got=%d want=%d", len(input), chars, got, want)
		}
	}
	runSet := func(chars string) {
		for _, n := range sizes {
			for _, atEnd := range []bool{false, true} {
				input := page[:n]
				if atEnd {
					input = page[len(page)-n:]
				}
				for _, pos := range []int{-1, 0, n / 2, n - 1} {
					if pos >= n {
						continue
					}
					if pos >= 0 {
						input[pos] = chars[0]
					}
					check(input, chars, pos)
					if pos >= 0 {
						input[pos] = '#'
					}
				}
			}
		}
	}
	for _, pattern := range []string{"\r\n", "\t\r\n", ",\"\n\\", ",\"\n\r\\", "\x00?", "0123456789abcdef", "x\xff", "xé", "x世?", "\x00\xff"} {
		for _, atEnd := range []bool{false, true} {
			rawChars := charsPage[:len(pattern)]
			if atEnd {
				rawChars = charsPage[len(charsPage)-len(pattern):]
			}
			copy(rawChars, pattern)
			runSet(unsafe.String(unsafe.SliceData(rawChars), len(rawChars)))
		}
	}

	// A string view must not let the byte-oriented kernels modify either page.
	for i := range page {
		page[i] = '#'
	}
	page[len(page)-1] = '\n'
	copy(charsPage[len(charsPage)-2:], "\r\n")
	if err := syscall.Mprotect(page, syscall.PROT_READ); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mprotect(charsPage, syscall.PROT_READ); err != nil {
		t.Fatal(err)
	}
	chars := unsafe.String(unsafe.SliceData(charsPage[len(charsPage)-2:]), 2)
	for _, n := range sizes {
		check(page[len(page)-n:], chars, n-1)
		want := -1
		if n == len(page) {
			want = n - 1
		}
		check(page[:n], chars, want)
	}
}

func TestIndexAnyUnicodeReadOnly(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ chars, suffix string }{
		{"x\xff", "\xff"},
		{"éx", "é"},
		{"x世?", "世"},
		{"\x00\xff", "\x00"},
	} {
		input := dangerousSlice(t)
		delimiters := dangerousSlice(t)
		for i := range input {
			input[i] = '#'
		}
		copy(input[len(input)-len(tt.suffix):], tt.suffix)
		copy(delimiters[len(delimiters)-len(tt.chars):], tt.chars)
		if err := syscall.Mprotect(input, syscall.PROT_READ); err != nil {
			t.Fatal(err)
		}
		if err := syscall.Mprotect(delimiters, syscall.PROT_READ); err != nil {
			t.Fatal(err)
		}
		chars := unsafe.String(&delimiters[len(delimiters)-len(tt.chars)], len(tt.chars))
		sizes := make([]int, 258)
		for i := range sizes {
			sizes[i] = i
		}
		sizes = append(sizes, len(input)-1, len(input))
		for _, n := range sizes {
			for _, data := range [][]byte{input[:n], input[len(input)-n:]} {
				view := unsafe.String(unsafe.SliceData(data), len(data))
				want := -1
			outer:
				for i, r := range view {
					for _, c := range chars {
						if r == c {
							want = i
							break outer
						}
					}
				}
				if got := IndexAny(data, chars); got != want {
					t.Fatalf("bytes n=%d chars=%q: got %d, want %d", n, chars, got, want)
				}
				if got := strings.IndexAny(view, chars); got != want {
					t.Fatalf("strings n=%d chars=%q: got %d, want %d", n, chars, got, want)
				}
			}
		}
	}
}
