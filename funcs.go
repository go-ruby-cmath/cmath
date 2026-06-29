// Copyright (c) the go-ruby-cmath/cmath authors
//
// SPDX-License-Identifier: BSD-3-Clause

package cmath

import "math"

// Each function below mirrors the same-named method of Ruby's CMath module: the
// real-branch guard, the real-result expression, and the complex fallback are a
// line-for-line port of the cmath gem's source, so the real-vs-complex shape and
// the float results agree with the `ruby -rcmath` oracle.

// Exp returns Math::E raised to the z power. CMath.exp(z).
func Exp(z Number) Number {
	if z.IsReal() {
		return Real(math.Exp(z.Real))
	}
	ere := math.Exp(z.re())
	return Number{Real: ere * math.Cos(z.im()), Imag: ere * math.Sin(z.im()), IsComplex: true}
}

// Log returns the natural logarithm of z, or the base-b logarithm when a base is
// supplied. CMath.log(z[, b]). A non-negative real z with a non-negative base
// stays real; anything else is complex.
func Log(z Number, base ...Number) Number {
	b := Real(math.E)
	if len(base) > 0 {
		b = base[0]
	}
	if z.IsReal() && z.Real >= 0 && b.IsReal() && b.Real >= 0 {
		// CMath's real branch is RealMath.log(z, b) with b defaulting to E, i.e.
		// Math.log(z, b) == log(z)/log(b). This holds even with no explicit base,
		// so CMath.log(10) is log(10)/log(E), not Math.log(10) — we always divide.
		return Real(math.Log(z.Real) / math.Log(b.Real))
	}
	// Complex(log(|z|), arg(z)) / log(b).
	num := Number{Real: math.Log(z.abs()), Imag: z.arg(), IsComplex: true}
	return cdiv(num, Log(b))
}

// Log2 returns the base-2 logarithm of z. CMath.log2(z).
func Log2(z Number) Number {
	if z.IsReal() && z.Real >= 0 {
		return Real(math.Log2(z.Real))
	}
	return scaleDiv(Log(z), math.Log(2))
}

// Log10 returns the base-10 logarithm of z. CMath.log10(z).
func Log10(z Number) Number {
	if z.IsReal() && z.Real >= 0 {
		return Real(math.Log10(z.Real))
	}
	return scaleDiv(Log(z), math.Log(10))
}

// Sqrt returns the principal square root of z. CMath.sqrt(z). A non-negative
// real stays real; a negative real becomes a pure imaginary; a complex argument
// uses Ruby's carry-of-the-sign formula and conjugate branch.
func Sqrt(z Number) Number {
	if z.IsReal() {
		if z.Real < 0 {
			return Complex(0, math.Sqrt(-z.Real))
		}
		return Real(math.Sqrt(z.Real))
	}
	// Complex branch, matching CMath.sqrt's conjugate handling and the
	// (r±x)/2 closed form so the principal root and its branch cut agree.
	if z.im() < 0 || (z.im() == 0 && math.Signbit(z.im())) {
		s := Sqrt(Number{Real: z.re(), Imag: -z.im(), IsComplex: true})
		return Number{Real: s.re(), Imag: -s.im(), IsComplex: true}
	}
	r := z.abs()
	x := z.re()
	return Number{Real: math.Sqrt((r + x) / 2.0), Imag: math.Sqrt((r - x) / 2.0), IsComplex: true}
}

// Cbrt returns the principal cube root of z. CMath.cbrt(z) == z ** (1.0/3).
func Cbrt(z Number) Number {
	return powFrac(z, 1.0/3.0)
}

// Sin returns the sine of z (radians). CMath.sin(z).
func Sin(z Number) Number {
	if z.IsReal() {
		return Real(math.Sin(z.Real))
	}
	return Number{
		Real:      math.Sin(z.re()) * math.Cosh(z.im()),
		Imag:      math.Cos(z.re()) * math.Sinh(z.im()),
		IsComplex: true,
	}
}

// Cos returns the cosine of z (radians). CMath.cos(z).
func Cos(z Number) Number {
	if z.IsReal() {
		return Real(math.Cos(z.Real))
	}
	return Number{
		Real:      math.Cos(z.re()) * math.Cosh(z.im()),
		Imag:      -math.Sin(z.re()) * math.Sinh(z.im()),
		IsComplex: true,
	}
}

// Tan returns the tangent of z (radians). CMath.tan(z).
func Tan(z Number) Number {
	if z.IsReal() {
		return Real(math.Tan(z.Real))
	}
	return cdiv(Sin(z), Cos(z))
}

// Sinh returns the hyperbolic sine of z. CMath.sinh(z).
func Sinh(z Number) Number {
	if z.IsReal() {
		return Real(math.Sinh(z.Real))
	}
	return Number{
		Real:      math.Sinh(z.re()) * math.Cos(z.im()),
		Imag:      math.Cosh(z.re()) * math.Sin(z.im()),
		IsComplex: true,
	}
}

// Cosh returns the hyperbolic cosine of z. CMath.cosh(z).
func Cosh(z Number) Number {
	if z.IsReal() {
		return Real(math.Cosh(z.Real))
	}
	return Number{
		Real:      math.Cosh(z.re()) * math.Cos(z.im()),
		Imag:      math.Sinh(z.re()) * math.Sin(z.im()),
		IsComplex: true,
	}
}

// Tanh returns the hyperbolic tangent of z. CMath.tanh(z).
func Tanh(z Number) Number {
	if z.IsReal() {
		return Real(math.Tanh(z.Real))
	}
	return cdiv(Sinh(z), Cosh(z))
}

// Asin returns the arc sine of z. CMath.asin(z). Real for z in [-1, 1].
func Asin(z Number) Number {
	if z.IsReal() && z.Real >= -1 && z.Real <= 1 {
		return Real(math.Asin(z.Real))
	}
	// (-1.0).i * log(1.0.i * z + sqrt(1.0 - z*z))
	inner := cadd(mulI(z), Sqrt(csub(Real(1.0), cmul(z, z))))
	return mulNegI(Log(inner))
}

// Acos returns the arc cosine of z. CMath.acos(z). Real for z in [-1, 1].
func Acos(z Number) Number {
	if z.IsReal() && z.Real >= -1 && z.Real <= 1 {
		return Real(math.Acos(z.Real))
	}
	// (-1.0).i * log(z + 1.0.i * sqrt(1.0 - z*z))
	inner := cadd(z, mulI(Sqrt(csub(Real(1.0), cmul(z, z)))))
	return mulNegI(Log(inner))
}

// Atan returns the arc tangent of z. CMath.atan(z).
func Atan(z Number) Number {
	if z.IsReal() {
		return Real(math.Atan(z.Real))
	}
	// 1.0.i * log((1.0.i + z) / (1.0.i - z)) / 2.0
	num := cadd(Complex(0, 1), z)
	den := csub(Complex(0, 1), z)
	return scaleDiv(mulI(Log(cdiv(num, den))), 2.0)
}

// Atan2 returns the arc tangent of y/x using the signs of both to select the
// quadrant. CMath.atan2(y, x). Real when both arguments are real.
func Atan2(y, x Number) Number {
	if y.IsReal() && x.IsReal() {
		return Real(math.Atan2(y.Real, x.Real))
	}
	// (-1.0).i * log((x + 1.0.i * y) / sqrt(x*x + y*y))
	num := cadd(x, mulI(y))
	den := Sqrt(cadd(cmul(x, x), cmul(y, y)))
	return mulNegI(Log(cdiv(num, den)))
}

// Asinh returns the inverse hyperbolic sine of z. CMath.asinh(z).
func Asinh(z Number) Number {
	if z.IsReal() {
		return Real(math.Asinh(z.Real))
	}
	// log(z + sqrt(1.0 + z*z))
	return Log(cadd(z, Sqrt(cadd(Real(1.0), cmul(z, z)))))
}

// Acosh returns the inverse hyperbolic cosine of z. CMath.acosh(z). Real for
// z >= 1.
func Acosh(z Number) Number {
	if z.IsReal() && z.Real >= 1 {
		return Real(math.Acosh(z.Real))
	}
	// log(z + sqrt(z*z - 1.0))
	return Log(cadd(z, Sqrt(csub(cmul(z, z), Real(1.0)))))
}

// Atanh returns the inverse hyperbolic tangent of z. CMath.atanh(z). Real for
// z in [-1, 1].
func Atanh(z Number) Number {
	if z.IsReal() && z.Real >= -1 && z.Real <= 1 {
		return Real(math.Atanh(z.Real))
	}
	// log((1.0 + z) / (1.0 - z)) / 2.0
	return scaleDiv(Log(cdiv(cadd(Real(1.0), z), csub(Real(1.0), z))), 2.0)
}
