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
- CLI now checks programs before execution for undefined names, invalid operators and comparisons, non-boolean conditions, mismatched equality, and mismatched conditional branches.
- Verified locally with `go test ./...` and an invalid program that reports diagnostic codes and exits before execution.

## Current interruption state

The working tree is clean. The last pushed commit is `ac94740` (`add static type checking and diagnostics`).

## Next action

Add source spans to AST expressions and diagnostics so errors identify the exact offending source location. Then extend the type representation as the grammar grows.

## Documentation sequence

- `docs/README-00-vision.md`
- `docs/README-01-bootstrap.md`
- `docs/README-02-values-and-control.md`
- `docs/README-03-hello-world.md`
- `docs/README-04-diagnostics-and-types.md`
