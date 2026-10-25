# Axiom

Axiom is an experimental AI-only programming language. Source is written and
consumed by language models: human readability is not a requirement and human
ergonomics are explicitly out of scope. The language optimizes for reliable
generation, deterministic parsing, machine-checkable correctness, and
portability across current hardware.

## AI-only by design

People do not read Axiom source. That has concrete consequences:

- Syntax and identifiers may be optimized for tokenization and unambiguity
  rather than for human familiarity.
- Formatting is canonical and deterministic, emitted by the toolchain instead of
  hand-maintained.
- Correctness is established by static diagnostics, tests, and typed effects,
  not by reading source during review.
- Human oversight happens through machine-produced artifacts: diagnostics,
  diffs, summaries, and audit reports.
- Documentation targets models first; text written for people is secondary.

When a person needs to understand an Axiom program, the answer is generated
output from tooling, not a more readable language.

## Status

The language kernel now includes structured values, user-defined functions,
controlled loops, mutable bindings, text/regex builtins, and source-spanned
static diagnostics. It also runs a real `Hello, world!` program end to end and
exposes explicit host boundaries for model inference, retrieval, and tool calls.

This repository is being developed incrementally. Each milestone has a short design record in [`docs/`](docs/).

## Quick start

Requirements: Go 1.15 or newer.

```sh
go run ./cmd/axiom examples/hello.ax
go test ./...
go run ./cmd/axiom check --json examples/hello.ax
go run ./cmd/axiom bench
go run ./cmd/axiom ast examples/hello.ax
go run ./cmd/axiom schema examples/hello.ax
go run ./cmd/axiom repair --json examples/hello.ax
```

To enable `infer` through a local Ollama or OpenAI-compatible endpoint, set
`AXIOM_MODEL_URL` and `AXIOM_MODEL`. Add `AXIOM_API_KEY` only for endpoints
that require bearer authentication.

`axiom schema` emits the declared JSON Schema catalogue for constrained model
decoding. `axiom repair` emits `axiom.repair/v1` feedback and a deterministic
whole-file canonical replacement when formatting repair is possible.

For model-generation evaluation, pass a JSON task manifest to `axiom bench`
with `--model-url`, `--model`, and optional `--samples`; generated code is
parsed and checked but never executed.

Example:

```axiom
fn main(){print"Hello, world!"}
```

## Direction

Axiom will eventually combine:

- a small canonical grammar for reliable AI generation;
- a canonical formatter and compact machine view, with no human-readability
  requirement;
- static types, algebraic data, and explicit effects;
- typed model, retrieval, tool, and infrastructure operations;
- explicit retry, timeout, fallback, budget, provenance, and parallel effects;
- a compiler pipeline targeting MLIR/LLVM and existing CPU/GPU/WebAssembly targets.

The long-term implementation target is TailA: a full Axiom-native port of the
active Tailg product. See the [TailA parity contract](docs/TAILA-PARITY.md) and
the existing [diagnostics milestone](docs/README-04-diagnostics-and-types.md).

The compiler backend is intentionally deferred until the language semantics are stable. See [`docs/README-00-vision.md`](docs/README-00-vision.md) and [`docs/README-01-bootstrap.md`](docs/README-01-bootstrap.md).

## License

Apache-2.0. This is an early research project; APIs and syntax will change.
