# README 01 — Bootstrap interpreter

This milestone establishes the repository and an executable language kernel.

## Implemented

- deterministic lexer;
- recursive-descent parser;
- integer literals and arithmetic expressions;
- `let` bindings;
- `print` statements;
- `fn main() { ... }` program wrapper;
- source locations in syntax errors;
- unit tests for lexing, parsing, and execution.

## Current syntax

```axiom
fn main(){let x=2*(20+1)print x}
```

## Deliberate limitations

There is no static type checker, optimizer, module system, model runtime, or native backend yet. Those are separate milestones so each layer can be tested independently.

## Exit criteria

The next milestone should add booleans, comparisons, conditional expressions, and a structured diagnostic format without making the grammar ambiguous.
