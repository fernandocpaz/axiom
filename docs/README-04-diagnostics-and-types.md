# README 04 — Diagnostics and static types

This milestone adds a compiler check between parsing and execution.

## Implemented

- machine-readable diagnostics with stable error codes;
- static checking for integer, boolean, and string expressions;
- undefined-name, operator, comparison, condition, equality, and branch-type checks;
- CLI validation before the interpreter runs.

The interpreter remains responsible for runtime failures such as division by zero. Static checking is intentionally small and grows alongside the language grammar. Typed function parameters, named record declarations, generic `Result<T,E>` signatures, declared effects, and unknown-type diagnostics are now checked before execution.

```axiom
type Label=record{value:string}
fn classify(text:string)->Result<Label,string> effects(model.infer){return ok(record{value:text})}
```

## Example

```axiom
fn main(){print true+1}
```

The CLI reports `operator-type` and exits without executing the program.

## Next direction

Diagnostics should gain source spans, and the parser should preserve enough syntax information to point at the exact offending expression. After that, the type representation can grow beyond the initial primitive types.
