# README 00 — Vision

## Problem

Existing languages optimize mainly for human typing. AI-generated code has a second cost model: token count, ambiguous grammar, context size, correction loops, and compiler feedback quality.

Short syntax alone is not enough. Axiom measures token efficiency with a named tokenizer and a reproducible benchmark, while also measuring correctness and human review time.

## Design principles

1. One obvious syntax for each construct.
2. Small grammar with deterministic parsing.
3. Static checks before execution.
4. Explicit permissions for external effects.
5. Model calls produce typed values, never unvalidated strings by default.
6. Compact and expanded source are lossless views of one program.
7. Portable compilation before custom hardware.

## Compiler direction

```text
source → canonical AST → typed/effect IR → AI/dataflow IR → MLIR → LLVM/vendor backend
```

LLVM and MLIR provide the practical route to current CPUs, GPUs, WebAssembly, and future accelerators. Axiom should not invent a processor until real workloads demonstrate that existing targets are insufficient.

## Non-goals for the first version

- replacing Python or Rust immediately;
- promising a fixed percentage of token savings;
- embedding an opaque autonomous agent in the compiler;
- designing a new CPU before the language and runtime are validated.
