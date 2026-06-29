// Copyright (c) the go-ruby-cmath/cmath authors
//
// SPDX-License-Identifier: BSD-3-Clause

// Package cmath is a pure-Go (no cgo) reimplementation of Ruby's CMath
// standard library — the complex-aware trigonometric and transcendental
// functions of MRI 4.0.5's `cmath` gem.
//
// CMath accepts real or complex arguments and returns a real result when the
// input lies on the real branch of the function (so CMath.sqrt(4) is 2.0) and a
// complex result otherwise (CMath.sqrt(-4) is (0+2i)). This package mirrors that
// real-or-complex value model with the [Number] type and reproduces MRI's exact
// branch-cut decisions and float formulas, so it can back a Ruby runtime such as
// go-embedded-ruby without any interpreter.
//
// The mapping to MRI is faithful to the reference implementation: every function
// here uses the same real-vs-complex test and the same closed-form expression
// that the cmath gem's Ruby source does (e.g. asin is built from this package's
// own Log and Sqrt, exactly as CMath.asin is), so results agree with the `ruby`
// binary to within floating-point rounding.
package cmath

import "math"

// Number is a real-or-complex value, the Go analogue of a Ruby Float-or-Complex.
// A real Number has IsComplex false and carries its value in Real; a complex one
// has IsComplex true and uses both Real and Imag. The CMath functions take and
// return Number, choosing the real shape exactly where MRI's CMath does.
type Number struct {
	Real      float64
	Imag      float64
	IsComplex bool
}

// Real builds a real Number (a Ruby Float / Integer argument).
func Real(x float64) Number { return Number{Real: x} }

// Complex builds a complex Number (a Ruby Complex argument). It stays complex
// even when the imaginary part is zero, matching Ruby: Complex(1,0).real? is
// false, so CMath treats it as complex.
func Complex(re, im float64) Number { return Number{Real: re, Imag: im, IsComplex: true} }

// IsReal reports whether the Number is on the real branch — the Go analogue of
// Ruby's Numeric#real?, which is true for Integer/Float and false for Complex.
func (n Number) IsReal() bool { return !n.IsComplex }

// re and im return the components regardless of shape.
func (n Number) re() float64 { return n.Real }
func (n Number) im() float64 {
	if n.IsComplex {
		return n.Imag
	}
	return 0
}

// abs is Ruby's Numeric#abs / Complex#abs (the modulus).
func (n Number) abs() float64 {
	if n.IsComplex {
		return math.Hypot(n.Real, n.Imag)
	}
	return math.Abs(n.Real)
}

// arg is Ruby's Numeric#arg / Complex#arg (the phase angle). For a real value it
// is π when the value is negative and 0 otherwise, matching MRI. Within this
// package real arg is only ever evaluated for a strictly negative real (Log's
// complex branch already excludes the non-negative reals, and -0.0 is treated as
// >= 0 there), so the negative-real case alone is exercised.
func (n Number) arg() float64 {
	if n.IsComplex {
		return math.Atan2(n.Imag, n.Real)
	}
	if n.Real < 0 {
		return math.Pi
	}
	return 0
}

// --- Ruby complex arithmetic, reproduced so results match MRI bit-for-bit ---
//
// CMath composes its functions out of Complex +, -, *, / and the polar power.
// Ruby uses the naive (textbook) complex multiply and divide on floats, not the
// scaled Smith algorithm that Go's built-in complex128 operators may select for
// overflow safety, so we implement the naive forms here to stay faithful.

// bothReal reports whether two operands are both on the real branch. Ruby's
// numeric coercion keeps an arithmetic result real exactly when both operands
// are real and goes complex otherwise; preserving that is essential because a
// negative real has argument +π while a complex value carrying a -0.0 imaginary
// part has argument -π, which would flip branch cuts (e.g. atanh's).
func bothReal(a, b Number) bool { return a.IsReal() && b.IsReal() }

func cadd(a, b Number) Number {
	if bothReal(a, b) {
		return Real(a.Real + b.Real)
	}
	return Number{Real: a.re() + b.re(), Imag: a.im() + b.im(), IsComplex: true}
}

func csub(a, b Number) Number {
	if bothReal(a, b) {
		return Real(a.Real - b.Real)
	}
	return Number{Real: a.re() - b.re(), Imag: a.im() - b.im(), IsComplex: true}
}

func cmul(a, b Number) Number {
	if bothReal(a, b) {
		return Real(a.Real * b.Real)
	}
	ar, ai, br, bi := a.re(), a.im(), b.re(), b.im()
	return Number{Real: ar*br - ai*bi, Imag: ar*bi + ai*br, IsComplex: true}
}

func cdiv(a, b Number) Number {
	if bothReal(a, b) {
		return Real(a.Real / b.Real)
	}
	ar, ai, br, bi := a.re(), a.im(), b.re(), b.im()
	den := br*br + bi*bi
	return Number{Real: (ar*br + ai*bi) / den, Imag: (ai*br - ar*bi) / den, IsComplex: true}
}

// scaleDiv divides a Number by a real scalar (the Ruby `expr / 2.0` etc.). It
// preserves realness so a real operand divided by a real scalar stays real.
func scaleDiv(a Number, s float64) Number {
	if a.IsReal() {
		return Real(a.Real / s)
	}
	return Number{Real: a.re() / s, Imag: a.im() / s, IsComplex: true}
}

// mulI multiplies by 1.0i (Ruby's 1.0.i * z): (a+bi)*i = -b + ai.
func mulI(z Number) Number { return Number{Real: -z.im(), Imag: z.re(), IsComplex: true} }

// mulNegI multiplies by -1.0i (Ruby's (-1.0).i * z): -(a+bi)*i = b - ai.
func mulNegI(z Number) Number { return Number{Real: z.im(), Imag: -z.re(), IsComplex: true} }

// polar builds Complex(r*cos(theta), r*sin(theta)) — Ruby's Complex.polar, used
// by the power operator.
func polar(r, theta float64) Number {
	return Number{Real: r * math.Cos(theta), Imag: r * math.Sin(theta), IsComplex: true}
}

// powFrac raises a Number to a real, generally fractional, power, reproducing
// Ruby's Numeric#** for that case: a non-negative real stays real; a negative
// real or any complex base goes through the polar form and becomes complex.
func powFrac(z Number, p float64) Number {
	if z.IsReal() && z.Real >= 0 {
		return Real(math.Pow(z.Real, p))
	}
	r := z.abs()
	return polar(math.Pow(r, p), p*z.arg())
}
