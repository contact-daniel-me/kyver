package cli

import (
	"fmt"
	"os"
)

func Execute() error {
	if len(os.Args) < 2 {
		fmt.Println("Kyver CLI - Engineering Knowledge Version Control")
		return nil
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "init":
		return runInit(args)
	case "ask":
		return runAsk(args)
	case "index":
		return runIndex(args)
	case "analyze":
		return runAnalyze(args)
	case "search":
		return runSearch(args)
	case "goto":
		return runGoto(args)
	case "peek":
		return runPeek(args)
	case "outline":
		return runOutline(args)
	case "hierarchy":
		return runHierarchy(args)
	case "callers":
		return runCallers(args)
	case "callees":
		return runCallees(args)
	case "graph":
		return runGraph(args)
	case "impacts":
		return runImpacts(args)
	case "references", "refs":
		return runReferences(args)
	case "implements":
		return runImplements(args)
	case "interface":
		return runInterface(args)
	case "context":
		return runContext(args)
	case "status":
		return runStatus(args)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		return nil
	}
}
