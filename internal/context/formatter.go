package context

import (
	"encoding/json"
	"fmt"

	"github.com/contact-daniel-me/kyver/internal/graph"
)

func PrintContextResult(res *ContextResult, opts ContextOptions) {
	if res.MultipleFound && res.Symbol == nil {
		if opts.JSONOutput {
			printJSON(res)
		} else {
			graph.PrintSymbolResult(&graph.SymbolResult{Candidates: res.Candidates}, false)
		}
		return
	}

	if opts.JSONOutput {
		if res.Callers != nil {
			res.Callers.Symbol = nil
			res.Callers.Candidates = nil
			res.Callers.MultipleFound = false
		}
		if res.Callees != nil {
			res.Callees.Symbol = nil
			res.Callees.Candidates = nil
			res.Callees.MultipleFound = false
		}
		if res.References != nil {
			res.References.Symbol = nil
			res.References.Candidates = nil
			res.References.MultipleFound = false
		}
		if res.Impact != nil {
			res.Impact.Symbol = nil
			res.Impact.Candidates = nil
			res.Impact.MultipleFound = false
		}
		printJSON(res)
		return
	}

	if opts.TreeFormat {
		printTree(res)
		return
	}

	printHuman(res, opts)
}

func printJSON(res *ContextResult) {
	b, _ := json.MarshalIndent(res, "", "  ")
	fmt.Println(string(b))
}

func printTree(res *ContextResult) {
	fmt.Println(res.Symbol.Name)
	
	type treeNode struct {
		name     string
		children []string
	}
	var nodes []treeNode

	if len(res.Methods) > 0 {
		var children []string
		for _, m := range res.Methods {
			children = append(children, m.Name+"()")
		}
		nodes = append(nodes, treeNode{"Methods", children})
	}
	if len(res.Interfaces) > 0 {
		var children []string
		for _, i := range res.Interfaces {
			children = append(children, i.Name)
		}
		nodes = append(nodes, treeNode{"Interfaces", children})
	}
	if res.Callers != nil && res.Statistics.Callers > 0 {
		var children []string
		for _, g := range res.Callers.Groups {
			for _, c := range g.Callers {
				children = append(children, c.Symbol.Name+"()")
			}
		}
		nodes = append(nodes, treeNode{"Callers", children})
	}
	if res.Callees != nil && res.Statistics.Callees > 0 {
		var children []string
		for _, g := range res.Callees.Groups {
			for _, c := range g.Callees {
				children = append(children, c.Symbol.Name+"()")
			}
		}
		nodes = append(nodes, treeNode{"Callees", children})
	}
	if res.References != nil && res.Statistics.References > 0 {
		var children []string
		fileCounts := make(map[string]int)
		var files []string
		for _, g := range res.References.Groups {
			for _, r := range g.References {
				if fileCounts[r.File] == 0 {
					files = append(files, r.File)
				}
				fileCounts[r.File]++
			}
		}
		for _, f := range files {
			if fileCounts[f] > 1 {
				children = append(children, fmt.Sprintf("%s (%d)", f, fileCounts[f]))
			} else {
				children = append(children, f)
			}
		}
		nodes = append(nodes, treeNode{"References", children})
	}

	for i, node := range nodes {
		isLastNode := i == len(nodes)-1
		prefix := "├── "
		if isLastNode {
			prefix = "└── "
		}
		fmt.Printf("%s%s\n", prefix, node.name)

		childPrefix := "│   "
		if isLastNode {
			childPrefix = "    "
		}

		for j, child := range node.children {
			isLastChild := j == len(node.children)-1
			cPrefix := "├── "
			if isLastChild {
				cPrefix = "└── "
			}
			fmt.Printf("%s%s%s\n", childPrefix, cPrefix, child)
		}
		if !isLastNode {
			fmt.Printf("│\n")
		}
	}
}

func printHuman(res *ContextResult, opts ContextOptions) {
	fmt.Printf("Target Symbol\n\n%s\n", res.Symbol.Name)
	fmt.Printf("Package: %s\n", res.Symbol.Package)
	if res.Symbol.Documentation != "" {
		fmt.Printf("\nDocumentation\n\n%s\n", res.Symbol.Documentation)
	}

	if opts.Summary {
		fmt.Printf("\nSummary\n\n")
		fmt.Printf("• Methods: %d\n", res.Statistics.Methods)
		fmt.Printf("• Interfaces: %d\n", res.Statistics.Interfaces)
		fmt.Printf("• Callers: %d\n", res.Statistics.Callers)
		fmt.Printf("• Callees: %d\n", res.Statistics.Callees)
		fmt.Printf("• References: %d\n", res.Statistics.References)
		if res.Impact != nil {
			fmt.Printf("• Impact Severity: %s\n", res.Impact.Severity)
		}
		fmt.Printf("\nLookup Time: %.2f ms\n", res.LookupTimeMs)
		return
	}

	if len(res.Methods) > 0 {
		fmt.Printf("\nMethods\n\n")
		for _, m := range res.Methods {
			fmt.Printf("• %s()\n", m.Name)
		}
	}

	if len(res.Interfaces) > 0 {
		fmt.Printf("\nInterfaces\n\n")
		for _, i := range res.Interfaces {
			fmt.Printf("• %s\n", i.Name)
		}
	}

	if res.Callers != nil && res.Statistics.Callers > 0 {
		fmt.Printf("\nCallers\n\n")
		graph.PrintCallersResult(res.Callers, graph.CallersOptions{}, false)
	}

	if res.Callees != nil && res.Statistics.Callees > 0 {
		fmt.Printf("\nCallees\n\n")
		graph.PrintCalleesResult(res.Callees, graph.CalleesOptions{}, false)
	}

	if res.References != nil && res.Statistics.References > 0 {
		fmt.Printf("\nReferences\n\n")
		graph.PrintReferenceResult(res.References, graph.ReferenceOptions{}, false)
	}

	if res.Impact != nil {
		fmt.Printf("\nImpact\n\n")
		graph.PrintImpactResult(res.Impact, graph.ImpactOptions{}, false)
	}

	fmt.Printf("\nLookup Time: %.2f ms\n", res.LookupTimeMs)
}
