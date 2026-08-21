# Agent Reliability Runtime

**Execution-layer security runtime for AI agents.**

An AI agent emits a tool call. Before it ever runs, the Reliability Runtime
decides what happens — in code, not in a prompt.

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
Tool Execution
```

The agent never sees the rules and cannot bypass them. Rules are code.

## The four decisions

This project is built around exactly four outcomes for every tool call:

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
  │   safe tool call
  ▼
Execute
```

## Minimal demo

```go
import (
    "context"
    "github.com/iwana888/agent-reliability"
    "github.com/iwana888/agent-reliability/agentruntime"
)

runtime := agentruntime.New()
runtime.UseReliability(agentreliability.NewGuard())

decision, _ := runtime.Check(ctx, agentruntime.Action{
    Tool:   "shell",
    Target: "rm -rf /",
})

switch decision.Type {
case agentruntime.ALLOW:
    execute(action)                       // run as-is
case agentruntime.DENY:
    reject(decision.Reason)               // never runs
case agentruntime.ASK:
    requestApproval(decision)             // human in the loop
case agentruntime.MODIFY:
    execute(*decision.Suggested)          // adopt the safe alternative
}
```

### The intuition in one line

```
Agent:  git push --force origin main
Runtime: DENY
Reason: protected branch: main
```

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

## Use (low-level SDK)

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

### One-line integration (`agent.UseReliability`)

Wrap every tool call the agent wants to make in a single safe gate:

```go
rt := agentreliability.NewGuard()
gate := agentreliability.UseReliability(rt)

err := gate.Run(ctx, agentreliability.Action{Tool: "shell", Target: "rm -rf /"},
    func(ctx context.Context) error { return runTheActualTool() })
// err != nil  -> the Runtime blocked it; nothing ran.
```

`DENY` is blocked outright; `ASK` defaults to *denied* (safe) unless you set
`gate.WithApprover(...)`; `MODIFY` never silently rewrites the agent's intent
unless you opt in with `gate.AdoptModify = true`.

### Audit trail

`agentruntime.Runtime.Audit()` returns a record of every check/execute — who
constrained the agent, and why.

## Add your own team rules

```go
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
```

## Validation tool (`collect-feedback`)

Turn real developer answers to *"what do you fear most when an agent writes
code for you?"* into a Policy backlog:

```bash
go run ./cmd/collect-feedback                 # interactive, one fear per line
go run ./cmd/collect-feedback --answers a.txt # batch (text or JSON per line)
```

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

## Use (CLI)

`agentworld-guard` reads one Action as JSON and prints the Decision as JSON.

```bash
echo '{"tool":"shell","target":"rm -rf /"}' | go run ./cmd/agentworld-guard
# {"decision":"DENY","policy":"NO_DELETE_FILES","reason":"File deletion is forbidden by default; restore from VCS instead."}

go run ./cmd/agentworld-guard --catalog          # list enabled policies
go run ./cmd/agentworld-guard --eval-file a.json # read Action from a file
```

---

Built and battle-tested in [AgentWorld](https://github.com/iwana888/AgentWorld).
This package is the part you can take to production without taking AgentWorld
with you.
