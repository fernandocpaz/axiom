# README 00 — Vision

## Problem

Existing languages optimize mainly for human typing. Axiom is not written or read
by people: the producer and consumer is a language model, so the cost model is
token count, ambiguous grammar, context size, correction loops, and compiler
feedback quality.

Short syntax alone is not enough. Axiom measures token efficiency with a named
tokenizer and a reproducible benchmark, and measures correctness through static
checks and tests rather than through human review of source.

## Design principles

1. One obvious syntax for each construct.
2. Small grammar with deterministic parsing.
3. Static checks before execution.
4. Explicit permissions for external effects.
5. Model calls produce typed values, never unvalidated strings by default.
6. Compact and expanded source are lossless views of one program.
7. Portable compilation before custom hardware.
8. Human readability is a non-goal. Source may optimize for tokenization,
   unambiguity, and machine parsing at the expense of human familiarity.
   Oversight comes from generated diagnostics, diffs, and reports.

## Compiler direction

```text
source ⇄ canonical AST/IR → typed/effect IR → AI/dataflow IR → MLIR → LLVM/vendor backend
```

LLVM and MLIR provide the practical route to current CPUs, GPUs, WebAssembly, and future accelerators. Axiom should not invent a processor until real workloads demonstrate that existing targets are insufficient.

The source and AST/IR forms are lossless machine-facing representations. Source
is optimized for model generation and repair; AST/IR is optimized for exact
patches, constrained decoding, caching, analysis, and backend lowering.

## Non-goals for the first version

- replacing Python or Rust immediately;
- promising a fixed percentage of token savings;
- human-readable or hand-maintained source, or syntax designed for hand-editing;
- embedding an opaque autonomous agent in the compiler;
- designing a new CPU before the language and runtime are validated.
