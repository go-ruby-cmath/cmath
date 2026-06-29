// Copyright (c) the go-ruby-cmath/cmath authors
//
// SPDX-License-Identifier: BSD-3-Clause

package cmath

import (
	"math"
	"testing"
)

// The deterministic suite below carries golden values captured from MRI 4.0.5's
// `ruby -rcmath` and asserts both the exact real-vs-complex shape and the float
// components to a tight tolerance. It needs no Ruby, so it alone keeps coverage
// at 100% on the Windows and cross-arch lanes where the oracle skips.

// tol is the relative+absolute tolerance for the golden float comparisons.
const tol = 1e-12

func approx(got, want float64) bool {
	if math.IsNaN(want) {
		return math.IsNaN(got)
	}
	if math.IsInf(want, 0) {
		return got == want
	}
	return math.Abs(got-want) <= tol*(1+math.Abs(want))
}

// wantReal asserts n is real and within tolerance of v.
func wantReal(t *testing.T, name string, n Number, v float64) {
	t.Helper()
	if n.IsComplex {
		t.Errorf("%s: got complex %v, want real %.17g", name, n, v)
		return
	}
	if !approx(n.Real, v) {
		t.Errorf("%s: got real %.17g, want %.17g", name, n.Real, v)
	}
}

// wantComplex asserts n is complex and within tolerance of re+im·i.
func wantComplex(t *testing.T, name string, n Number, re, im float64) {
	t.Helper()
	if !n.IsComplex {
		t.Errorf("%s: got real %.17g, want complex (%.17g,%.17g)", name, n.Real, re, im)
		return
	}
	if !approx(n.Real, re) || !approx(n.Imag, im) {
		t.Errorf("%s: got (%.17g,%.17g), want (%.17g,%.17g)", name, n.Real, n.Imag, re, im)
	}
}

// TestReal collects every real-branch golden — the shape MRI keeps real.
func TestReal(t *testing.T) {
	wantReal(t, "sqrt(4)", Sqrt(Real(4)), 2)
	wantReal(t, "sqrt(0.25)", Sqrt(Real(0.25)), 0.5)
	wantReal(t, "exp(2)", Exp(Real(2)), 7.3890560989306504)
	wantReal(t, "log(7,3)", Log(Real(7), Real(3)), 1.7712437491614224)
	wantReal(t, "log2(16)", Log2(Real(16)), 4)
	wantReal(t, "log10(0.001)", Log10(Real(0.001)), -3)
	wantReal(t, "sin(2.5)", Sin(Real(2.5)), 0.59847214410395644)
	wantReal(t, "cos(3)", Cos(Real(3)), -0.98999249660044542)
	wantReal(t, "tan(0.7)", Tan(Real(0.7)), 0.8422883804630793)
	wantReal(t, "sinh(2)", Sinh(Real(2)), 3.6268604078470186)
	wantReal(t, "cosh(1.5)", Cosh(Real(1.5)), 2.3524096152432472)
	wantReal(t, "tanh(0.9)", Tanh(Real(0.9)), 0.71629787019902447)
	wantReal(t, "asin(0.3)", Asin(Real(0.3)), 0.30469265401539747)
	wantReal(t, "acos(-0.7)", Acos(Real(-0.7)), 2.3461938234056494)
	wantReal(t, "atan(0.4)", Atan(Real(0.4)), 0.3805063771123649)
	wantReal(t, "atan2(2,3)", Atan2(Real(2), Real(3)), 0.5880026035475675)
	wantReal(t, "asinh(3)", Asinh(Real(3)), 1.8184464592320668)
	wantReal(t, "acosh(4)", Acosh(Real(4)), 2.0634370688955608)
	wantReal(t, "atanh(0.2)", Atanh(Real(0.2)), 0.20273255405408219)
	wantReal(t, "cbrt(27)", Cbrt(Real(27)), 3)
	// Log with no explicit base divides by log(E): CMath.log(10) is log(10)/log(E).
	wantReal(t, "log(10)", Log(Real(10)), 2.302585092994045457)
}

// TestComplex collects every complex-branch golden — the shape MRI makes complex,
// covering negative reals, out-of-domain reals, and genuine complex arguments.
func TestComplex(t *testing.T) {
	wantComplex(t, "sqrt(-4)", Sqrt(Real(-4)), 0, 2)
	wantComplex(t, "sqrt(C3,-4)", Sqrt(Complex(3, -4)), 2, -1)
	wantComplex(t, "sqrt(C-1,0)", Sqrt(Complex(-1, 0)), 0, 1)
	wantComplex(t, "sqrt(C0,-0)", Sqrt(Complex(0, math.Copysign(0, -1))), 0, math.Copysign(0, -1))
	wantComplex(t, "exp(C1,2)", Exp(Complex(1, 2)), -1.1312043837568135, 2.4717266720048188)
	wantComplex(t, "log(-5)", Log(Real(-5)), 1.6094379124341003, 3.1415926535897931)
	wantComplex(t, "log(C2,3 b10)", Log(Complex(2, 3), Real(10)), 0.55697167615341847, 0.42682189085546668)
	wantComplex(t, "log(8,-2)", Log(Real(8), Real(-2)), 0.13926097063622436, -0.6311808726237905)
	wantComplex(t, "log2(-16)", Log2(Real(-16)), 4, 4.5323601418271942)
	wantComplex(t, "log10(C1,1)", Log10(Complex(1, 1)), 0.1505149978319906, 0.3410940884604603)
	wantComplex(t, "sin(C-1,0.5)", Sin(Complex(-1, 0.5)), -0.94886453143716798, 0.28154899513533443)
	wantComplex(t, "cos(C0.3,-1.2)", Cos(Complex(0.3, -1.2)), 1.7297853327034005, 0.44607633169871092)
	wantComplex(t, "tan(C1,2)", Tan(Complex(1, 2)), 0.033812826079896767, 1.0147936161466338)
	wantComplex(t, "sinh(C0.5,0.5)", Sinh(Complex(0.5, 0.5)), 0.45730415318424927, 0.54061268571315335)
	wantComplex(t, "cosh(C-1,1)", Cosh(Complex(-1, 1)), 0.83373002513114913, -0.98889770576286506)
	wantComplex(t, "tanh(C2,1)", Tanh(Complex(2, 1)), 1.0147936161466338, 0.033812826079896767)
	wantComplex(t, "asin(3)", Asin(Real(3)), 1.5707963267948966, -1.7627471740390861)
	wantComplex(t, "asin(C1,1)", Asin(Complex(1, 1)), 0.66623943249251527, 1.0612750619050355)
	wantComplex(t, "acos(5)", Acos(Real(5)), 0, 2.2924316695611733)
	wantComplex(t, "acos(C2,-1)", Acos(Complex(2, -1)), 0.50735630321714442, 1.4693517443681854)
	wantComplex(t, "atan(C0,3)", Atan(Complex(0, 3)), -1.5707963267948966, 0.34657359027997264)
	wantComplex(t, "atan2(C1,2;C0,1)", Atan2(Complex(1, 2), Complex(0, 1)), 1.1780972450961724, -0.17328679513998629)
	wantComplex(t, "asinh(C1,1)", Asinh(Complex(1, 1)), 1.0612750619050357, 0.66623943249251527)
	wantComplex(t, "acosh(0.2)", Acosh(Real(0.2)), 0, 1.3694384060045659)
	wantComplex(t, "acosh(C1,2)", Acosh(Complex(1, 2)), 1.528570919480998, 1.1437177404024204)
	wantComplex(t, "atanh(3)", Atanh(Real(3)), 0.34657359027997264, 1.5707963267948966)
	wantComplex(t, "atanh(C0.5,0.5)", Atanh(Complex(0.5, 0.5)), 0.40235947810852513, 0.5535743588970452)
	wantComplex(t, "cbrt(-27)", Cbrt(Real(-27)), 1.5, 2.598076211353316)
	wantComplex(t, "cbrt(C2,3)", Cbrt(Complex(2, 3)), 1.4518566183526649, 0.49340353410400467)
}

// TestNumberHelpers exercises the small value-model surface directly.
func TestNumberHelpers(t *testing.T) {
	if r := Real(2.5); r.IsComplex || !r.IsReal() || r.Real != 2.5 {
		t.Errorf("Real(2.5) = %v", r)
	}
	if c := Complex(1, -2); !c.IsComplex || c.IsReal() || c.Real != 1 || c.Imag != -2 {
		t.Errorf("Complex(1,-2) = %v", c)
	}
	// im() returns 0 for a real Number.
	if got := Real(5).im(); got != 0 {
		t.Errorf("Real(5).im() = %v, want 0", got)
	}
	if got := Complex(3, 4).im(); got != 4 {
		t.Errorf("Complex(3,4).im() = %v, want 4", got)
	}
	// abs and arg on both shapes.
	if got := Complex(3, 4).abs(); got != 5 {
		t.Errorf("Complex(3,4).abs() = %v, want 5", got)
	}
	if got := Real(-3).abs(); got != 3 {
		t.Errorf("Real(-3).abs() = %v, want 3", got)
	}
	if got := Complex(0, 1).arg(); !approx(got, math.Pi/2) {
		t.Errorf("Complex(0,1).arg() = %v, want π/2", got)
	}
	if got := Real(-2).arg(); got != math.Pi {
		t.Errorf("Real(-2).arg() = %v, want π", got)
	}
	if got := Real(2).arg(); got != 0 {
		t.Errorf("Real(2).arg() = %v, want 0", got)
	}
}

// TestArithmeticHelpers covers both branches (real-preserving and complex) of the
// internal Ruby-faithful arithmetic, including the i-multiplications and polar.
func TestArithmeticHelpers(t *testing.T) {
	// Real-preserving branches.
	if r := cadd(Real(1), Real(2)); r.IsComplex || r.Real != 3 {
		t.Errorf("cadd reals = %v", r)
	}
	if r := csub(Real(1), Real(2)); r.IsComplex || r.Real != -1 {
		t.Errorf("csub reals = %v", r)
	}
	if r := cmul(Real(2), Real(3)); r.IsComplex || r.Real != 6 {
		t.Errorf("cmul reals = %v", r)
	}
	if r := cdiv(Real(6), Real(3)); r.IsComplex || r.Real != 2 {
		t.Errorf("cdiv reals = %v", r)
	}
	if r := scaleDiv(Real(6), 2); r.IsComplex || r.Real != 3 {
		t.Errorf("scaleDiv real = %v", r)
	}
	// Complex branches.
	if r := cadd(Complex(1, 2), Real(3)); !r.IsComplex || r.Real != 4 || r.Imag != 2 {
		t.Errorf("cadd mixed = %v", r)
	}
	if r := csub(Complex(1, 2), Complex(0, 1)); !r.IsComplex || r.Real != 1 || r.Imag != 1 {
		t.Errorf("csub complex = %v", r)
	}
	if r := cmul(Complex(1, 2), Complex(3, 4)); !r.IsComplex || r.Real != -5 || r.Imag != 10 {
		t.Errorf("cmul complex = %v", r)
	}
	if r := cdiv(Complex(1, 2), Complex(3, 4)); !r.IsComplex || !approx(r.Real, 0.44) || !approx(r.Imag, 0.08) {
		t.Errorf("cdiv complex = %v", r)
	}
	if r := scaleDiv(Complex(4, 8), 2); !r.IsComplex || r.Real != 2 || r.Imag != 4 {
		t.Errorf("scaleDiv complex = %v", r)
	}
	if r := mulI(Complex(1, 2)); r.Real != -2 || r.Imag != 1 {
		t.Errorf("mulI = %v", r)
	}
	if r := mulNegI(Complex(1, 2)); r.Real != 2 || r.Imag != -1 {
		t.Errorf("mulNegI = %v", r)
	}
	if r := polar(2, 0); !approx(r.Real, 2) || !approx(r.Imag, 0) {
		t.Errorf("polar(2,0) = %v", r)
	}
	// powFrac on a positive real stays real; on a complex it goes polar.
	if r := powFrac(Real(8), 1.0/3); r.IsComplex {
		t.Errorf("powFrac(8) = %v, want real", r)
	}
	if r := powFrac(Complex(0, 1), 1.0/3); !r.IsComplex {
		t.Errorf("powFrac(C0,1) = %v, want complex", r)
	}
}

// TestLogDefaultBaseMatchesDivision pins the subtle CMath.log default-base path:
// even with no explicit base, the result is log(z)/log(E), not Math.log(z).
func TestLogDefaultBaseMatchesDivision(t *testing.T) {
	got := Log(Real(10))
	want := math.Log(10) / math.Log(math.E)
	if got.IsComplex || got.Real != want {
		t.Errorf("Log(10) = %v, want exactly %.17g", got, want)
	}
}

// TestSqrtConjugateBranch checks the negative-imaginary conjugate path of Sqrt,
// distinct from the upper-half-plane closed form.
func TestSqrtConjugateBranch(t *testing.T) {
	up := Sqrt(Complex(1, 1))
	lo := Sqrt(Complex(1, -1))
	// The lower-half root is the conjugate of the upper-half one.
	if !approx(lo.Real, up.Real) || !approx(lo.Imag, -up.Imag) {
		t.Errorf("Sqrt conjugate mismatch: up=%v lo=%v", up, lo)
	}
}
