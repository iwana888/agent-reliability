// Command file-protection shows the Runtime blocking a destructive file
// deletion before it ever reaches the filesystem.
//
//	go run ./examples/file-protection
package main

import (
	"context"
	"fmt"

	"github.com/iwana888/agent-reliability"
	"github.com/iwana888/agent-reliability/agentruntime"
)

func main() {
	runtime := agentruntime.New().UseReliability(agentreliability.NewGuard())

	action := agentruntime.Action{Tool: "shell", Target: "rm -rf /"}

	dec := runtime.Check(context.Background(), action)
	if dec.Type != agentruntime.DENY {
		fmt.Println("UNEXPECTED: deletion was not denied")
		return
	}
	fmt.Printf("Agent:  %s\n", action.Target)
	fmt.Printf("Runtime: DENY\n")
	fmt.Printf("Reason: %s\n", dec.Reason)
	fmt.Println("-> filesystem untouched")
}
