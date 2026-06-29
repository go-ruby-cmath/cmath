// Copyright (c) the go-ruby-cmath/cmath authors
//
// SPDX-License-Identifier: BSD-3-Clause

package cmath

import (
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

// The oracle suite drives the real `ruby -rcmath` binary and compares its output,
// shape and value, against this package across a broad corpus of real, negative,
// out-of-domain and genuinely complex inputs. It skips itself when ruby is absent
// (the Windows lane and the qemu cross-arch lanes) or too old, so the
// deterministic suite alone holds the 100% gate there; on the ubuntu/macos lanes
// it confirms genuine MRI parity.

// rubyBin locates a usable `ruby` once, requiring CMath (cmath) to load and the
// interpreter to be MRI >= 4.0; otherwise it skips the oracle.
func rubyBin(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("ruby")
	if err != nil {
		t.Skip("ruby not on PATH; skipping MRI oracle")
	}
	// Gate on RUBY_VERSION >= "4.0" and that the cmath library is loadable.
	out, err := exec.Command(path, "-e",
		`exit(RUBY_VERSION >= "4.0" ? (begin; require "cmath"; 0; rescue LoadError; 2; end) : 1)`).CombinedOutput()
	if err != nil {
		t.Skipf("ruby unsuitable for oracle (version/cmath gate): %v\n%s", err, out)
	}
	return path
}

// rubyEval runs a CMath script under MRI and returns its stdout. The script
// binmodes both standard streams so Windows text-mode never rewrites the bytes
// (the go-ruby-erb lesson), and the package preamble formats a result as either
// "R<tab>real" or "C<tab>real<tab>imag" so the shape is unambiguous.
func rubyEval(t *testing.T, bin, expr string) string {
	t.Helper()
	preamble := "$stdout.binmode\n$stdin.binmode\nrequire 'cmath'\n" +
		"def emit(n)\n" +
		"  if n.is_a?(Complex)\n" +
		"    printf(\"C\\t%.17g\\t%.17g\\n\", n.real, n.imaginary)\n" +
		"  else\n" +
		"    printf(\"R\\t%.17g\\n\", n)\n" +
		"  end\n" +
		"end\n"
	cmd := exec.Command(bin, "-e", preamble+"emit("+expr+")")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ruby error: %v\nexpr: %s\noutput:\n%s", err, expr, out)
	}
	return strings.TrimSpace(string(out))
}

// parseOracle turns "R<tab>v" / "C<tab>re<tab>im" into a Number.
func parseOracle(t *testing.T, s string) Number {
	t.Helper()
	f := strings.Split(s, "\t")
	switch f[0] {
	case "R":
		v, err := strconv.ParseFloat(f[1], 64)
		if err != nil {
			t.Fatalf("parse real %q: %v", s, err)
		}
		return Real(v)
	case "C":
		re, err1 := strconv.ParseFloat(f[1], 64)
		im, err2 := strconv.ParseFloat(f[2], 64)
		if err1 != nil || err2 != nil {
			t.Fatalf("parse complex %q: %v / %v", s, err1, err2)
		}
		return Complex(re, im)
	default:
		t.Fatalf("unrecognised oracle output %q", s)
		return Number{}
	}
}

// oracleTol is looser than the deterministic tolerance because the Go and MRI
// libm differ in the last couple of ULPs on the composed transcendental paths
// (e.g. cbrt via pow); the shape, however, must match exactly.
const oracleTol = 1e-9

func assertMatch(t *testing.T, name string, got, want Number) {
	t.Helper()
	if got.IsComplex != want.IsComplex {
		t.Errorf("%s: shape mismatch got complex=%v want complex=%v (got=%v want=%v)",
			name, got.IsComplex, want.IsComplex, got, want)
		return
	}
	closeEnough := func(a, b float64) bool {
		if math.IsNaN(b) {
			return math.IsNaN(a)
		}
		if math.IsInf(b, 0) {
			return a == b
		}
		return math.Abs(a-b) <= oracleTol*(1+math.Abs(b))
	}
	if !closeEnough(got.Real, want.Real) || (got.IsComplex && !closeEnough(got.Imag, want.Imag)) {
		t.Errorf("%s: value mismatch got=%v want=%v", name, got, want)
	}
}

// rubyArg renders a Number as a Ruby literal for the oracle script.
func rubyArg(n Number) string {
	if n.IsComplex {
		return fmt.Sprintf("Complex(%s,%s)", floatLit(n.Real), floatLit(n.Imag))
	}
	return floatLit(n.Real)
}

func floatLit(f float64) string {
	switch {
	case math.IsInf(f, 1):
		return "Float::INFINITY"
	case math.IsInf(f, -1):
		return "-Float::INFINITY"
	}
	return strconv.FormatFloat(f, 'g', 17, 64)
}

// TestOracleUnary compares every single-argument CMath function against MRI over a
// corpus spanning the real branch, the negative/out-of-domain branch, and genuine
// complex inputs.
func TestOracleUnary(t *testing.T) {
	bin := rubyBin(t)

	fns := []struct {
		name string
		ruby string
		go_  func(Number) Number
	}{
		{"sqrt", "CMath.sqrt", Sqrt},
		{"exp", "CMath.exp", Exp},
		{"log", "CMath.log", func(n Number) Number { return Log(n) }},
		{"log2", "CMath.log2", Log2},
		{"log10", "CMath.log10", Log10},
		{"sin", "CMath.sin", Sin},
		{"cos", "CMath.cos", Cos},
		{"tan", "CMath.tan", Tan},
		{"sinh", "CMath.sinh", Sinh},
		{"cosh", "CMath.cosh", Cosh},
		{"tanh", "CMath.tanh", Tanh},
		{"asin", "CMath.asin", Asin},
		{"acos", "CMath.acos", Acos},
		{"atan", "CMath.atan", Atan},
		{"asinh", "CMath.asinh", Asinh},
		{"acosh", "CMath.acosh", Acosh},
		{"atanh", "CMath.atanh", Atanh},
		{"cbrt", "CMath.cbrt", Cbrt},
	}

	args := []Number{
		Real(4), Real(0.25), Real(2), Real(0.7), Real(-0.7), Real(0.3),
		Real(-4), Real(3), Real(5), Real(-16), Real(0.001), Real(0.2),
		Complex(1, 1), Complex(3, -4), Complex(-1, 0.5), Complex(0, 3),
		Complex(2, -1), Complex(0.5, 0.5),
	}

	for _, fn := range fns {
		for _, a := range args {
			name := fmt.Sprintf("%s(%v)", fn.name, a)
			t.Run(name, func(t *testing.T) {
				want := parseOracle(t, rubyEval(t, bin, fn.ruby+"("+rubyArg(a)+")"))
				assertMatch(t, name, fn.go_(a), want)
			})
		}
	}
}

// TestOracleLogBase exercises the two-argument CMath.log, including non-E and
// negative bases that force the complex divide.
func TestOracleLogBase(t *testing.T) {
	bin := rubyBin(t)
	cases := []struct {
		z, b Number
	}{
		{Real(8), Real(2)},
		{Real(7), Real(3)},
		{Real(8), Real(-2)},
		{Real(-8), Real(2)},
		{Complex(2, 3), Real(10)},
		{Complex(1, 4), Real(10)},
		{Real(1000), Real(10)},
	}
	for _, c := range cases {
		name := fmt.Sprintf("log(%v,%v)", c.z, c.b)
		t.Run(name, func(t *testing.T) {
			expr := "CMath.log(" + rubyArg(c.z) + "," + rubyArg(c.b) + ")"
			want := parseOracle(t, rubyEval(t, bin, expr))
			assertMatch(t, name, Log(c.z, c.b), want)
		})
	}
}

// TestOracleAtan2 exercises atan2 over real-real (real result) and mixed-complex
// (complex result) argument pairs.
func TestOracleAtan2(t *testing.T) {
	bin := rubyBin(t)
	cases := [][2]Number{
		{Real(2), Real(3)},
		{Real(-1), Real(-1)},
		{Complex(1, 2), Complex(0, 1)},
		{Complex(1, 1), Complex(1, 0)},
	}
	for _, c := range cases {
		name := fmt.Sprintf("atan2(%v,%v)", c[0], c[1])
		t.Run(name, func(t *testing.T) {
			expr := "CMath.atan2(" + rubyArg(c[0]) + "," + rubyArg(c[1]) + ")"
			want := parseOracle(t, rubyEval(t, bin, expr))
			assertMatch(t, name, Atan2(c[0], c[1]), want)
		})
	}
}
