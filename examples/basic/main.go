// Command basic is the smallest possible integration: an agent emits a tool
// call, the Reliability Runtime decides, and you act on the decision.
//
//	go run ./examples/basic
package main

import (
	"context"
	"fmt"

	"github.com/iwana888/agent-reliability"
	"github.com/iwana888/agent-reliability/agentruntime"
)

func main() {
	runtime := agentruntime.New()
	runtime.UseReliability(agentreliability.NewGuard())

	// An agent wants to run a shell command.
	action := agentruntime.Action{Tool: "shell", Target: "rm -rf /"}

	decision := runtime.Check(context.Background(), action)

	switch decision.Type {
	case agentruntime.ALLOW:
		fmt.Printf("ALLOW  -> execute %q\n", action.Target)
	case agentruntime.DENY:
		fmt.Printf("DENY   -> blocked: %s\n", decision.Reason)
	case agentruntime.ASK:
		fmt.Printf("ASK    -> needs human approval: %s\n", decision.Reason)
	case agentruntime.MODIFY:
		fmt.Printf("MODIFY -> suggested instead: %s\n", decision.Suggested.Target)
	}
}
