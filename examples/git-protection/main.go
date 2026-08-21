// Command git-protection shows the Runtime blocking a forced push to a
// protected branch.
//
//	go run ./examples/git-protection
package main

import (
	"context"
	"fmt"

	"github.com/iwana888/agent-reliability"
	"github.com/iwana888/agent-reliability/agentruntime"
)

func main() {
	runtime := agentruntime.New().UseReliability(agentreliability.NewGuard())

	action := agentruntime.Action{Tool: "git", Target: "push --force origin main"}

	dec := runtime.Check(context.Background(), action)
	if dec.Type != agentruntime.DENY {
		fmt.Println("UNEXPECTED: force push was not denied")
		return
	}
	fmt.Printf("Agent:  git push --force origin main\n")
	fmt.Printf("Runtime: DENY\n")
	fmt.Printf("Reason: %s\n", dec.Reason)
	fmt.Println("-> branch protected")
}
