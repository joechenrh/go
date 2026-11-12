// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package x86

import (
	"cmd/internal/obj"
	"cmd/internal/objabi"
	"fmt"
	"internal/testenv"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/arch/x86/x86asm"
)

type oclassTest struct {
	arg  *obj.Addr
	want int // Expected oclass return value for a given arg
}

// Filled inside init, because it's easier to do with helper functions.
var (
	oclassTestsAMD64 []*oclassTest
	oclassTests386   []*oclassTest
)

func init() {
	// Required for tests that access any of
	// opindex/ycover/reg/regrex global tables.
	var ctxt obj.Link
	instinit(&ctxt)

	regAddr := func(reg int16) *obj.Addr {
		return &obj.Addr{Type: obj.TYPE_REG, Reg: reg}
	}
	immAddr := func(v int64) *obj.Addr {
		return &obj.Addr{Type: obj.TYPE_CONST, Offset: v}
	}
	regListAddr := func(regFrom, regTo int16) *obj.Addr {
		return &obj.Addr{Type: obj.TYPE_REGLIST, Offset: EncodeRegisterRange(regFrom, regTo)}
	}
	memAddr := func(base, index int16) *obj.Addr {
		return &obj.Addr{Type: obj.TYPE_MEM, Reg: base, Index: index}
	}

	// TODO(quasilyte): oclass doesn't return Yxxx for X/Y regs with
	// ID higher than 7. We don't encode such instructions, but this
	// behavior seems inconsistent. It should probably either
	// never check for arch or do it in all cases.

	oclassTestsCommon := []*oclassTest{
		{&obj.Addr{Type: obj.TYPE_NONE}, Ynone},
		{&obj.Addr{Type: obj.TYPE_BRANCH}, Ybr},
		{&obj.Addr{Type: obj.TYPE_TEXTSIZE}, Ytextsize},

		{&obj.Addr{Type: obj.TYPE_INDIR, Name: obj.NAME_EXTERN}, Yindir},
		{&obj.Addr{Type: obj.TYPE_INDIR, Name: obj.NAME_GOTREF}, Yindir},

		{&obj.Addr{Type: obj.TYPE_ADDR, Name: obj.NAME_AUTO}, Yiauto},
		{&obj.Addr{Type: obj.TYPE_ADDR, Name: obj.NAME_PARAM}, Yiauto},
		{&obj.Addr{Type: obj.TYPE_ADDR, Name: obj.NAME_EXTERN}, Yiauto},
		{&obj.Addr{Type: obj.TYPE_ADDR, Sym: &obj.LSym{Name: "runtime.duff"}}, Yi32},
		{&obj.Addr{Type: obj.TYPE_ADDR, Offset: 4}, Yu7},
		{&obj.Addr{Type: obj.TYPE_ADDR, Offset: 255}, Yu8},

		{immAddr(0), Yi0},
		{immAddr(1), Yi1},
		{immAddr(2), Yu2},
		{immAddr(3), Yu2},
		{immAddr(4), Yu7},
		{immAddr(86), Yu7},
		{immAddr(127), Yu7},
		{immAddr(128), Yu8},
		{immAddr(200), Yu8},
		{immAddr(255), Yu8},
		{immAddr(-1), Yi8},
		{immAddr(-100), Yi8},
		{immAddr(-128), Yi8},

		{regAddr(REG_AL), Yal},
		{regAddr(REG_AX), Yax},
		{regAddr(REG_DL), Yrb},
		{regAddr(REG_DH), Yrb},
		{regAddr(REG_BH), Yrb},
		{regAddr(REG_CL), Ycl},
		{regAddr(REG_CX), Ycx},
		{regAddr(REG_DX), Yrx},
		{regAddr(REG_BX), Yrx},
		{regAddr(REG_F0), Yf0},
		{regAddr(REG_F3), Yrf},
		{regAddr(REG_F7), Yrf},
		{regAddr(REG_M0), Ymr},
		{regAddr(REG_M3), Ymr},
		{regAddr(REG_M7), Ymr},
		{regAddr(REG_X0), Yxr0},
		{regAddr(REG_X6), Yxr},
		{regAddr(REG_X13), Yxr},
		{regAddr(REG_X20), YxrEvex},
		{regAddr(REG_X31), YxrEvex},
		{regAddr(REG_Y0), Yyr},
		{regAddr(REG_Y6), Yyr},
		{regAddr(REG_Y13), Yyr},
		{regAddr(REG_Y20), YyrEvex},
		{regAddr(REG_Y31), YyrEvex},
		{regAddr(REG_Z0), Yzr},
		{regAddr(REG_Z6), Yzr},
		{regAddr(REG_K0), Yk0},
		{regAddr(REG_K5), Yknot0},
		{regAddr(REG_K7), Yknot0},
		{regAddr(REG_CS), Ycs},
		{regAddr(REG_SS), Yss},
		{regAddr(REG_DS), Yds},
		{regAddr(REG_ES), Yes},
		{regAddr(REG_FS), Yfs},
		{regAddr(REG_GS), Ygs},
		{regAddr(REG_TLS), Ytls},
		{regAddr(REG_GDTR), Ygdtr},
		{regAddr(REG_IDTR), Yidtr},
		{regAddr(REG_LDTR), Yldtr},
		{regAddr(REG_MSW), Ymsw},
		{regAddr(REG_TASK), Ytask},
		{regAddr(REG_CR0), Ycr0},
		{regAddr(REG_CR5), Ycr5},
		{regAddr(REG_CR8), Ycr8},
		{regAddr(REG_DR0), Ydr0},
		{regAddr(REG_DR5), Ydr5},
		{regAddr(REG_DR7), Ydr7},
		{regAddr(REG_TR0), Ytr0},
		{regAddr(REG_TR5), Ytr5},
		{regAddr(REG_TR7), Ytr7},

		{regListAddr(REG_X0, REG_X3), YxrEvexMulti4},
		{regListAddr(REG_X4, REG_X7), YxrEvexMulti4},
		{regListAddr(REG_Y0, REG_Y3), YyrEvexMulti4},
		{regListAddr(REG_Y4, REG_Y7), YyrEvexMulti4},
		{regListAddr(REG_Z0, REG_Z3), YzrMulti4},
		{regListAddr(REG_Z4, REG_Z7), YzrMulti4},

		{memAddr(REG_AL, REG_NONE), Ym},
		{memAddr(REG_AL, REG_SI), Ym},
		{memAddr(REG_SI, REG_CX), Ym},
		{memAddr(REG_DI, REG_X0), Yxvm},
		{memAddr(REG_DI, REG_X7), Yxvm},
		{memAddr(REG_DI, REG_Y0), Yyvm},
		{memAddr(REG_DI, REG_Y7), Yyvm},
		{memAddr(REG_DI, REG_Z0), Yzvm},
		{memAddr(REG_DI, REG_Z7), Yzvm},
	}

	oclassTestsAMD64 = []*oclassTest{
		{immAddr(-200), Ys32},
		{immAddr(500), Ys32},
		{immAddr(0x7FFFFFFF), Ys32},
		{immAddr(0x7FFFFFFF + 1), Yi32},
		{immAddr(0xFFFFFFFF), Yi32},
		{immAddr(0xFFFFFFFF + 1), Yi64},

		{regAddr(REG_BPB), Yrb},
		{regAddr(REG_SIB), Yrb},
		{regAddr(REG_DIB), Yrb},
		{regAddr(REG_R8B), Yrb},
		{regAddr(REG_R12B), Yrb},
		{regAddr(REG_R8), Yrl},
		{regAddr(REG_R13), Yrl},
		{regAddr(REG_R15), Yrl},
		{regAddr(REG_SP), Yrl},
		{regAddr(REG_SI), Yrl},
		{regAddr(REG_DI), Yrl},
		{regAddr(REG_Z13), Yzr},
		{regAddr(REG_Z20), Yzr},
		{regAddr(REG_Z31), Yzr},

		{regListAddr(REG_X10, REG_X13), YxrEvexMulti4},
		{regListAddr(REG_X24, REG_X27), YxrEvexMulti4},
		{regListAddr(REG_Y10, REG_Y13), YyrEvexMulti4},
		{regListAddr(REG_Y24, REG_Y27), YyrEvexMulti4},
		{regListAddr(REG_Z10, REG_Z13), YzrMulti4},
		{regListAddr(REG_Z24, REG_Z27), YzrMulti4},

		{memAddr(REG_DI, REG_X20), YxvmEvex},
		{memAddr(REG_DI, REG_X27), YxvmEvex},
		{memAddr(REG_DI, REG_Y20), YyvmEvex},
		{memAddr(REG_DI, REG_Y27), YyvmEvex},
		{memAddr(REG_DI, REG_Z20), Yzvm},
		{memAddr(REG_DI, REG_Z27), Yzvm},
	}

	oclassTests386 = []*oclassTest{
		{&obj.Addr{Type: obj.TYPE_ADDR, Name: obj.NAME_EXTERN, Sym: &obj.LSym{}}, Yi32},

		{immAddr(-200), Yi32},

		{regAddr(REG_SP), Yrl32},
		{regAddr(REG_SI), Yrl32},
		{regAddr(REG_DI), Yrl32},
	}

	// Add tests that are arch-independent for all sets.
	oclassTestsAMD64 = append(oclassTestsAMD64, oclassTestsCommon...)
	oclassTests386 = append(oclassTests386, oclassTestsCommon...)
}

func TestOclass(t *testing.T) {
	runTest := func(t *testing.T, ctxt *obj.Link, tests []*oclassTest) {
		var p obj.Prog
		for _, test := range tests {
			have := oclass(ctxt, &p, test.arg)
			if have != test.want {
				t.Errorf("oclass(%q):\nhave: %d\nwant: %d",
					obj.Dconv(&p, test.arg), have, test.want)
			}
		}
	}

	// TODO(quasilyte): test edge cases for Hsolaris, etc?

	t.Run("linux/AMD64", func(t *testing.T) {
		ctxtAMD64 := obj.Linknew(&Linkamd64)
		ctxtAMD64.Headtype = objabi.Hlinux // See #32028
		runTest(t, ctxtAMD64, oclassTestsAMD64)
	})

	t.Run("linux/386", func(t *testing.T) {
		ctxt386 := obj.Linknew(&Link386)
		ctxt386.Headtype = objabi.Hlinux // See #32028
		runTest(t, ctxt386, oclassTests386)
	})
}

func TestRegisterListEncDec(t *testing.T) {
	tests := []struct {
		printed string
		reg0    int16
		reg1    int16
	}{
		{"[R10-R13]", REG_R10, REG_R13},
		{"[X0-AX]", REG_X0, REG_AX},

		{"[X0-X3]", REG_X0, REG_X3},
		{"[X21-X24]", REG_X21, REG_X24},

		{"[Y0-Y3]", REG_Y0, REG_Y3},
		{"[Y21-Y24]", REG_Y21, REG_Y24},

		{"[Z0-Z3]", REG_Z0, REG_Z3},
		{"[Z21-Z24]", REG_Z21, REG_Z24},
	}

	for _, test := range tests {
		enc := EncodeRegisterRange(test.reg0, test.reg1)
		reg0, reg1 := decodeRegisterRange(enc)

		if int16(reg0) != test.reg0 {
			t.Errorf("%s reg0 mismatch: have %d, want %d",
				test.printed, reg0, test.reg0)
		}
		if int16(reg1) != test.reg1 {
			t.Errorf("%s reg1 mismatch: have %d, want %d",
				test.printed, reg1, test.reg1)
		}
		wantPrinted := test.printed
		if rlconv(enc) != wantPrinted {
			t.Errorf("%s string mismatch: have %s, want %s",
				test.printed, rlconv(enc), wantPrinted)
		}
	}
}

func TestRegIndex(t *testing.T) {
	tests := []struct {
		regFrom int
		regTo   int
	}{
		{REG_AL, REG_R15B},
		{REG_AX, REG_R15},
		{REG_M0, REG_M7},
		{REG_K0, REG_K7},
		{REG_X0, REG_X31},
		{REG_Y0, REG_Y31},
		{REG_Z0, REG_Z31},
	}

	for _, test := range tests {
		for index, reg := 0, test.regFrom; reg <= test.regTo; index, reg = index+1, reg+1 {
			have := regIndex(int16(reg))
			want := index
			if have != want {
				regName := rconv(reg)
				t.Errorf("regIndex(%s):\nhave: %d\nwant: %d",
					regName, have, want)
			}
		}
	}
}

// TestPCALIGN verifies the correctness of the PCALIGN by checking if the
// code can be aligned to the alignment value.
func TestPCALIGN(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	dir := t.TempDir()
	tmpfile := filepath.Join(dir, "test.s")
	tmpout := filepath.Join(dir, "test.o")

	var testCases = []struct {
		name string
		code string
		out  string
	}{
		{
			name: "8-byte alignment",
			code: "TEXT ·foo(SB),$0-0\nMOVQ $0, AX\nPCALIGN $8\nMOVQ $1, BX\nRET\n",
			out:  `0x0008\s00008\s\(.*\)\tMOVQ\t\$1,\sBX`,
		},
		{
			name: "16-byte alignment",
			code: "TEXT ·foo(SB),$0-0\nMOVQ $0, AX\nPCALIGN $16\nMOVQ $2, CX\nRET\n",
			out:  `0x0010\s00016\s\(.*\)\tMOVQ\t\$2,\sCX`,
		},
		{
			// Issue 81792: a forward jump to PCALIGN that must be
			// widened to rel32 used to hang the assembler.
			name: "long forward jump to PCALIGN",
			code: "TEXT ·foo(SB),$0-0\nJMP skip\n" + strings.Repeat("BYTE $0xcc\n", 136) + "skip:\nPCALIGN $64\nRET\n",
			out:  `0x0000\s00000\s\(.*\)\tJMP\t141`,
		},
	}

	for _, test := range testCases {
		if err := os.WriteFile(tmpfile, []byte(test.code), 0644); err != nil {
			t.Fatal(err)
		}
		cmd := testenv.Command(t, testenv.GoToolPath(t), "tool", "asm", "-S", "-o", tmpout, tmpfile)
		cmd.Env = append(os.Environ(), "GOARCH=amd64", "GOOS=linux")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("The %s build failed: %v, output: %s", test.name, err, out)
			continue
		}

		matched, err := regexp.MatchString(test.out, string(out))
		if err != nil {
			t.Fatal(err)
		}
		if !matched {
			t.Errorf("The %s testing failed!\ninput: %s\noutput: %s\n", test.name, test.code, out)
		}
	}
}

// TestSmallLoopAlignment exercises the complete instruction-layout pass, rather
// than depending on a particular Go function name or register allocation.
func TestSmallLoopAlignment(t *testing.T) {
	type testCase struct {
		name          string
		prefix        int
		precedingLoop bool
		fallsThrough  bool
		isAsm         bool
		calls         bool
		singleExit    bool
		longExit      bool
		multiple      bool
		wantAligned   bool
		wantFit       bool
		wantSize      int64
		checkOnly     bool
		mode32        bool
		metadata      bool
		entryAlias    bool
		indirect      bool
		explicit      bool
		nested        bool
		indexReg      int16
		tail          int
	}
	tests := []testCase{
		{name: "short rotated search", prefix: 29, wantAligned: true},
		{name: "already fits window", prefix: 0},
		{name: "preserve following fitting loop", prefix: 6, precedingLoop: true, wantAligned: true},
		{name: "function footprint budget", prefix: 29, tail: 24, wantSize: 190},
		{name: "extended index register", prefix: 29, indexReg: REG_R8, wantAligned: true},
		{name: "zero width metadata", prefix: 29, metadata: true, wantAligned: true},
		{name: "entry through metadata alias", prefix: 29, entryAlias: true},
		{name: "indirect jump in loop", prefix: 29, indirect: true},
		{name: "explicit alignment", prefix: 29, explicit: true},
		{name: "nested loop", prefix: 29, nested: true},
		{name: "386 unchanged", prefix: 23, mode32: true},
		{name: "larger unreachable padding", prefix: 25, wantAligned: true},
		{name: "fallthrough entry", prefix: 19, fallsThrough: true},
		{name: "handwritten assembly", prefix: 29, isAsm: true},
		{name: "call in loop", prefix: 29, calls: true},
		{name: "single branch", prefix: 29, singleExit: true},
		{name: "long exit", prefix: 29, longExit: true},
		{name: "multiple backedges", prefix: 29, multiple: true},
	}
	for prefix := 0; prefix < 32; prefix++ {
		tests = append(tests, testCase{name: fmt.Sprintf("offset/%d", prefix), prefix: prefix, wantFit: true})
	}
	for prefix := 0; prefix < 32; prefix++ {
		for _, tail := range []int{24, 32, 40, 48, 56, 64} {
			tests = append(tests, testCase{
				name:      fmt.Sprintf("branch-span/%d/%d", prefix, tail),
				prefix:    prefix,
				tail:      tail,
				checkOnly: true,
			})
		}
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			arch, inc, cmpAs, mode := &Linkamd64, AINCQ, ACMPQ, 64
			if test.mode32 {
				arch, inc, cmpAs, mode = &Link386, AINCL, ACMPL, 32
			}
			ctxt := obj.Linknew(arch)
			ctxt.Headtype = objabi.Hlinux
			ctxt.IsAsm = test.isAsm
			ctxt.DiagFunc = t.Errorf
			s := &obj.LSym{Name: "example.scan", Type: objabi.STEXT}
			s.NewFuncInfo()
			text := ctxt.NewProg()
			text.As = obj.ATEXT
			text.From = obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: s}
			text.To = obj.Addr{Type: obj.TYPE_TEXTSIZE, Val: int32(0)}
			s.Func().Text = text
			last := text
			add := func(as obj.As) *obj.Prog {
				p := ctxt.NewProg()
				p.As = as
				if as == obj.AJMP || as == AJLE || as == AJEQ || as == AJNE {
					p.To.Type = obj.TYPE_BRANCH
				}
				last.Link = p
				last = p
				return p
			}
			reg := func(r int16) obj.Addr { return obj.Addr{Type: obj.TYPE_REG, Reg: r} }
			for i := 0; i < test.prefix; i++ {
				add(inc).To = reg(REG_AX)
			}
			var precedingExit *obj.Prog
			if test.precedingLoop {
				firstEntry := add(obj.AJMP)
				firstHead := add(AINCQ)
				firstHead.To = reg(REG_DX)
				firstCmp := add(ACMPQ)
				firstCmp.From, firstCmp.To = reg(REG_BX), reg(REG_DX)
				firstEntry.To.SetTarget(firstCmp)
				precedingExit = add(AJLE)
				firstTest := add(ATESTQ)
				firstTest.From, firstTest.To = reg(REG_AX), reg(REG_AX)
				add(AJEQ).To.SetTarget(firstHead)
			}
			var alias *obj.Prog
			if test.entryAlias {
				alias = add(AJLE)
			}
			var entry *obj.Prog
			if !test.fallsThrough {
				entry = add(obj.AJMP)
				if precedingExit != nil {
					precedingExit.To.SetTarget(entry)
				}
			}
			if alias != nil {
				a := add(obj.APCDATA)
				a.From.SetConst(0)
				a.To.SetConst(-1)
				alias.To.SetTarget(a)
			}
			var target *obj.Prog
			if test.metadata {
				target = add(obj.APCDATA)
				target.From.SetConst(0)
				target.To.SetConst(-1)
			}
			index := test.indexReg
			if index == 0 {
				index = REG_SI
			}
			head := add(inc)
			head.To = reg(index)
			if target == nil {
				target = head
			}
			if test.explicit {
				a := add(obj.APCALIGNMAX)
				a.From.SetConst(64)
				a.To.SetConst(31)
			}
			cmp := add(cmpAs)
			cmp.From, cmp.To = reg(REG_BX), reg(index)
			if entry != nil {
				entry.To.SetTarget(cmp)
			}
			var exit *obj.Prog
			if !test.singleExit {
				exit = add(AJLE)
			}
			load := add(AMOVBLZX)
			load.From = obj.Addr{Type: obj.TYPE_MEM, Reg: REG_AX, Index: index, Scale: 1}
			load.To = reg(REG_DX)
			if test.nested {
				inner := add(ADECQ)
				inner.To = reg(REG_R10)
				add(AJNE).To.SetTarget(inner)
			}
			lookup := add(AMOVBLZX)
			lookup.From = obj.Addr{Type: obj.TYPE_MEM, Reg: REG_SP, Index: REG_DX, Scale: 1, Offset: 0x12c}
			lookup.To = reg(REG_DX)
			if test.calls || test.nested || test.indirect {
				lookup.From.Offset = 0
			}
			if test.calls {
				call := add(obj.ACALL)
				call.To = obj.Addr{Type: obj.TYPE_MEM, Name: obj.NAME_EXTERN, Sym: ctxt.Lookup("example.callee")}
			}
			if test.indirect {
				add(obj.AJMP).To = reg(REG_R11)
			}
			check := add(ATESTB)
			check.From, check.To = reg(REG_DL), reg(REG_DL)
			back := add(AJEQ)
			back.To.SetTarget(target)
			if test.multiple {
				add(obj.AJMP).To.SetTarget(target)
			}
			tail := test.tail
			if test.longExit {
				tail = 80
			}
			if tail != 0 {
				for i := 0; i < tail; i++ {
					add(inc).To = reg(REG_AX)
				}
			}
			ret := add(obj.ARET)
			if exit != nil {
				exit.To.SetTarget(ret)
			}
			span6(ctxt, s, ctxt.NewProg)
			if ctxt.Errors != 0 {
				t.Fatalf("assembly reported %d errors", ctxt.Errors)
			}
			for p := text; p != nil; p = p.Link {
				if dst := p.To.Target(); dst != nil {
					inst, err := x86asm.Decode(s.P[p.Pc:], mode)
					if err != nil {
						t.Fatalf("decode branch at %#x: %v", p.Pc, err)
					}
					rel, ok := inst.Args[0].(x86asm.Rel)
					if !ok {
						t.Fatalf("branch at %#x has argument %T", p.Pc, inst.Args[0])
					}
					if got := p.Pc + int64(inst.Len) + int64(rel); got != dst.Pc {
						t.Fatalf("branch at %#x targets %#x, want %#x", p.Pc, got, dst.Pc)
					}
				}
			}
			if test.wantSize != 0 && s.Size != test.wantSize {
				t.Fatalf("function size %d, want %d", s.Size, test.wantSize)
			}
			loopSize := back.Pc + int64(back.Isize) - head.Pc
			if test.wantFit {
				if size := loopSize; head.Pc&31+size > 31 {
					t.Fatalf("loop at %#x, size %d crosses its decoded-cache window", head.Pc, size)
				}
			} else if aligned := head.Pc%32 == 0; !test.checkOnly && aligned != test.wantAligned {
				t.Fatalf("loop at %#x, aligned=%v, want %v (body %d bytes)", head.Pc, aligned, test.wantAligned, loopSize)
			}
			if test.wantAligned && loopSize > 31 {
				t.Fatal("aligned loop does not fit in a 32-byte window")
			}
		})
	}
}
