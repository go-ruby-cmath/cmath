<p align="center"><img src="https://raw.githubusercontent.com/go-ruby-cmath/brand/main/social/go-ruby-cmath-cmath.png" alt="go-ruby-cmath/cmath" width="720"></p>

# cmath — go-ruby-cmath

[![Docs](https://img.shields.io/badge/docs-mkdocs--material-DC2626)](https://go-ruby-cmath.github.io/docs/)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.26.4%2B-00ADD8)](https://go.dev/dl/)
[![Coverage](https://img.shields.io/badge/coverage-100%25-1a7f37)](#tests--coverage)

**A pure-Go (no cgo) reimplementation of Ruby's
[CMath](https://docs.ruby-lang.org/en/master/CMath.html) standard library** — the
complex-aware trigonometric and transcendental functions of MRI 4.0.5's `cmath`
gem. CMath accepts real *or* complex arguments and returns a **real** result when
the input lies on the real branch of the function and a **complex** result
otherwise, so `CMath.sqrt(4)` is `2.0` while `CMath.sqrt(-4)` is `(0+2i)`. This
library reproduces that exact real-vs-complex decision and MRI's branch cuts and
float formulas — **without any Ruby runtime**.

It is a CMath backend for
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby), but is a
**standalone, reusable** module — a sibling of
[go-ruby-regexp](https://github.com/go-ruby-regexp/regexp) (the Onigmo engine),
[go-ruby-marshal](https://github.com/go-ruby-marshal/marshal) (the Marshal codec)
and [go-ruby-yaml](https://github.com/go-ruby-yaml/yaml) (the Psych port).

> **What it is — and isn't.** The mathematics of CMath — the real-vs-complex
> branch test, the closed-form complex expressions, the principal branch cuts — is
> fully deterministic and needs **no interpreter**, so it lives here as pure Go.
> Binding the results to a host's live `Float`/`Complex` objects is the host's
> job; this library takes and returns a small, explicit value model ([`Number`](#the-number-value-model)).

## The `Number` value model

A `Number` is a real-or-complex value, the Go analogue of a Ruby
`Float`-or-`Complex`:

```go
n := cmath.Real(4)         // a real argument (Ruby Integer / Float)
z := cmath.Complex(0, 1)   // a complex argument (Ruby Complex)

r := cmath.Sqrt(n)         // r.IsComplex == false, r.Real == 2.0
c := cmath.Sqrt(cmath.Real(-4))
                           // c.IsComplex == true, c.Real == 0, c.Imag == 2.0
```

`cmath.Complex(re, 0)` stays complex even with a zero imaginary part, matching
Ruby: `Complex(1, 0).real?` is `false`, so CMath treats it as complex.

## API

Every function mirrors the same-named method of Ruby's `CMath` module, taking and
returning [`Number`](#the-number-value-model):

| Go | Ruby | Real branch (returns real) |
| --- | --- | --- |
| `Sqrt(z)` | `CMath.sqrt` | real `z >= 0` |
| `Exp(z)` | `CMath.exp` | real `z` |
| `Log(z[, base])` | `CMath.log` | real `z >= 0` and real `base >= 0` |
| `Log2(z)` | `CMath.log2` | real `z >= 0` |
| `Log10(z)` | `CMath.log10` | real `z >= 0` |
| `Cbrt(z)` | `CMath.cbrt` | real `z >= 0` |
| `Sin/Cos/Tan(z)` | `CMath.sin/cos/tan` | real `z` |
| `Sinh/Cosh/Tanh(z)` | `CMath.sinh/cosh/tanh` | real `z` |
| `Asin(z)` | `CMath.asin` | real `z` in `[-1, 1]` |
| `Acos(z)` | `CMath.acos` | real `z` in `[-1, 1]` |
| `Atan(z)` | `CMath.atan` | real `z` |
| `Atan2(y, x)` | `CMath.atan2` | real `y` and real `x` |
| `Asinh(z)` | `CMath.asinh` | real `z` |
| `Acosh(z)` | `CMath.acosh` | real `z >= 1` |
| `Atanh(z)` | `CMath.atanh` | real `z` in `[-1, 1]` |

Outside the real branch each function returns the complex result, computed with
the same closed-form expression CMath uses — `Asin`, for instance, is built from
this package's own `Log` and `Sqrt`, exactly as `CMath.asin` is — so the branch
cuts agree with MRI.

A few representative results, matching `ruby -rcmath`:

```go
cmath.Sqrt(cmath.Real(4))    // real 2.0
cmath.Sqrt(cmath.Real(-4))   // complex (0+2i)
cmath.Log(cmath.Real(-1))    // complex (0+πi)
cmath.Asin(cmath.Real(2))    // complex (1.5707963267948966-1.3169578969248166i)
cmath.Cbrt(cmath.Real(-8))   // complex (1+1.7320508075688772i)
```

## Tests & coverage

The suite is in two halves. The **deterministic** suite carries golden values
captured from MRI 4.0.5 and asserts both the exact real-vs-complex shape and the
float components to a tight tolerance; it needs no Ruby and alone holds coverage
at **100%**. The **oracle** suite drives the real `ruby -rcmath` binary across a
broad real / negative / out-of-domain / complex corpus and confirms genuine MRI
parity; it skips itself where `ruby` is absent or older than 4.0 (the Windows and
qemu cross-arch lanes), so the gate stays green everywhere. CI runs the suite on
three operating systems and all six supported 64-bit architectures
(amd64, arm64, riscv64, loong64, ppc64le, s390x), with cgo disabled.

```sh
go test ./...                       # deterministic + (if ruby present) oracle
COVERPKG=$(go list ./... | paste -sd, -)
go test -coverpkg="$COVERPKG" -coverprofile=cover.out ./...
go tool cover -func=cover.out | tail -1   # total: 100.0%
```

## License

BSD-3-Clause — see [LICENSE](LICENSE). Copyright (c) the go-ruby-cmath/cmath authors.
