// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build amd64 && goexperiment.simd

package runtime_test

import (
	"runtime"
	"simd/archsimd"
	"sync/atomic"
	"testing"
)

//go:noinline
func spinPreemptSIMD256(stop, entered *atomic.Uint32, input *[12][4]uint64, step *[4]uint64) (result [12][4]uint64, iterations uint64) {
	v0 := archsimd.LoadUint64x4Array(&input[0])
	v1 := archsimd.LoadUint64x4Array(&input[1])
	v2 := archsimd.LoadUint64x4Array(&input[2])
	v3 := archsimd.LoadUint64x4Array(&input[3])
	v4 := archsimd.LoadUint64x4Array(&input[4])
	v5 := archsimd.LoadUint64x4Array(&input[5])
	v6 := archsimd.LoadUint64x4Array(&input[6])
	v7 := archsimd.LoadUint64x4Array(&input[7])
	v8 := archsimd.LoadUint64x4Array(&input[8])
	v9 := archsimd.LoadUint64x4Array(&input[9])
	v10 := archsimd.LoadUint64x4Array(&input[10])
	v11 := archsimd.LoadUint64x4Array(&input[11])
	delta := archsimd.LoadUint64x4Array(step)
	entered.Store(1)
	for stop.Load() == 0 {
		v0 = v0.Add(delta)
		v1 = v1.Add(delta)
		v2 = v2.Add(delta)
		v3 = v3.Add(delta)
		v4 = v4.Add(delta)
		v5 = v5.Add(delta)
		v6 = v6.Add(delta)
		v7 = v7.Add(delta)
		v8 = v8.Add(delta)
		v9 = v9.Add(delta)
		v10 = v10.Add(delta)
		v11 = v11.Add(delta)
		iterations++
	}
	v0.StoreArray(&result[0])
	v1.StoreArray(&result[1])
	v2.StoreArray(&result[2])
	v3.StoreArray(&result[3])
	v4.StoreArray(&result[4])
	v5.StoreArray(&result[5])
	v6.StoreArray(&result[6])
	v7.StoreArray(&result[7])
	v8.StoreArray(&result[8])
	v9.StoreArray(&result[9])
	v10.StoreArray(&result[10])
	v11.StoreArray(&result[11])
	archsimd.ClearAVXUpperBits()
	return
}
func TestAsyncPreemptSIMD256(t *testing.T) {
	if !runtime.AsyncPreemptEnabled() {
		t.Skip("asynchronous preemption is unavailable or disabled")
	}
	if !archsimd.X86.AVX2() {
		t.Skip("feature unavailable")
	}
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(1))
	for _, upperLive := range []bool{false, true} {
		var stop, entered atomic.Uint32
		var input [12][4]uint64
		var step [4]uint64
		for i := range input {
			for j := range input[i] {
				if j < 2 || upperLive {
					input[i][j] = uint64(i*100 + j + 1)
					step[j] = uint64(j + 1)
				}
			}
		}
		done := make(chan bool, 1)
		go func() {
			got, n := spinPreemptSIMD256(&stop, &entered, &input, &step)
			ok := n > 0
			for i := range got {
				for j := range got[i] {
					if got[i][j] != input[i][j]+n*step[j] {
						ok = false
					}
				}
			}
			done <- ok
		}()
		for entered.Load() == 0 {
			runtime.Gosched()
		}
		// The worker has no calls in its loop. With one P, reaching here requires
		// asynchronous preemption while its vector values are live.
		for i := 0; i < 4; i++ {
			runtime.GC()
			runtime.Gosched()
		}
		stop.Store(1)
		if !<-done {
			t.Fatalf("upperLive=%v: live vector state corrupted", upperLive)
		}
	}
}

//go:noinline
func spinPreemptSIMD512(stop, entered *atomic.Uint32, input *[28][8]uint64, step *[8]uint64) (result [28][8]uint64, iterations uint64) {
	v0 := archsimd.LoadUint64x8Array(&input[0])
	v1 := archsimd.LoadUint64x8Array(&input[1])
	v2 := archsimd.LoadUint64x8Array(&input[2])
	v3 := archsimd.LoadUint64x8Array(&input[3])
	v4 := archsimd.LoadUint64x8Array(&input[4])
	v5 := archsimd.LoadUint64x8Array(&input[5])
	v6 := archsimd.LoadUint64x8Array(&input[6])
	v7 := archsimd.LoadUint64x8Array(&input[7])
	v8 := archsimd.LoadUint64x8Array(&input[8])
	v9 := archsimd.LoadUint64x8Array(&input[9])
	v10 := archsimd.LoadUint64x8Array(&input[10])
	v11 := archsimd.LoadUint64x8Array(&input[11])
	v12 := archsimd.LoadUint64x8Array(&input[12])
	v13 := archsimd.LoadUint64x8Array(&input[13])
	v14 := archsimd.LoadUint64x8Array(&input[14])
	v15 := archsimd.LoadUint64x8Array(&input[15])
	v16 := archsimd.LoadUint64x8Array(&input[16])
	v17 := archsimd.LoadUint64x8Array(&input[17])
	v18 := archsimd.LoadUint64x8Array(&input[18])
	v19 := archsimd.LoadUint64x8Array(&input[19])
	v20 := archsimd.LoadUint64x8Array(&input[20])
	v21 := archsimd.LoadUint64x8Array(&input[21])
	v22 := archsimd.LoadUint64x8Array(&input[22])
	v23 := archsimd.LoadUint64x8Array(&input[23])
	v24 := archsimd.LoadUint64x8Array(&input[24])
	v25 := archsimd.LoadUint64x8Array(&input[25])
	v26 := archsimd.LoadUint64x8Array(&input[26])
	v27 := archsimd.LoadUint64x8Array(&input[27])
	delta := archsimd.LoadUint64x8Array(step)
	entered.Store(1)
	for stop.Load() == 0 {
		v0 = v0.Add(delta)
		v1 = v1.Add(delta)
		v2 = v2.Add(delta)
		v3 = v3.Add(delta)
		v4 = v4.Add(delta)
		v5 = v5.Add(delta)
		v6 = v6.Add(delta)
		v7 = v7.Add(delta)
		v8 = v8.Add(delta)
		v9 = v9.Add(delta)
		v10 = v10.Add(delta)
		v11 = v11.Add(delta)
		v12 = v12.Add(delta)
		v13 = v13.Add(delta)
		v14 = v14.Add(delta)
		v15 = v15.Add(delta)
		v16 = v16.Add(delta)
		v17 = v17.Add(delta)
		v18 = v18.Add(delta)
		v19 = v19.Add(delta)
		v20 = v20.Add(delta)
		v21 = v21.Add(delta)
		v22 = v22.Add(delta)
		v23 = v23.Add(delta)
		v24 = v24.Add(delta)
		v25 = v25.Add(delta)
		v26 = v26.Add(delta)
		v27 = v27.Add(delta)
		iterations++
	}
	v0.StoreArray(&result[0])
	v1.StoreArray(&result[1])
	v2.StoreArray(&result[2])
	v3.StoreArray(&result[3])
	v4.StoreArray(&result[4])
	v5.StoreArray(&result[5])
	v6.StoreArray(&result[6])
	v7.StoreArray(&result[7])
	v8.StoreArray(&result[8])
	v9.StoreArray(&result[9])
	v10.StoreArray(&result[10])
	v11.StoreArray(&result[11])
	v12.StoreArray(&result[12])
	v13.StoreArray(&result[13])
	v14.StoreArray(&result[14])
	v15.StoreArray(&result[15])
	v16.StoreArray(&result[16])
	v17.StoreArray(&result[17])
	v18.StoreArray(&result[18])
	v19.StoreArray(&result[19])
	v20.StoreArray(&result[20])
	v21.StoreArray(&result[21])
	v22.StoreArray(&result[22])
	v23.StoreArray(&result[23])
	v24.StoreArray(&result[24])
	v25.StoreArray(&result[25])
	v26.StoreArray(&result[26])
	v27.StoreArray(&result[27])
	archsimd.ClearAVXUpperBits()
	return
}
func TestAsyncPreemptSIMD512(t *testing.T) {
	if !runtime.AsyncPreemptEnabled() {
		t.Skip("asynchronous preemption is unavailable or disabled")
	}
	if !archsimd.X86.AVX512() {
		t.Skip("feature unavailable")
	}
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(1))
	for _, upperLive := range []bool{false, true} {
		var stop, entered atomic.Uint32
		var input [28][8]uint64
		var step [8]uint64
		for i := range input {
			for j := range input[i] {
				if j < 2 || upperLive {
					input[i][j] = uint64(i*100 + j + 1)
					step[j] = uint64(j + 1)
				}
			}
		}
		done := make(chan bool, 1)
		go func() {
			got, n := spinPreemptSIMD512(&stop, &entered, &input, &step)
			ok := n > 0
			for i := range got {
				for j := range got[i] {
					if got[i][j] != input[i][j]+n*step[j] {
						ok = false
					}
				}
			}
			done <- ok
		}()
		for entered.Load() == 0 {
			runtime.Gosched()
		}
		// The worker has no calls in its loop. With one P, reaching here requires
		// asynchronous preemption while its vector values are live.
		for i := 0; i < 4; i++ {
			runtime.GC()
			runtime.Gosched()
		}
		stop.Store(1)
		if !<-done {
			t.Fatalf("upperLive=%v: live vector state corrupted", upperLive)
		}
	}
}

//go:noinline
func spinPreemptSIMDFloat256(stop, entered *atomic.Uint32, input *[12][4]float64, step *[4]float64) (result [12][4]float64, iterations uint64) {
	v0 := archsimd.LoadFloat64x4Array(&input[0])
	v1 := archsimd.LoadFloat64x4Array(&input[1])
	v2 := archsimd.LoadFloat64x4Array(&input[2])
	v3 := archsimd.LoadFloat64x4Array(&input[3])
	v4 := archsimd.LoadFloat64x4Array(&input[4])
	v5 := archsimd.LoadFloat64x4Array(&input[5])
	v6 := archsimd.LoadFloat64x4Array(&input[6])
	v7 := archsimd.LoadFloat64x4Array(&input[7])
	v8 := archsimd.LoadFloat64x4Array(&input[8])
	v9 := archsimd.LoadFloat64x4Array(&input[9])
	v10 := archsimd.LoadFloat64x4Array(&input[10])
	v11 := archsimd.LoadFloat64x4Array(&input[11])
	delta := archsimd.LoadFloat64x4Array(step)
	entered.Store(1)
	for stop.Load() == 0 {
		v0 = v0.Add(delta)
		v1 = v1.Add(delta)
		v2 = v2.Add(delta)
		v3 = v3.Add(delta)
		v4 = v4.Add(delta)
		v5 = v5.Add(delta)
		v6 = v6.Add(delta)
		v7 = v7.Add(delta)
		v8 = v8.Add(delta)
		v9 = v9.Add(delta)
		v10 = v10.Add(delta)
		v11 = v11.Add(delta)
		iterations++
	}
	v0.StoreArray(&result[0])
	v1.StoreArray(&result[1])
	v2.StoreArray(&result[2])
	v3.StoreArray(&result[3])
	v4.StoreArray(&result[4])
	v5.StoreArray(&result[5])
	v6.StoreArray(&result[6])
	v7.StoreArray(&result[7])
	v8.StoreArray(&result[8])
	v9.StoreArray(&result[9])
	v10.StoreArray(&result[10])
	v11.StoreArray(&result[11])
	archsimd.ClearAVXUpperBits()
	return
}
func TestAsyncPreemptSIMDFloat256(t *testing.T) {
	if !runtime.AsyncPreemptEnabled() {
		t.Skip("asynchronous preemption is unavailable or disabled")
	}
	if !archsimd.X86.AVX() {
		t.Skip("feature unavailable")
	}
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(1))
	for _, upperLive := range []bool{false, true} {
		var stop, entered atomic.Uint32
		var input [12][4]float64
		var step [4]float64
		for i := range input {
			for j := range input[i] {
				if j < 2 || upperLive {
					input[i][j] = float64(i*100 + j + 1)
					step[j] = float64(j + 1)
				}
			}
		}
		done := make(chan bool, 1)
		go func() {
			got, n := spinPreemptSIMDFloat256(&stop, &entered, &input, &step)
			ok := n > 0
			for i := range got {
				for j := range got[i] {
					if got[i][j] != input[i][j]+float64(n)*step[j] {
						ok = false
					}
				}
			}
			done <- ok
		}()
		for entered.Load() == 0 {
			runtime.Gosched()
		}
		// The worker has no calls in its loop. With one P, reaching here requires
		// asynchronous preemption while its vector values are live.
		for i := 0; i < 4; i++ {
			runtime.GC()
			runtime.Gosched()
		}
		stop.Store(1)
		if !<-done {
			t.Fatalf("upperLive=%v: live vector state corrupted", upperLive)
		}
	}
}
