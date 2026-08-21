// Command database-approval shows the Runtime pausing a production database
// write for human approval (ASK).
//
//	go run ./examples/database-approval
package main

import (
	"context"
	"fmt"

	"github.com/iwana888/agent-reliability"
	"github.com/iwana888/agent-reliability/agentruntime"
)

// in a real product this would open a UI / Slack prompt; here we simulate approve.
func main() {
	runtime := agentruntime.New().UseReliability(agentreliability.NewGuard())
	runtime.WithApprover(func(ctx context.Context, a agentruntime.Action, d agentruntime.Decision) bool {
		return false // default: deny; flip to true to simulate approval
	})

	action := agentruntime.Action{Tool: "sql", Target: "UPDATE users SET plan='pro' WHERE id=1 (env: production)"}

	dec := runtime.Check(context.Background(), action)
	if dec.Type != agentruntime.ASK {
		fmt.Println("UNEXPECTED: prod write was not flagged for approval")
		return
	}
	fmt.Printf("Agent:  %s\n", action.Target)
	fmt.Printf("Runtime: ASK\n")
	fmt.Printf("Reason: %s\n", dec.Reason)
	fmt.Println("-> waiting for human approval (default: denied)")
}
