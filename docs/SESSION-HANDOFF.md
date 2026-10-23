# Session handoff

Updated 2026-10-23.

## Project

Repository: https://github.com/fernandocpaz/axiom

Local path: `/Users/fernandopaz/axiom`

Branch: `main`

## Completed

- Public GitHub repository created and published.
- Go 1.15-compatible lexer, parser, and interpreter bootstrapped.
- Arithmetic, `let`, `print`, booleans, comparisons, and `if ... else ...` expressions implemented.
- String literals implemented with `\\n`, `\\t`, `\\"`, and `\\\\` escapes.
- `examples/hello.ax` now contains:

  ```axiom
  fn main() {
    print "Hello, world!"
  }
  ```

- Verified locally with `go test ./...` and `go run ./cmd/axiom examples/hello.ax`; output is exactly `Hello, world!`.
- Static checker and structured diagnostics added in commit `ac94740`.
- Source spans added to diagnostics in commit `e8f8682`.
- Lexer string position tracking and parser trailing-source validation tightened in commit `f6a8cca`.
- TailA parity contract added in commit `51375d7`.
- Structured runtime foundations added through commits `03d2423`, `6d29e8b`, `b0b9e35`, `a28d34e`, `d425d6d`, and `7fa4582`.
- CLI now checks programs before execution for undefined names, invalid operators and comparisons, non-boolean conditions, mismatched equality, and mismatched conditional branches.
- Verified locally with `go test ./...` and an invalid program that reports diagnostic codes and exits before execution.

## Current interruption state

The working tree is clean. The last pushed commit is `7fa4582` (`add collection iteration`).

## Next action

Implement errors, modules, capabilities, streams, cancellation, and structured
I/O next, keeping the TailA parity contract and Hello World regression intact.

## Documentation sequence

- `docs/README-00-vision.md`
- `docs/README-01-bootstrap.md`
- `docs/README-02-values-and-control.md`
- `docs/README-03-hello-world.md`
- `docs/README-04-diagnostics-and-types.md`
- `docs/README-05-structured-runtime.md`
- `docs/TAILA-PARITY.md`
