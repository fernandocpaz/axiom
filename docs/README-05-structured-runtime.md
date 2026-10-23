# README 05 — Structured runtime foundations

This milestone moves Axiom beyond scalar expressions toward the data model
needed by TailA.

Implemented:

- user-defined functions with parameters and returns;
- records and lists;
- field access and indexing;
- mutable bindings with explicit `set`;
- `while`, `for ... in`, `break`, and `continue`;
- unary and short-circuit logical operators;
- string concatenation;
- pure builtins for length, conversion, text predicates, and regular-expression
  matching;
- relative module imports with cycle detection;
- explicit `allow` declarations for capability-scoped file reads and process
  execution and file writes;
- UTC timestamps and capability-scoped sleeping for timed workflows;
- static checking and runtime tests for these features.

Example:

```axiom
fn render(pod) {
  return pod.name + " restarts=" + text(pod.restarts)
}

fn main() {
  let pods = [
    record { name: "api", restarts: 2 },
    record { name: "worker", restarts: 0 }
  ]
  for pod in pods {
    print render(pod)
  }
}
```

These are language foundations, not TailA parity by themselves. External
effects, cancellation, streams, terminal capabilities, Kubernetes access,
structured output, AI, and MCP remain required milestones in the parity
contract.
