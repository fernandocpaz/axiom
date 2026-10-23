# TailA parity contract

TailA is the Axiom implementation of Tailg. Tailg is the reference product;
TailA must preserve its user-visible capabilities and important edge-case
behavior even when the implementation and programming model are different.

Reference snapshot: Tailg `v0.1.36`, commit `d9d3ae7`, repository
`github.com/fernandocpaz/tailg`.

## Product surface

Every item below is required for TailA parity.

### Root command

- target and namespace selection, including wildcard and multi-target selection;
- Kubernetes context and namespace comparison;
- pod/container selection and label selector overrides;
- tail count, time windows, buffer limits, refresh cadence, and heartbeat windows;
- include/exclude regular-expression filters and default probe filtering;
- snapshot and follow modes;
- live filtering, pod labels, detail retention, themes, color control, and intro control;
- split panes, tiled Windows Terminal windows, and shared filter state;
- troubleshooting and deployment troubleshooting bundles;
- status scanning, version comparison, and namespace network review.

### Commands

- `version`;
- `update`, including check, download, verification, replacement, and restart;
- `issues`, bounded grouped issue output for scripts and AI agents;
- `diagnose`, bounded Kubernetes diagnostic reports;
- `issue`, bounded context for one issue ID;
- `mcp`, stdio MCP server behavior;
- `monitor`, persistent incident and change monitoring.

### Kubernetes behavior

- discovery across pods, deployments, statefulsets, daemonsets, jobs, and
  cronjobs;
- selectors, verified workload identity, replica inventory, readiness, phases,
  restarts, image versions, and resource details;
- timestamped log snapshots and live streams;
- reconnect, retry backoff, cursor-based replay suppression, stream lifecycle,
  container replacement detection, and completed-job handling;
- events, services, ingress, HTTPProxy, endpoints, EndpointSlices, policies,
  and namespace network relationships;
- bounded collection with explicit incomplete and truncated states.

### Terminal and interaction behavior

- full-screen log browsing and direct streaming;
- filtering, selection, scrolling, paging, home/end navigation, wrapped detail
  views, copying, modal dialogs, pickers, status views, network views, and
  namespace activity views;
- keyboard behavior, adaptive colors, alternate-screen handling, terminal
  resize behavior, and platform-specific behavior.

### AI and agent behavior

- opt-in log explanations with local source repositories;
- provider setup and repair flows;
- bounded evidence collection and redaction;
- structured JSON and NDJSON output with size limits;
- MCP diagnostics and persistent-monitoring tools;
- Azure DevOps work-item creation where configured.

## Axiom implementation strategy

TailA must not be a line-for-line translation of Tailg. Axiom should express
the same product as a network of typed streams, cells, capabilities, and
reactions:

```text
Kubernetes capabilities
        → inventory stream
        → reconciled log streams
        → typed event stream
        → filters and analyzers
        → terminal / JSON / MCP sinks
```

The parity rule is semantic: if Tailg handles a reconnect, selector failure,
partial API, terminal mode, or bounded-output case, TailA must handle the same
case. A different internal architecture is encouraged; a missing behavior is
not.

## Milestones

1. Language kernel: functions, structured values, pattern matching, modules,
   errors, and tests.
2. Runtime kernel: streams, cancellation, bounded buffers, timers, cells, and
   deterministic event testing.
3. Capability layer: processes, files, terminal I/O, JSON, HTTP, and Kubernetes
   command/API access with explicit permissions.
4. TailA core: inventory, log cursors, reconnecting streams, filtering, and
   structured output.
5. TailA terminal: picker, live view, detail views, status, network review, and
   platform integration.
6. TailA operations: bundles, AI, MCP, monitoring, updates, and release parity.
7. Differential validation: run equivalent Tailg and TailA scenarios against
   recorded fixtures and compare outputs, state transitions, and exit codes.

No milestone is complete merely because the happy path works. Each capability
requires fixture tests for success, cancellation, partial failure, retry,
replacement, truncation, and user interruption where applicable.
