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

## Current interruption state

The Hello World changes are present in the working tree but were not committed when the prior session was interrupted. The last pushed commit is `c42e554` (`add boolean values and conditional expressions`).

## Next action

Run the tests again, commit the Hello World milestone, and push it to `origin/main`. Then continue with structured diagnostics and a static type-checking pass.

## Documentation sequence

- `docs/README-00-vision.md`
- `docs/README-01-bootstrap.md`
- `docs/README-02-values-and-control.md`
- `docs/README-03-hello-world.md`
