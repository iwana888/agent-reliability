// Command modify shows the Runtime's key differentiator: instead of rejecting a
// test-file edit, it MODIFYs the action to point at the real source unit. The
// agent's intent (improve coverage of auth) is preserved; the target is safe.
//
//	go run ./examples/modify
package main

import (
	"context"
	"fmt"

	"github.com/iwana888/agent-reliability"
	"github.com/iwana888/agent-reliability/agentruntime"
)

func main() {
	runtime := agentruntime.New().UseReliability(agentreliability.NewGuard())

	action := agentruntime.Action{Tool: "write_file", Target: "tests/test_auth.pas"}

	dec := runtime.Check(context.Background(), action)
	if dec.Type != agentruntime.MODIFY || dec.Suggested == nil {
		fmt.Println("UNEXPECTED: test edit was not redirected")
		return
	}
	fmt.Printf("Agent:        write_file(%s)\n", action.Target)
	fmt.Printf("Runtime:      MODIFY\n")
	fmt.Printf("Safe instead: write_file(%s)\n", dec.Suggested.Target)
	fmt.Printf("Reason:       %s\n", dec.Reason)
	fmt.Println("-> execute the suggested action, not the original")
}
