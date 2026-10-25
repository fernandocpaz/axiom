# README 06 — AI-native design

## Position

Axiom is a machine-native language. Its primary author and consumer is a model,
not a person typing or reviewing source. The compiler therefore optimizes for
the properties that reduce model failure:

- deterministic grammar;
- one canonical representation;
- static types and effects;
- explicit model, retrieval, tool, and infrastructure operations;
- diagnostics that are instructions and structured data;
- fast parse/check/test/repair loops;
- portable execution.

Token count is a metric, not the definition of success. A shorter program that
needs more repair attempts is less efficient than a longer program that passes
on the first attempt.

## Prior art

Axiom is part of an active design space rather than an isolated invention:

- [Vera](https://github.com/aallan/vera) is the closest public language
  comparison. It uses canonical formatting, structural references, mandatory
  contracts, typed effects, LLM-oriented diagnostics, and a model benchmark.
- [Isu/Sui](https://github.com/TakatoHonda/sui-lang) makes a closed-vocabulary
  intermediate representation the primary artifact and supports canonical
  text/AST round trips and step-level repair.
- [Pel](https://arxiv.org/abs/2505.13453) explores pipes, closures, natural
  language conditions, inter-agent communication, capability control, and
  parallelization for agent orchestration.
- [MoonBit](https://www.moonbitlang.com/blog/moonbit-ai) explores an AI-native
  toolchain with flattened structure and compiler-assisted semantic sampling.
- [NanoLang](https://github.com/jordanhubbard/nanolang) combines LLM-oriented
  generation with mandatory tests, formal semantics, capabilities, effects,
  and a verified runtime.

The Axiom-specific opportunity is to combine this model-facing toolchain with
typed infrastructure operations, streams, cancellation, and multi-model
orchestration for TailA-class workloads.

## Syntax policy

### Keep learned control vocabulary

Common English control words such as `fn`, `let`, `if`, `return`, `for`, and
`while` are normally one token in common code tokenizers and have strong model
priors. Replacing them with arbitrary sigils is not automatically an
optimization.

### Use ASCII symbols only when they remove ambiguity

Do not introduce Unicode mathematical symbols for ordinary syntax. Their token
cost varies by tokenizer, they are weakly represented in code corpora, and they
make constrained decoding more complicated.

### Prefer explicit model operations

The language should grow operations such as:

- `infer`: a typed model call;
- `recall`: typed retrieval;
- `invoke`: a capability-scoped tool call;
- `parallel`: independent work that may be scheduled concurrently;
- `retry`, `within`, and `fallback`: explicit failure policy;
- `budget`: token, time, or attempt limits;
- `trace`: provenance and decision metadata.

These should begin as typed builtins/effects. A construct becomes a keyword only
when the compiler needs special semantics that a normal function cannot provide.

The first host boundary is implemented as structured builtins:

```axiom
allow"model.infer"
allow"retrieval.read"
allow"tool.invoke"
fn main(){let answer=infer("classify",record{kind:"label"})let hits=recall("docs","streams")let result=invoke("lookup",["axiom"])print answer.label}
```

The runtime accepts explicit `Host` callbacks for these operations. A missing
host is an error, never an implicit network call or fake response. The inference
host must return a structured record; retrieval must return a list; tool output
remains a host-defined structured value.

The CLI can connect `infer` to any OpenAI-compatible endpoint without putting
credentials in source:

```sh
AXIOM_MODEL_URL=http://127.0.0.1:11434/v1 \
AXIOM_MODEL=deepseek-r1:8b \
go run ./cmd/axiom program.ax
```

The schema record uses field names mapped to primitive type names, for example
`record{label:"string",confidence:"int"}`. The adapter rejects missing fields
and wrong types before the value reaches the program.

`axiom schema file.ax` emits `axiom.schema/v1`, including JSON Schema forms for
declared records, lists, and `Result<T,E>`. This is the stable contract a host
can pass to a constrained decoder. `axiom repair --json file.ax` emits
`axiom.repair/v1` diagnostics and a complete canonical replacement action when
the source only needs normalization.

## Representation

Axiom has two machine-facing forms:

```text
compact canonical source ⇄ canonical AST/IR
```

Source is useful for model generation and diagnostics. The AST/IR is useful for
exact patches, constrained decoding, caching, analysis, and backend lowering.
Both forms must be lossless with respect to program semantics.

Human-readable output is a generated projection, not the source of truth.

## Names

Do not remove all semantic labels without measurement. Structural references can
reduce name-reuse and binding-coherence errors, but models also use names as
semantic retrieval cues. The preferred experiment is short stable labels with
compiler-tracked identity, not arbitrary human prose and not anonymous indices
by default.

## Evidence required before syntax changes

Every proposed syntax variant must be evaluated against the same corpus and
models. The minimum report includes:

1. parse success;
2. static-check success;
3. test pass@1;
4. first-pass compile rate;
5. repair turns;
6. total input/output tokens, including diagnostics and retries;
7. latency and cost;
8. semantic, type, name, and syntax error categories.

`axiom bench` reports deterministic source properties and is deliberately
explicit that lexer-token counts are only a proxy. Model-specific BPE counts
must be measured using the tokenizer of each evaluated model.

`axiom ast file.ax` emits the stable `axiom.ast/v1` JSON representation. It is
the preferred input for exact machine inspection, structural diffs, and future
AST patch operations.
