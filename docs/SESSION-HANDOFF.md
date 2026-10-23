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
- Relative module imports added in commit `1bfc18f`.
- Capability-scoped file and process effects added in commit `a5e34af`.
- File output and time capabilities added in commit `1467296`.
- Cancellable, buffered, merged event streams added in commit `ed22fc2`.
- CLI now checks programs before execution for undefined names, invalid operators and comparisons, non-boolean conditions, mismatched equality, and mismatched conditional branches.
- Verified locally with `go test ./...` and an invalid program that reports diagnostic codes and exits before execution.

## Current interruption state

The working tree is clean after the handoff update. The last implementation commit is `ed22fc2` (`add cancellable merged event streams`).

## Next action

Implement capability-scoped processes/files, streams, cancellation, and
structured I/O next, keeping the TailA parity contract and Hello World
regression intact.

## Documentation sequence

- `docs/README-00-vision.md`
- `docs/README-01-bootstrap.md`
- `docs/README-02-values-and-control.md`
- `docs/README-03-hello-world.md`
- `docs/README-04-diagnostics-and-types.md`
- `docs/README-05-structured-runtime.md`
- `docs/TAILA-PARITY.md`
