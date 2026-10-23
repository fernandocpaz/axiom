# README 03 — Hello world

This milestone makes the language visibly executable for a first-time user.

## Implemented

- double-quoted string literals;
- common escapes: `\\n`, `\\t`, `\\"`, and `\\\\`;
- string values in the interpreter;
- string equality with `==` and `!=`;
- `examples/hello.ax` now prints exactly `Hello, world!`;
- lexer, parser, and runtime tests for strings.

## Run it

```sh
go run ./cmd/axiom examples/hello.ax
```

Output:

```text
Hello, world!
```

## Exit criteria

The beginner path is now verified end-to-end: source file → lexer → parser → interpreter → terminal output. Future compiler work can build on this executable language core.
