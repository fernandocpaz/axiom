# Axiom

Axiom is an experimental AI-native programming language for software that is efficient for language models to generate, easy for people to audit, and portable across current hardware.

## Status

Milestone 3 is complete: Axiom can statically check primitive expressions and emit machine-readable diagnostics before running a program. It also runs a real `Hello, world!` program with string literals, alongside arithmetic, boolean values, comparisons, and conditional expressions.

This repository is being developed incrementally. Each milestone has a short design record in [`docs/`](docs/).

## Quick start

Requirements: Go 1.15 or newer.

```sh
go run ./cmd/axiom examples/hello.ax
go test ./...
```

Example:

```axiom
fn main() { print "Hello, world!" }
```

## Direction

Axiom will eventually combine:

- a small canonical grammar for reliable AI generation;
- a human-readable formatter and compact machine view;
- static types, algebraic data, and explicit effects;
- typed model, retrieval, and tool operations;
- a compiler pipeline targeting MLIR/LLVM and existing CPU/GPU/WebAssembly targets.

The long-term implementation target is TailA: a full Axiom-native port of the
active Tailg product. See the [TailA parity contract](docs/TAILA-PARITY.md) and
the existing [diagnostics milestone](docs/README-04-diagnostics-and-types.md).

The compiler backend is intentionally deferred until the language semantics are stable. See [`docs/README-00-vision.md`](docs/README-00-vision.md) and [`docs/README-01-bootstrap.md`](docs/README-01-bootstrap.md).

## License

Apache-2.0. This is an early research project; APIs and syntax will change.
