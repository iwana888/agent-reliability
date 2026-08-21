# Agent Reliability Runtime

**Execution-layer security runtime for AI agents.**

> Prompts tell agents what they should do.
> Runtime controls what they can actually execute.

```
Agent
  │   tool call
  ▼
Agent Reliability Runtime
  │
  ▼
ALLOW / DENY / ASK / MODIFY
  │
  ▼
Execution
```

The Runtime enforces the decision at execution time, independently of the
model's instructions or intent. The agent cannot bypass the execution gate.

## Why?

Prompts can say "don't delete the database." But a prompt is advice — the model
can forget it, be steered around it, or simply be told to ignore it.

```
LLM
  │   "delete the production database"
  ▼
Agent Reliability Runtime
  │   DENY
  ▼
Database stays intact
```

Agent Reliability Runtime sits between an AI agent and actual tool execution
and evaluates every tool call *before* it runs — in code, not in a prompt. The
agent's executor is only ever called for `ALLOW`, approved `ASK`, or an adopted
`MODIFY`. It is never called for `DENY`, and the agent's intent is never
rewritten behind your back.

## Four core decisions

This project is built around four core decisions for tool calls:

| Decision | What happens | Executed? |
|---|---|---|
| `ALLOW`  | run the action as-is | yes |
| `DENY`   | forbidden; never runs | **never** |
| `ASK`    | pause for human approval, then `ALLOW` | only after approval |
| `MODIFY` | Runtime returns a *suggested* action; it does **not** silently rewrite the agent's intent | only if you adopt it |

Most agent safety tooling stops at allow / deny. `MODIFY` is the differentiator:
instead of rejecting a risky call, the Runtime steers it to a safe equivalent and
lets **you** decide whether to adopt it.

```
Agent
  │   wrong tool call
  ▼
MODIFY
  │   safe tool call (suggested)
  ▼
Execute  (only if you adopt it)
```

The intuition in one line:

```
Agent:  git push --force origin main
Runtime: DENY
Reason: protected branch: main
```

## Minimal demo

`agentruntime` is the single entry point. A consumer imports exactly one module.

```go
import (
    "context"
    "github.com/iwana888/agent-reliability/agentruntime"
)

runtime := agentruntime.New()
runtime.UseReliability(agentreliability.NewGuard())

action := agentruntime.Action{Tool: "shell", Target: "rm -rf /"}

decision := runtime.Check(ctx, action)

switch decision.Type {
case agentruntime.ALLOW:
    runtime.Execute(ctx, decision, action, runTheTool)   // run as-is
case agentruntime.DENY:
    // never runs
case agentruntime.ASK:
    if approve(decision) {
        runtime.Execute(ctx, decision, action, runTheTool)
    }
case agentruntime.MODIFY:
    // adopt the safe alternative only if you choose to
    runtime.Execute(ctx, decision, *decision.Suggested, runTheTool)
}
```

> `runTheTool` has the signature `func(ctx context.Context, a agentruntime.Action) error` —
> the agent's real executor. The Runtime only calls it when the decision permits execution.

## Run the examples

Clone the repo, then:

```bash
go run ./examples/basic              # the four-state switch
go run ./examples/file-protection    # rm -rf /            -> DENY
go run ./examples/git-protection     # git push --force    -> DENY
go run ./examples/database-approval  # UPDATE production   -> ASK
go run ./examples/modify             # test_x.pas → src/x  -> MODIFY
```

## Install

```bash
go get github.com/iwana888/agent-reliability
```

## Quick start (one-line integration)

Wrap every tool call the agent wants to make in a single safe gate:

```go
import (
    "context"
    "github.com/iwana888/agent-reliability/agentruntime"
)

runtime := agentruntime.New()
runtime.UseReliability(agentreliability.NewGuard())

// wrap the agent's tool call — do is only invoked when the Runtime permits it
err := runtime.Execute(ctx, runtime.Check(ctx, action), action,
    func(ctx context.Context, a agentruntime.Action) error {
        return runTheActualTool(a)
    })
// err != nil (BlockedError)  -> the Runtime blocked it; nothing ran.
```

`DENY` is blocked outright; `ASK` defaults to *denied* (safe) unless you wire
`runtime.WithApprover(...)`; `MODIFY` never silently rewrites the agent's intent
unless you pass `decision.Suggested` yourself.

## The default policies

Sourced from the things developers fear most when an agent (Claude Code, Codex,
Cursor, …) writes code for them:

| Policy | Triggers on | Decision |
|---|---|---|
| `NO_DELETE_FILES` | `delete_file` / `rm` / `rmdir` | `DENY` |
| `NO_GIT_RESET_HARD` | `git reset --hard` | `DENY` |
| `NO_FORCE_PUSH` | `git push --force` / `-f` | `DENY` |
| `NO_MODIFY_ENV` | editing `.env` / secret files | `DENY` |
| `NO_FORK_BOMB` | recursive/unbounded shell | `DENY` |
| `NO_DROP_TABLE` | `DROP TABLE` / `TRUNCATE` / `DELETE FROM` | `DENY` |
| `NO_READ_SECRETS` | reading a vault/credential file | `ASK` |
| `NO_PROD_DB` | any write to a prod database | `ASK` |
| `NO_PROD_DEPLOY` | deploy to production | `ASK` |
| `NO_MODIFY_TESTS` | writing `test_*.go` / `test_*.pas` | `MODIFY` (redirect to source) |

## Add your own team rules

`agentruntime` re-exports the `Policy` type, so you stay on the single entry point:

```go
rt := agentreliability.NewGuard()
rt.AddPolicy(agentreliability.Policy{
    ID:       "NO_TOUCH_BILLING",
    Name:     "Never touch the billing service",
    Severity: "critical",
    Enabled:  true,
    Eval: func(a agentreliability.Action) (agentreliability.Decision, bool) {
        if a.Tool == "deploy" && strings.Contains(a.Target, "billing") {
            return agentreliability.Decision{
                Type:     agentreliability.Deny,
                PolicyID: "NO_TOUCH_BILLING",
                Reason:   "billing is out of scope",
            }, true
        }
        return agentreliability.Decision{}, false
    },
})
runtime.UseReliability(rt)
```

## Audit trail

`runtime.Audit()` returns a record of every check/execute — who constrained the
agent, and why:

```go
for _, e := range runtime.Audit() {
    fmt.Printf("%s -> %s (ran=%v) %s\n",
        e.Action.Target, e.Decision.Type, e.Ran, e.Decision.Reason)
}
```

## Validation tool (`collect-feedback`)

Turn real developer answers to *"what do you fear most when an agent writes
code for you?"* into a Policy backlog:

```
Developer Fear
      ↓
collect-feedback
      ↓
Policy backlog
      ↓
Runtime rule
```

```bash
go run ./cmd/collect-feedback                 # interactive, one fear per line
go run ./cmd/collect-feedback --answers a.txt # batch (text or JSON per line)
```

## CLI

`agent-reliability-guard` reads one Action as JSON and prints the Decision as JSON.

```bash
echo '{"tool":"shell","target":"rm -rf /"}' | go run ./cmd/agent-reliability-guard
# {"decision":"DENY","policy":"NO_DELETE_FILES","reason":"File deletion is forbidden by default; restore from VCS instead."}

go run ./cmd/agent-reliability-guard --catalog          # list enabled policies
go run ./cmd/agent-reliability-guard --eval-file a.json # read Action from a file
```

## Advanced / low-level API

`agentruntime` is a thin, recommended facade over the core package
`agentreliability`. If you want direct access to the guard without the runtime
facade, use it directly:

```go
import "github.com/iwana888/agent-reliability"

rt := agentreliability.NewGuard()

dec := rt.Check(ctx, agentreliability.Action{
    Tool:   "shell",
    Target: "rm -rf /",
})

switch dec.Type {
case agentreliability.Allow:
    // execute
case agentreliability.Deny:
    // do NOT execute
case agentreliability.Ask:
    // block until a human approves; then execute
case agentreliability.Modify:
    // optionally adopt dec.Suggested instead of the original action
}
```

`agentreliability.UseReliability(rt)` also returns a `*Gate` — the low-level
one-line integration seam — for callers who do not need the `agentruntime`
facade.

---

This project is independent and has zero dependency on AgentWorld. It originated
from patterns battle-tested in [AgentWorld](https://github.com/iwana888/AgentWorld),
and is the part you can take to production on its own.
