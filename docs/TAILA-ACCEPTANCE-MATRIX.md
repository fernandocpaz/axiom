# TailA acceptance matrix

This is the executable-product acceptance contract for TailA, the Axiom
migration of Tailg. Tailg remains the reference behavior: a different Axiom
implementation is encouraged, but a user-visible capability or important
failure behavior may not be removed.

Reference inspected:

- repository: `github.com/fernandocpaz/tailg`
- checkout: `d9d3ae737200c484164889762c87f0a348d35a64`
- date inspected: 2026-10-23
- primary sources: `README.md`, Cobra command definitions, Kubernetes/core/TUI/
  agent/AI tests, and test fixtures

The local Go tool could not run Tailg's CLI because the checkout declares
`go 1.25.0` and the installed tool rejects that version format. Command
surface rows below were cross-checked from the Cobra definitions and README;
runtime acceptance must still run the reference and TailA with a compatible
toolchain.

## How to read this matrix

Every row is required. `Acceptance oracle` means the test must compare TailA
with Tailg where the behavior is user-visible, or compare both against a
recorded fixture where the behavior is environmental. `Partial`, `unknown`,
and `unavailable` are explicit results, never implicit success.

Suggested status values for implementation tracking are `not-started`,
`in-progress`, `implemented`, and `verified`; this document deliberately does
not claim that an unimplemented TailA row is complete.

## 1. CLI commands and invocation contract

| ID | Surface | Acceptance criteria | Acceptance oracle |
| --- | --- | --- | --- |
| CLI-001 | root with no target | Resolve active kube context/namespace and open the application picker. | Interactive fixture; no target must not mean wildcard `*`. |
| CLI-002 | positional target | Accept Kubernetes resource targets, application names, wildcard application patterns, and comma-separated applications. | Same target resolves to the same workload/pod scope. |
| CLI-003 | positional namespace | Accept `tailg TARGET NAMESPACE` compatibility form. | Namespace scope and conflict errors match. |
| CLI-004 | `version` | Print build version/commit/time and exit without Kubernetes discovery or TUI startup. | Captured stdout, stderr, exit code, and no side effects. |
| CLI-005 | `update` | Check latest stable release, optionally install it, verify the release digest/version, replace the executable, and report restart behavior. | Fake release server and executable fixture. |
| CLI-006 | `update --check` | Report installed/latest versions without modifying the executable. | File hash and output comparison. |
| CLI-007 | `issues` | Produce bounded grouped issues for scripts/agents; default output is machine-readable JSON. | Schema, redaction, limits, and exit code comparison. |
| CLI-008 | `diagnose` | Produce bounded Kubernetes diagnostics with pod health, logs, crashes, events, recommendations, and network evidence by default. | JSON/NDJSON/text fixture comparison. |
| CLI-009 | `troubleshoot` | Alias `diagnose` and default to human-readable text. | Command alias and output mode comparison. |
| CLI-010 | `issue ISSUE_ID` | Return bounded context for one stable issue ID; fail clearly when the ID is absent from the collection window. | Dynamic-value-normalized ID fixture. |
| CLI-011 | `monitor` | Require explicit context, namespace, and state path; poll once or continuously and emit one JSON monitor record per poll. | State-file and stream comparison. |
| CLI-012 | `monitor --once` | Perform exactly one poll, persist state/cursor, emit one record, and exit. | Process lifetime and state diff. |
| CLI-013 | `mcp` | Serve MCP over stdio with modern discovery and legacy initialize compatibility. | JSON-RPC transcript comparison. |
| CLI-014 | root `--network` | Review the selected namespace without launching the log TUI when structured output is requested. | Network fixture and terminal/non-terminal matrix. |
| CLI-015 | root `--status` | Scan namespace health and recent errors; accept an optional positive whole-minute lookback. | Valid/invalid argument table and status fixture. |
| CLI-016 | root `--versions` | List image/tag/runtime IDs for all container types and support comparisons. | Image-version fixture and comparison table. |
| CLI-017 | `--dump` | Create a navigable troubleshooting bundle containing README, manifest, index, cluster context, events, pod state, YAML, descriptions, and current/previous logs. | Bundle manifest and archive/content comparison. |
| CLI-018 | `--deployment-dump` / `--deploy-dump` | Create deployment-focused bundle with rollout status/history; aliases have equivalent behavior and conflicting dump modes fail. | Path/content/validation comparison. |
| CLI-019 | unsupported mode combinations | Reject invalid combinations before Kubernetes calls, including network root output without `--network`, repeated comparison flags in normal mode, target with status/versions, and split plus tile panes. | Error text class, exit code, and zero API-call assertion. |

## 2. Root flags and configuration

| ID | Flag | Acceptance criteria |
| --- | --- | --- |
| FLAG-001 | `--namespace`, `-n` | Select one namespace normally; repeat only for `--versions` same-context comparison; conflict with positional namespace is rejected. |
| FLAG-002 | `--context` | Select kubeconfig context; repeat only for `--versions` cross-context comparison. |
| FLAG-003 | `--selector` | Override workload label selection without changing displayed scope semantics. |
| FLAG-004 | `--container` | Apply a regular expression to container names; invalid regex fails before collection. |
| FLAG-005 | `--tail` | Set initial lines per container; `-1` explicitly requests all retained lines. |
| FLAG-006 | `--since` | Accept relative windows such as `30m`, `2h`, `4d`; day units normalize for kubectl and do not make startup unbounded. |
| FLAG-007 | `--buffer-lines` | Retain a positive bounded live buffer; zero/negative values fail validation. |
| FLAG-008 | `--refresh-seconds` | Change inventory refresh cadence in live and plain modes. |
| FLAG-009 | `--heartbeat-window` | Change F5 heartbeat grouping window. |
| FLAG-010 | `--include` | Repeatable regex allow filter; all include rules and interaction with complete-history search match Tailg. |
| FLAG-011 | `--exclude` | Repeatable regex deny filter; matching and failure behavior match Tailg. |
| FLAG-012 | `--no-default-exclude` | Restore health/ready/live probe traffic, including probe failures. |
| FLAG-013 | `--detail` | Retain trailing structured JSON properties in compact rows while preserving explicit response bodies. |
| FLAG-014 | `--no-follow` | Load bounded recent logs and exit without live follow. |
| FLAG-015 | `--live-filter`, `--no-live-filter` | Select full-screen filter UI or direct stream; explicit no-live-filter disables the UI. |
| FLAG-016 | `--show-pod`, `--no-show-pod` | Force pod labels visible/hidden; using both is rejected. |
| FLAG-017 | `--split-panes` | On Windows Terminal open one pane per resolved pod; preserve inherited scope and shared filter state. |
| FLAG-018 | `--tile-windows` | On Windows open/tile one window per pod; conflict with split panes is rejected. |
| FLAG-019 | `--output`, `-o` | With root `--network`, accept `json`, `ndjson`, or `text`; reject invalid/unscoped use. |
| FLAG-020 | `--max-bytes` | Bound encoded structured network output, including network evidence inside diagnose. |
| FLAG-021 | `--status-interval` | Change status recovery polling interval. |
| FLAG-022 | `--status-timeout` | Bound status recovery wait. |
| FLAG-023 | `--ai-repo` | Repeatable explicit local source roots for opt-in AI explanations; paths are validated and read-only. |
| FLAG-024 | `--theme` | Accept `auto`, `light`, and `dark`; invalid values fail before kubectl. |
| FLAG-025 | `--no-color` | Disable ANSI colors without changing content or interactive layout. |
| FLAG-026 | `--no-intro` | Skip the interactive Road Runner intro only. |
| FLAG-027 | `--no-update-check` | Suppress startup release checks; machine commands/noninteractive output suppress them by default. |
| FLAG-028 | hidden trace/filter flags | Preserve internal `--trace-pod`, `--trace-selector`, and `--filter-file` semantics for child panes and shared filters. |

## 3. TUI modes, navigation, and interaction

| ID | Mode/interaction | Acceptance criteria |
| --- | --- | --- |
| TUI-001 | startup | Interactive picker uses alternate screen, optional animation, adaptive palette, and restores the shell on exit. Any key dismisses the intro without accidentally activating a picker command. |
| TUI-002 | application picker | Show newest applications, all CronJobs, selection state, context/namespace, responsive columns, workload summary, and scrollable help. |
| TUI-003 | picker navigation | Up/Down, PageUp/PageDown, Home/End, resize, and selection preservation behave consistently. |
| TUI-004 | picker actions | `Enter` opens highlighted/checked applications; `Space` marks; `P` enters exact-pod mode; `R` refreshes; `C` chooses context; `N` chooses namespace; `?` opens help; `Q`/Ctrl+C exits. |
| TUI-005 | context modal | Show frequent choices and More; retain parent state on cancel; reject stale or late replies after scope changes. |
| TUI-006 | namespace modal | List namespaces, support typed `/` entry when listing is forbidden, and preserve context explicitly. |
| TUI-007 | exact-pod picker | Select pods across applications; `Space` toggles; `A` toggles the highlighted app; `Enter` opens pinned pods; cancel preserves monitor and picker state. |
| TUI-008 | CronJob modal | Show schedule/timezone/suspension/active jobs; `X` opens confirmation; Cancel is initially focused; `Y` or selecting Run now plus Enter confirms; too-small terminals cannot confirm. |
| TUI-009 | CronJob podless behavior | A podless CronJob cannot fall through to unrelated namespace logs; show guidance to run it. |
| TUI-010 | namespace Changes view | `H` opens retained changes; collect all workload types, retain observation timestamps, and distinguish incomplete sources from deletions. |
| TUI-011 | namespace Events view | `E` shows Normal/Warning events newest-first with reason/object/count/first/last/source/message. |
| TUI-012 | activity controls | Activity Up/Down/Page/Home/End, `/` filter, Enter details, Ctrl+U clear, W warning toggle, R refresh, H/E switch, Esc back; refreshes serialize and stale replies are ignored. |
| TUI-013 | live log view | Show context, pod/container source, connection state, receive age, log age, date separators, severity formatting, and responsive wrapping without truncation. |
| TUI-014 | live filter | Typing searches retained buffer and complete Kubernetes history; matching-only mode, clear behavior, pause/resume, and shared filter synchronization match. |
| TUI-015 | live keys | F1 matching-only; F2 mapped resources; F3 retained errors; F4 freshness; F5 heartbeat; F6 trace; F7 replica explanation; F8 Issue Radar; Up/Down/Page/Home/End; Enter detail/copy; Esc close; Ctrl+C/Ctrl+Q exit. |
| TUI-016 | detail inspector | Reload full selected entry and adjacent exception/stack frames from Kubernetes, preserve buffered text on failure, wrap/page huge lines, and copy complete off-screen content. |
| TUI-017 | mapped resources | Inspect ConfigMaps/Secrets mapped into pods; do not fetch Secret values until explicitly opened; scroll full documents; Enter copies; mouse capture is opt-in and turns off on exit. |
| TUI-018 | retained errors | F3 reads all retained ERR/FATAL logs across selected containers, groups by local date, retains stack blocks, reports partial failures, supports reload/cancel/copy, and does not apply live tail/display filters. |
| TUI-019 | Issue Radar | Group operational signals, normalize dynamic IDs, mark NEW/KNOWN relative to a session baseline, detect slow requests over 250ms, and support B clear plus C reset. |
| TUI-020 | trace view | F6 searches same trace across selected workloads, sorts by Kubernetes time, preserves raw fields, reports missing/timeout/truncated coverage, and supports original-log navigation/reload/close. |
| TUI-021 | heartbeat view | F5 groups heartbeat failures, open breakers, restarts, and gaps without treating normal heartbeat activity as an issue. |
| TUI-022 | replica view | F7 explains verified controller chain, desired/current/ready counts, rollout evidence, HPA range/conditions, and restricted-RBAC incompleteness. |
| TUI-023 | status UI | Interactive status table supports adaptive colors, wrapped narrow layout, row selection/mouse click, opening an error into the normal log view, and Esc return. |
| TUI-024 | network UI | Interactive namespace network review provides tabs/filtering/events, serialized refresh, responsive detail, and strips terminal control sequences from output. |
| TUI-025 | update modal | Startup update prompt is deferrable, does not steal an existing modal, fits small terminals, and Ctrl+C from log update exits correctly. |
| TUI-026 | AI setup/detail | A opens masked setup when unconfigured; S reopens setup; A detail preserves original log; stale responses are ignored; Esc unwinds nested views. |

## 4. Kubernetes inventory and log behavior

| ID | Area | Acceptance criteria |
| --- | --- | --- |
| KUBE-001 | scope | Every request uses the selected context and namespace explicitly; switching scope clears checked apps and rejects late responses from the prior scope. |
| KUBE-002 | inventory | Discover Deployments, StatefulSets, DaemonSets, ReplicaSets, Pods, Jobs, CronJobs, and namespace events; partial list failures preserve usable sources and prior data. |
| KUBE-003 | workload identity | Resolve verified owner chains; use pod-name fallback with contextual error; never substitute an unrelated workload. |
| KUBE-004 | refresh | Picker refreshes every five seconds by default, preserves surviving highlights/selections, adds new workloads, and prunes only deleted selections. |
| KUBE-005 | replicas | Follow one verified replacement per rollout slot for owned workloads; pin standalone and Job pods; do not guess ownership. |
| KUBE-006 | runtime identity | Detect pod UID, container ID, restart count, and image/runtime changes even when pod name remains stable; restart the affected stream only. |
| KUBE-007 | initial logs | Load up to 500 visible lines/container by default; progressively scan older lines when filters remove data, bounded by history, explicit tail, and buffer. |
| KUBE-008 | follow | Stream all selected pods/containers concurrently with bounded channels and backpressure; batch redraws without reordering records. |
| KUBE-009 | reconnect | Reconnect with capped backoff near the last Kubernetes timestamp; fetch missed records beyond initial tail; preserve independent cursors. |
| KUBE-010 | replay | Suppress duplicate timestamp/message occurrences with bounded per-container replay identity memory; preserve undated/untracked records. |
| KUBE-011 | lifecycle | Clean EOF for completed CronJob/Job is COMPLETED and not an issue; clean EOF on active workloads is REATTACHING; transport errors are RECONNECTING and create stream evidence. |
| KUBE-012 | retention | Handle log rotation and missing history as partial/incomplete evidence rather than inventing records. |
| KUBE-013 | parsing | Parse plain, JSON, Serilog/HTTP fields, severity aliases, trace IDs, status, method, path, and duration; preserve explicit fields over nested payload values. |
| KUBE-014 | filters | Exact-match level/method/trace, case-insensitive substring service/pod/path, comparison operators for status/duration, unit-aware durations, and free-text terms. Invalid values fail before search. |
| KUBE-015 | probe filtering | Hide routine `health|ready|live` probes by default but retain probe WARN/ERROR/FATAL; search/error history must still find retained failures. |
| KUBE-016 | long data | Preserve blank lines, huge lines/payloads, repeated messages, exception blocks, and full SQL/stack traces without fixed per-line truncation. |
| KUBE-017 | errors | F3 reads every retained container, classifies ERR/FATAL aliases and adjacent stack frames, and reports per-source failures alongside successful results. |
| KUBE-018 | network sources | Review Ingress, HTTPProxy, Services, Endpoints, EndpointSlices, NetworkPolicies, pods, and related events; scope HTTPProxy outside references safely and never infer allow/deny without evidence. |
| KUBE-019 | metrics | Poll Metrics API every 15 seconds when monitoring is enabled; aggregate CPU/memory without replica double-counting; show limits/requests/utilization. |
| KUBE-020 | metric gaps | Missing Metrics Server, RBAC denial, missing replica coverage, or pod-spec denial render unavailable `—`, not zero and not an application issue. |

## 5. Monitoring and persistent incidents

| ID | Area | Acceptance criteria |
| --- | --- | --- |
| MON-001 | picker monitoring | `M` toggles monitoring off by default; scans all containers in displayed apps immediately, then previous-minute logs every 60 seconds; max four workers and 20-second scan timeout. |
| MON-002 | health classification | Clean gray/green, warning orange, error red, and five-or-more errors/minute flashing red match replica aggregation and trend rules. |
| MON-003 | silence | A pod whose latest event is error/fatal stays red for three minutes, then becomes unknown/possibly unhealthy; later non-error activity clears pending silence; exception frames are not new entries. |
| MON-004 | scan failures | Failed reads preserve last observation and display scan failure; they do not falsely clear a pending error. |
| MON-005 | lifecycle | Restarting monitor, changing scope, replacing a pod, opening logs, quitting, or leaving picker cancels/resets the correct state; exact-pod mode shares lifecycle. |
| MON-006 | deployment activity | Observe rollout/image/configuration/scale/restart/CronJob changes with Kubernetes timestamps distinct from observation time; retain newest 1,000 session changes and bounded toasts. |
| MON-007 | persistent state | State file survives process restarts, has a replayable cursor, rejects corruption safely, and serializes writers with a lock. |
| MON-008 | incident identity | Incidents key by workload; unchanged snapshots are silent; windows are snapshot-based; fresh evidence is required to reopen after resolution. |
| MON-009 | coverage | Complete/partial/unavailable coverage is emitted; partial scans pause resolution and unknown health is never treated as healthy. |
| MON-010 | resolution | Incident resolves only after complete quiet period and confirmed workload absence/healthy evidence according to Tailg semantics; collection gaps do not resolve it. |
| MON-011 | changes API | `after` cursor and limit pagination are validated; next cursor is stable and replayable. |

## 6. AI, repository context, and MCP

| ID | Area | Acceptance criteria |
| --- | --- | --- |
| AI-001 | opt-in | No provider request occurs without configured key/model and explicit user action. |
| AI-002 | setup wizard | Mask key input; support OpenAI-compatible provider/base URL/model; test with `Reply OK`; save only after confirmation; apply settings immediately; Esc saves nothing. |
| AI-003 | settings | Save owner-only settings in the user config location; reject unsafe permissions/symlinks; never put secrets in repository files or argv. |
| AI-004 | endpoint safety | Require HTTPS except loopback HTTP; reject malformed URLs, URL credentials, query, and fragment. |
| AI-005 | repository selection | Support explicit local folders, discovered Git checkouts, service-scoped mappings, and optional HTTPS/SSH clone using advertised symbolic HEAD; never reset/pull existing clones. |
| AI-006 | source isolation | A service only reads its namespace/service mapping; missing/disappeared mapping prompts again; log-only mode sends no source. |
| AI-007 | bounded context | Read-only source scan excludes secrets/config/dependency/build/symlink paths and bounds files, bytes, and file count; report matched files. |
| AI-008 | redaction | Redact selected log credentials and sensitive values best-effort; clearly state model output is hypothesis, not verified cause. |
| AI-009 | usage | Track provider tokens, optional local price estimate, local session allowance, request count, and rate-limit headers without pretending to know provider account credit. |
| AI-010 | budget failure | Fail closed before a request when worst-case request exceeds remaining allowance; if provider omits token usage, future budget accounting fails closed. |
| AI-011 | provider errors | Do not echo response bodies that may contain secrets; report status/config/quota class only; ignore stale/cancelled responses. |
| MCP-001 | discovery | Advertise `tailg_list_issues`, `tailg_network_diagnose`, `tailg_diagnose`, `tailg_get_issue_context`, `tailg_monitor`, `tailg_list_incidents`, `tailg_get_changes`, `tailg_capture_issue_evidence`, `tailg_read_evidence_page`, and optional Azure DevOps work-item tool. |
| MCP-002 | scope defaults | Tool arguments override server flags; server flags override active kubeconfig; persistent monitor requires explicit state/context/namespace defaults. |
| MCP-003 | monitor resume | Return structured monitor result with cursor/coverage/events and support `nextCursor` continuation. |
| MCP-004 | evidence | Capture issue evidence only into configured private directory; page bounded evidence with cursor/page bytes; verify issue/workload identity. |
| MCP-005 | Azure DevOps | Require configured state and token environment; create/reuse by incident ID; reject wrong/resolved evidence; attach large evidence via chunked upload; do not expose server error bodies. |
| MCP-006 | untrusted data | Treat logs, resource text, events, and model/source excerpts as data, never as instructions. |

## 7. Bundles, structured output, and automation

| ID | Area | Acceptance criteria |
| --- | --- | --- |
| OUT-001 | schema | JSON and NDJSON use versioned `tailg.ai/v1`-compatible shapes with scope, summary, coverage, issues, pods, collection errors, and recommendations. |
| OUT-002 | formats | `issues`/`diagnose` support JSON, NDJSON, and explicit human text where documented; JSON/NDJSON never launch the TUI. |
| OUT-003 | NDJSON | Emit independently parseable records, including summary and coverage; preserve network as a separate record when enabled. |
| OUT-004 | network schema | Network report includes resource IDs, routes, services, endpoints, policies, findings, events, coverage, truncation, and evidence-based correlations. |
| OUT-005 | shared limits | `--max-lines`, `--max-issues`, context lines, timeout, and `--max-bytes` cap the whole report; summary/counts remain truthful when detail is truncated. |
| OUT-006 | coverage | `complete`, `partial`, `unknown`, `unavailable`, and `truncated` are distinguishable; `checks_passed` never implies live connectivity. |
| OUT-007 | exit codes | Automation-compatible result: 0 healthy, 1 warnings, 2 errors/unhealthy, 3 collection/output failure; validation errors retain their documented failure class. |
| OUT-008 | redaction | Common bearer/JWT/password/API-key/secret forms, quoted credentials, and Basic/Bearer headers are redacted before output; Kubernetes Secret values are never fetched. |
| OUT-009 | bundle contents | Every bundle contains `README.md`, `manifest.json`, navigable `index.html`, context/namespace metadata, events, pod/resource evidence, descriptions, and requested logs. |
| OUT-010 | deployment bundle | Deployment bundle adds rollout status/history and retains bounded/incomplete annotations. |
| OUT-011 | plain output | Redirected output is stable ASCII without ANSI control sequences; `--no-color` changes presentation only. |
| OUT-012 | deterministic comparison | Sort records/rows and normalize timestamps, dynamic IDs, paths, and build metadata for differential tests without hiding semantic differences. |

## 8. Edge-case and failure matrix

| ID | Condition | Required result |
| --- | --- | --- |
| EDGE-001 | kubectl absent | Clear error and nonzero exit before attempting TUI/Kubernetes work. |
| EDGE-002 | expired credentials | Picker remains usable; show kubectl error; `C` can change context and `R` retries. |
| EDGE-003 | forbidden list | Preserve successful inventory; show incomplete source; namespace picker permits manual entry. |
| EDGE-004 | partial inventory | Never fabricate deletion/removal; retain prior data and visible coverage error. |
| EDGE-005 | all streams fail | Mark collection unavailable; do not report healthy or resolve incidents. |
| EDGE-006 | one container fails | Keep other containers/results, identify failed source, and mark partial coverage. |
| EDGE-007 | cancellation | Cancel HTTP/kubectl/stream/search/AI/bundle work; ignore late results; do not leak goroutines or overwrite newer state. |
| EDGE-008 | timeout | Preserve sorted partial data, mark timeout/truncation, and avoid claiming completeness. |
| EDGE-009 | output byte limit | Bound encoded output while keeping aggregate counts and coverage truthful. |
| EDGE-010 | empty namespace | Stable empty report/table; no panic and correct zero counts. |
| EDGE-011 | no pods for CronJob | Keep CronJob visible and require explicit Run now; never open namespace logs. |
| EDGE-012 | completed Job EOF | Mark completed and stop retries; retained logs remain inspectable. |
| EDGE-013 | active workload EOF | Reattach with cursor; no false stream issue unless transport reports an error. |
| EDGE-014 | pod replacement | Follow only verified replacement; preserve unrelated streams and reset replacement-specific history. |
| EDGE-015 | same tag, different image ID | Report image mismatch; runtime ID is authoritative. |
| EDGE-016 | missing image ID | Report unknown, never a match. |
| EDGE-017 | same image ID, different tags | Treat as same runtime image for comparison. |
| EDGE-018 | huge/unbroken log line | Preserve and page/wrap it; no scanner-size truncation. |
| EDGE-019 | blank/repeated/undated logs | Preserve blank lines and repeated messages; bound replay tracking without dropping unknown-timestamp data. |
| EDGE-020 | malformed filter/trace | Reject invalid query/trace before search/API collection. |
| EDGE-021 | invalid theme/regex/duration | Validate before Kubernetes calls and provide actionable diagnostics. |
| EDGE-022 | stale async response | Ignore response after scope/generation/view/request changes. |
| EDGE-023 | tiny terminal | Keep close/cancel controls visible; disable unsafe confirmation; wrap content. |
| EDGE-024 | resize | Preserve selection/scroll anchors and recompute layout without data loss. |
| EDGE-025 | ANSI/untrusted text | Strip control sequences from structured/plain output; render evidence as inert text. |
| EDGE-026 | AI missing/invalid config | Do not call provider; open editable setup or report configuration error. |
| EDGE-027 | AI quota/unauthorized | Show safe provider error, preserve editable settings, and do not expose response body. |
| EDGE-028 | AI cancelled/stale | No save, no usage increment, no stale explanation replacement. |
| EDGE-029 | corrupt monitor state | Fail safely with actionable error; never silently reset persistent incidents. |
| EDGE-030 | concurrent monitor writer | Lock/serialize state updates and preserve cursor/incident consistency. |
| EDGE-031 | Azure DevOps retry/conflict | Reuse incident-tagged work item and recheck attachment after revision conflict. |
| EDGE-032 | Windows/non-Windows | Windows pane/tile behavior is tested separately; other platforms use combined streams without Windows-only calls. |

## 9. Differential test harness requirements

TailA parity is verified by scenario, not by checking that similarly named
functions exist. The harness should provide:

1. Recorded Kubernetes command/API fixtures for inventory, logs, metrics,
   events, network resources, failures, timeouts, rotations, and replacements.
2. A fake clock and deterministic scheduler for refresh intervals, backoff,
   silence windows, monitor resolution, and update checks.
3. Terminal transcripts for normal, narrow, resized, redirected, color, and
   Windows layout sessions, including key/mouse input.
4. AI/MCP fake servers that record JSON-RPC/request bodies while asserting
   redaction, size limits, cancellation, and no-secret leakage.
5. Bundle unpacking and schema validation for every requested dump mode.
6. Normalization only for nondeterministic fields: timestamps, temporary paths,
   process/build metadata, random job suffixes, and provider request IDs.

For each matrix row, the minimum test set is success, cancellation, partial
failure, timeout, stale-result/replacement behavior where applicable, and
bounded-output behavior. A row is `verified` only when the differential
scenario passes for Tailg and TailA and the Axiom implementation has a
regression test or fixture proving the result.

## 10. Release gates

TailA cannot claim Tailg parity until all of the following are true:

- every CLI, flag, TUI, Kubernetes, AI/MCP, bundle, monitoring, output, and
  edge-case row has an owner and status;
- all rows marked environment-dependent have fixtures plus at least one live
  integration run;
- no incomplete/unknown coverage is rendered as healthy or resolved;
- differential transcripts match after approved nondeterministic
  normalization;
- `taila` supports the complete Hello World-to-production path: it can run a
  trivial program, invoke its capability runtime, and execute the full Tailg
  parity test suite through Axiom rather than through a Go compatibility shim.
