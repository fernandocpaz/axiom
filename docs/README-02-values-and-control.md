# README 02 — Values and control expressions

This milestone extends the executable kernel without introducing grammar ambiguity.

## Implemented

- integer and boolean runtime values;
- `true` and `false` literals;
- equality and inequality: `==`, `!=`;
- ordered comparisons: `<`, `<=`, `>`, `>=`;
- conditional expressions: `if condition { yes } else { no }`;
- precedence-aware expression parsing;
- runtime diagnostics for invalid comparisons, non-boolean conditions, undefined names, and division by zero.

## Example

```axiom
fn main() {
  let score = if 4 >= 2 { 7 } else { 0 }
  print score
  print score == 7
}
```

Output:

```text
7
true
```

## Design note

The current interpreter uses a small tagged `Value` representation so semantics can be tested before committing to a static type-system representation. The next milestone will move these checks into a compiler pass and emit machine-readable diagnostics.
