# README 07 — Model evaluation

## Goal

Measure whether Axiom reduces total work required to produce a correct program.
Raw source token count is insufficient: a compact program that requires three
repair loops can cost more than a longer program accepted on the first attempt.

## Variants

The first benchmark should compare:

- Axiom current syntax;
- Axiom canonical compact syntax;
- an ASCII-sigil experiment;
- short semantic machine labels;
- a structured Axiom AST/IR representation;
- Go or TypeScript as established-language baselines;
- Vera as the closest AI-targeted baseline.

The same semantic tasks must be represented in every variant. Hand-porting
individual examples is not sufficient because it introduces author bias.

## Task tiers

1. lexer/parser completion;
2. local expression and function generation;
3. type and effect repair;
4. module and dataflow changes;
5. capability-scoped infrastructure tasks;
6. streaming, cancellation, retry, and fallback workflows;
7. multi-step repository changes.

Each task needs executable tests and a canonical expected result. Tasks should
be generated from a semantic specification, then rendered into each syntax.

## Metrics

Report paired results per model, task, and temperature:

- valid parse rate;
- static-check rate;
- first-pass test pass rate;
- pass@k;
- repair turns;
- total generated tokens;
- total context tokens;
- diagnostic tokens;
- wall-clock latency;
- inference cost;
- syntax/type/effect/name/semantic error counts;
- tool-call and capability violations.

The primary metric is **correct pass@1 per total token and repair turn**. Token
counts must use the actual tokenizer for each model. Lexer tokens and bytes are
useful reproducible secondary metrics only.

## Required ablations

- formatted versus canonical compact source;
- English keywords versus ASCII alternatives;
- named bindings versus structural references;
- unconstrained decoding versus grammar-constrained decoding;
- syntax-only checking versus syntax plus type checking;
- plain diagnostics versus structured repair diagnostics;
- model operations as ordinary functions versus typed effects.

## Reproducibility

Record model identifier, provider, tokenizer, prompt, system instructions,
temperature, seed where supported, maximum output, tool policy, compiler commit,
corpus commit, and every repair attempt. Do not claim a syntax improvement from
one model or one tokenizer.
