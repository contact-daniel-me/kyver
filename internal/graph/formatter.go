package graph

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
)

func PrintSymbolResult(res *SymbolResult, jsonOutput bool) {
	if jsonOutput {
		printJSON(res)
		return
	}
	if len(res.Candidates) > 1 && res.Symbol == nil {
		printMultiple(res.Candidates)
		return
	}
	for _, f := range res.Found {
		fmt.Printf("%s\n", f.Name)
	}
}



func PrintImpactResult(res *ImpactResult, opts ImpactOptions, jsonOutput bool) {
	if jsonOutput {
		printJSON(res)
		return
	}
	if len(res.Candidates) > 1 && res.Symbol == nil {
		printMultiple(res.Candidates)
		return
	}
	if res.Symbol == nil {
		fmt.Println("No symbol found.")
		return
	}

	if opts.TreeFormat {
		printImpactTree(res, opts.SortBy)
		return
	}

	sym := res.Symbol

	fmt.Print("Target Symbol\n\n")
	fmt.Printf("Symbol\n%s\n", sym.Name)
	fmt.Printf("Kind\n%s\n", sym.Kind)
	if sym.Receiver != "" {
		fmt.Printf("Receiver (if applicable)\n%s\n", sym.Receiver)
	}
	fmt.Printf("Package\n%s\n", sym.Package)
	fmt.Printf("File\n%s\n", sym.FilePath)
	fmt.Printf("Line\n%d\n", sym.StartLine)
	if sym.Documentation != "" {
		doc := strings.TrimSpace(sym.Documentation)
		if len(doc) > 80 {
			doc = doc[:77] + "..."
		}
		fmt.Printf("Documentation\n%s\n", doc)
	}

	fmt.Print("\n──────────────────────────────────────\n\n")
	fmt.Print("Severity\n\n")
	fmt.Println(res.Severity)
	if res.SeverityRationale != "" {
		fmt.Print("\nReason\n\n")
		fmt.Println(res.SeverityRationale)
	}

	fmt.Print("\n──────────────────────────────────────\n\n")
	fmt.Print("Direct Callers\n\n")
	if len(res.DirectCallers) > 0 {
		for _, dc := range res.DirectCallers {
			fmt.Println(formatCallerLabel(dc))
		}
	} else {
		fmt.Println("None")
	}

	fmt.Print("\n──────────────────────────────────────\n\n")
	fmt.Print("Indirect Callers\n\n")
	if len(res.IndirectCallers) > 0 {
		for _, ic := range res.IndirectCallers {
			fmt.Println(formatCallerLabel(ic))
		}
	} else {
		fmt.Println("None")
	}

	fmt.Print("\n──────────────────────────────────────\n\n")
	fmt.Print("Affected Files\n\n")
	if len(res.AffectedFiles) > 0 {
		for _, f := range res.AffectedFiles {
			fmt.Println(f)
		}
	} else {
		fmt.Println("None")
	}

	fmt.Print("\n──────────────────────────────────────\n\n")
	fmt.Print("Affected Packages\n\n")
	if len(res.AffectedPackages) > 0 {
		for _, p := range res.AffectedPackages {
			fmt.Println(p)
		}
	} else {
		fmt.Println("None")
	}

	fmt.Print("\n──────────────────────────────────────\n\n")
	fmt.Print("Affected Tests\n\n")
	if len(res.Tests) > 0 {
		for _, tc := range res.Tests {
			fmt.Println(formatCallerLabel(tc))
		}
	} else {
		fmt.Println("None")
	}

	fmt.Print("\n──────────────────────────────────────\n\n")
	fmt.Print("Statistics\n\n")
	fmt.Printf("%-18s: %d\n", "Direct Callers", res.Stats.DirectCallers)
	fmt.Printf("%-18s: %d\n", "Indirect Callers", res.Stats.IndirectCallers)
	fmt.Printf("%-18s: %d\n", "Affected Files", res.Stats.AffectedFiles)
	fmt.Printf("%-18s: %d\n", "Affected Packages", res.Stats.AffectedPkgs)
	fmt.Printf("%-18s: %d\n", "Public APIs", res.Stats.PublicAPIs)
	if res.LookupTimeMs < 1.0 {
		fmt.Printf("%-18s: %.0f µs\n", "Lookup Time", res.LookupTimeMs*1000)
	} else {
		fmt.Printf("%-18s: %.2f ms\n", "Lookup Time", res.LookupTimeMs)
	}
}

func printImpactTree(res *ImpactResult, sortBy string) {
	targetLabel := formatImpactSymbolLabel(res.Symbol)
	fmt.Println(targetLabel)

	if len(res.DirectCallers) == 0 {
		return
	}

	children := make(map[string][]ImpactCaller)
	for _, dc := range res.DirectCallers {
		children[dc.Calls] = append(children[dc.Calls], dc)
	}
	for _, ic := range res.IndirectCallers {
		children[ic.Calls] = append(children[ic.Calls], ic)
	}

	var walk func(parentLabel, prefix string)
	walk = func(parentLabel, prefix string) {
		kids := append([]ImpactCaller(nil), children[parentLabel]...)
		sortImpactCallers(kids, sortBy)
		
		if len(kids) > 0 {
			if prefix == "" {
				fmt.Print("    ↑\n")
			} else {
				fmt.Printf("%s│\n", prefix)
			}
		}

		for i, child := range kids {
			isLast := i == len(kids)-1
			
			var branch string
			if prefix == "" && len(kids) == 1 {
				// Single direct caller, simple arrow
				fmt.Printf("%s\n", formatCallerLabel(child))
				walk(formatCallerLabel(child), "")
				continue
			}
			
			if isLast {
				branch = "└── "
			} else {
				branch = "├── "
			}
			
			fmt.Printf("%s%s%s\n", prefix, branch, formatCallerLabel(child))
			
			// For child of child, if it exists we need to print the up arrow for it
			hasKids := len(children[formatCallerLabel(child)]) > 0
			if hasKids {
				if isLast {
					fmt.Printf("%s      ↑\n", prefix)
					walk(formatCallerLabel(child), prefix+"     ")
				} else {
					fmt.Printf("%s│     ↑\n", prefix)
					walk(formatCallerLabel(child), prefix+"│    ")
				}
			}
		}
	}

	walk(targetLabel, "")
}

func formatImpactSymbolLabel(sym *analyzer.Symbol) string {
	if sym.Kind == "Method" {
		if strings.HasPrefix(sym.Receiver, "*") {
			return fmt.Sprintf("%s.(%s).%s()", sym.Package, sym.Receiver, sym.Name)
		}
		return fmt.Sprintf("%s.%s.%s()", sym.Package, sym.Receiver, sym.Name)
	}
	return fmt.Sprintf("%s.%s()", sym.Package, sym.Name)
}

func PrintPackageGraph(res *PackageGraphResult, jsonOutput bool) {
	if jsonOutput {
		printJSON(res)
		return
	}
	fmt.Println("Package:", res.Package)
	fmt.Println("\nImports:")
	for _, i := range res.Imports {
		fmt.Printf("- %s\n", i)
	}
	fmt.Println("\nImported By:")
	for _, i := range res.Imported {
		fmt.Printf("- %s\n", i)
	}
}

func PrintFileGraph(res *FileGraphResult, jsonOutput bool) {
	if jsonOutput {
		printJSON(res)
		return
	}
	fmt.Println("File:", res.File)
	fmt.Println("\nImports:")
	for _, i := range res.Imports {
		fmt.Printf("- %s\n", i)
	}
	fmt.Println("\nSymbols:")
	for _, s := range res.Symbols {
		fmt.Printf("- %s\n", s.Name)
	}
	fmt.Println("\nIncoming Dependencies:")
	for _, i := range res.Incoming {
		fmt.Printf("- %s\n", i)
	}
}



func PrintGraph(res *SymbolGraphResult, opts GraphOptions, jsonOutput bool) {
	if jsonOutput {
		printJSON(res)
		return
	}
	if len(res.Candidates) > 1 && res.Symbol == nil {
		printMultiple(res.Candidates)
		return
	}
	
	if res.Symbol == nil {
		fmt.Println("No symbol found.")
		return
	}

	sym := res.Symbol
	fmt.Println("Graph View")
	fmt.Println()
	fmt.Printf("Symbol      : %s %s.%s\n", sym.Kind, sym.Package, sym.Name)
	fmt.Printf("Package     : %s\n", sym.Package)
	fmt.Printf("File        : %s\n", sym.FilePath)
	fmt.Printf("Line        : %d\n", sym.StartLine)
	fmt.Printf("Language    : %s\n", "Go") // Assuming Go for now, or sym.Language if available. Wait, no Language field in Symbol. Just hardcode Go as per example.
	if sym.Documentation != "" {
		fmt.Println()
		fmt.Println("Documentation")
		fmt.Println()
		fmt.Println(strings.TrimSpace(sym.Documentation))
	}
	fmt.Println()

	if opts.TreeFormat {
		fmt.Printf("%s.%s\n", sym.Package, sym.Name)
		fmt.Println("│")
		
		// References usually show FilePath for "Referenced By"
		var refStrings []string
		for _, ref := range res.References {
			refStrings = append(refStrings, ref.File)
		}
		// deduplicate
		refMap := make(map[string]bool)
		var uniqueRefs []string
		for _, r := range refStrings {
			if !refMap[r] {
				refMap[r] = true
				uniqueRefs = append(uniqueRefs, r)
			}
		}

		sections := []struct {
			name  string
			items []string
		}{
			{"Methods", getSymbolNames(res.Methods, true)},
			{"Called By", getSymbolNames(res.Callers, true)},
			{"Calls", getSymbolNames(res.Callees, true)},
			{"Dependencies", res.Dependencies},
			{"Referenced By", uniqueRefs},
		}
		
		// Render tree sections
		for idx, sec := range sections {
			if len(sec.items) == 0 {
				continue
			}
			isLastSection := true
			for j := idx + 1; j < len(sections); j++ {
				if len(sections[j].items) > 0 {
					isLastSection = false
					break
				}
			}
			
			prefix := "├── "
			if isLastSection {
				prefix = "└── "
			}
			fmt.Printf("%s%s\n", prefix, sec.name)
			
			childPrefix := "│   "
			if isLastSection {
				childPrefix = "    "
			}
			
			for i, item := range sec.items {
				itemPrefix := "├── "
				if i == len(sec.items)-1 {
					itemPrefix = "└── "
				}
				fmt.Printf("%s%s%s\n", childPrefix, itemPrefix, item)
			}
			if !isLastSection {
				fmt.Println("│")
			}
		}
	} else {
		fmt.Println("====================================================")
		fmt.Println()
		
		printSection("Methods", getSymbolNames(res.Methods, true))
		printSection("Callers", getSymbolNames(res.Callers, true))
		printSection("Callees", getSymbolNames(res.Callees, true))
		printSection("Dependencies", res.Dependencies)
		
		// References usually show FilePath for "Referenced By". The prompt specifies "internal/cli/cli.go" for Referenced By
		var refStrings []string
		for _, ref := range res.References {
			refStrings = append(refStrings, ref.File)
		}
		// unique refs
		refMap := make(map[string]bool)
		var uniqueRefs []string
		for _, r := range refStrings {
			if !refMap[r] {
				refMap[r] = true
				uniqueRefs = append(uniqueRefs, r)
			}
		}
		printSection("Referenced By", uniqueRefs)
		
		fmt.Println("====================================================")
	}

	fmt.Println()
	fmt.Println("Statistics")
	fmt.Println()
	fmt.Printf("%-13s: %d\n", "Methods", res.Stats.Methods)
	fmt.Printf("%-13s: %d\n", "Callers", res.Stats.Callers)
	fmt.Printf("%-13s: %d\n", "Callees", res.Stats.Callees)
	fmt.Printf("%-13s: %d\n", "Dependencies", res.Stats.Dependencies)
	fmt.Printf("%-13s: %d\n", "References", res.Stats.References)
	fmt.Println()
	
	if res.Stats.LookupTimeMs < 1.0 {
		fmt.Printf("%-13s: %.0f µs\n", "Lookup Time", res.Stats.LookupTimeMs*1000)
	} else {
		fmt.Printf("%-13s: %.2f ms\n", "Lookup Time", res.Stats.LookupTimeMs)
	}
}

func getSymbolNames(syms []analyzer.Symbol, useReceiver bool) []string {
	var names []string
	for _, sym := range syms {
		if sym.Kind == "Method" && useReceiver {
			if strings.HasPrefix(sym.Receiver, "*") {
				names = append(names, fmt.Sprintf("%s.(%s).%s()", sym.Package, sym.Receiver, sym.Name))
			} else {
				names = append(names, fmt.Sprintf("%s.%s.%s()", sym.Package, sym.Receiver, sym.Name))
			}
		} else if sym.Kind == "Function" {
			names = append(names, fmt.Sprintf("%s.%s()", sym.Package, sym.Name))
		} else {
			names = append(names, fmt.Sprintf("%s.%s", sym.Package, sym.Name))
		}
	}
	
	// Deduplicate
	dedup := make(map[string]bool)
	var unique []string
	for _, n := range names {
		if !dedup[n] {
			dedup[n] = true
			unique = append(unique, n)
		}
	}
	return unique
}

func printSection(title string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Println(title)
	fmt.Println()
	for _, item := range items {
		fmt.Printf("• %s\n", item)
	}
	fmt.Println()
}

func printMultiple(candidates []analyzer.Symbol) {
	fmt.Println("Multiple definitions exist:")
	for i, c := range candidates {
		fmt.Printf("%d %s.%s (%s)\n", i+1, c.Package, c.Name, c.FilePath)
	}
	fmt.Println("\nUse --select <number> to specify.")
}

func printJSON(data interface{}) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(data)
}

func PrintCallersResult(res *CallersResult, opts CallersOptions, jsonOutput bool) {
	if jsonOutput {
		printJSON(res)
		return
	}
	if res.MultipleFound {
		printMultiple(res.Candidates)
		return
	}
	if len(res.Groups) == 0 {
		if opts.PackageFilter != "" {
			fmt.Printf("No callers found in package %q.\n", opts.PackageFilter)
		} else {
			fmt.Println("No callers found.")
		}
		return
	}

	sym := res.Symbol
	fmt.Println("Target Symbol")
	fmt.Printf("Symbol     : %s\n", sym.Name)
	fmt.Printf("Kind       : %s\n", sym.Kind)
	if sym.Receiver != "" {
		fmt.Printf("Receiver   : %s\n", sym.Receiver)
	}
	fmt.Printf("Package    : %s\n", sym.Package)
	fmt.Printf("File       : %s\n", sym.FilePath)
	fmt.Printf("Line       : %d\n", sym.StartLine)
	if sym.Documentation != "" {
		doc := strings.TrimSpace(sym.Documentation)
		if len(doc) > 80 {
			doc = doc[:77] + "..."
		}
		fmt.Printf("Doc        : %s\n", doc)
	}
	fmt.Println()

	for _, group := range res.Groups {
		fmt.Printf("Package %s:\n", group.Package)
		for _, caller := range group.Callers {
			fmt.Println("----------------------------------------")
			nameStr := caller.Name
			if caller.Kind == "Method" {
				nameStr = "Method " + caller.Name
			} else {
				nameStr = "Function " + caller.Name
			}
			if caller.IsRecursive {
				nameStr += " [Recursive Call]"
			}
			fmt.Printf("%-9s : %s\n", "Caller", nameStr)
			fmt.Printf("%-9s : %s\n", "File", caller.FilePath)
			fmt.Printf("%-9s : %d\n", "Line", caller.StartLine)
			
			if caller.Documentation != "" {
				doc := strings.TrimSpace(caller.Documentation)
				lines := strings.Split(doc, "\n")
				if len(lines) > 0 {
					fmt.Printf("%-9s : %s\n", "Doc", lines[0])
				}
			}

			if opts.TreeFormat {
				fmt.Println()
				fmt.Printf("%s()\n", caller.Name)
				fmt.Println("    │")
				
				targetName := sym.Name
				if sym.Kind == "Method" {
					if strings.HasPrefix(sym.Receiver, "*") {
						targetName = fmt.Sprintf("%s.(%s).%s", sym.Package, sym.Receiver, sym.Name)
					} else {
						targetName = fmt.Sprintf("%s.%s.%s", sym.Package, sym.Receiver, sym.Name)
					}
				}
				fmt.Printf("    └── %s()\n", targetName)
			}
		}
		fmt.Println()
	}

	fmt.Println("========================================")
	fmt.Printf("Total Callers : %d\n", res.Stats.TotalCallers)
	fmt.Printf("Files         : %d\n", res.Stats.TotalFiles)
	fmt.Printf("Packages      : %d\n", res.Stats.TotalPkgs)
	
	if res.Stats.LookupTimeMs < 1.0 {
		fmt.Printf("Lookup Time   : %.0f µs\n", res.Stats.LookupTimeMs*1000)
	} else {
		fmt.Printf("Lookup Time   : %.2f ms\n", res.Stats.LookupTimeMs)
	}
	
	fmt.Println("========================================")
}

func PrintCalleesResult(res *CalleesResult, opts CalleesOptions, jsonOutput bool) {
	if jsonOutput {
		printJSON(res)
		return
	}
	if res.MultipleFound {
		printMultiple(res.Candidates)
		return
	}
	if len(res.Groups) == 0 {
		if opts.PackageFilter != "" {
			fmt.Printf("No callees found in package %q.\n", opts.PackageFilter)
		} else {
			fmt.Println("No callees found.")
		}
		return
	}

	sym := res.Symbol
	fmt.Println("Target Symbol")
	fmt.Printf("Symbol     : %s\n", sym.Name)
	fmt.Printf("Kind       : %s\n", sym.Kind)
	if sym.Receiver != "" {
		fmt.Printf("Receiver   : %s\n", sym.Receiver)
	}
	fmt.Printf("Package    : %s\n", sym.Package)
	fmt.Printf("File       : %s\n", sym.FilePath)
	fmt.Printf("Line       : %d\n", sym.StartLine)
	if sym.Documentation != "" {
		doc := strings.TrimSpace(sym.Documentation)
		if len(doc) > 80 {
			doc = doc[:77] + "..."
		}
		fmt.Printf("Doc        : %s\n", doc)
	}
	fmt.Println()

	if opts.TreeFormat {
		fmt.Printf("%s()\n", sym.Name)
		fmt.Println("    │")
		
		var allNodes []CalleeNode
		for _, group := range res.Groups {
			allNodes = append(allNodes, group.Callees...)
		}
		
		for i, callee := range allNodes {
			prefix := "├── "
			if i == len(allNodes)-1 {
				prefix = "└── "
			}
			calleeName := callee.Name
			if callee.Kind == "Method" {
				if strings.HasPrefix(callee.Receiver, "*") {
					calleeName = fmt.Sprintf("%s.(%s).%s", callee.Package, callee.Receiver, callee.Name)
				} else {
					calleeName = fmt.Sprintf("%s.%s.%s", callee.Package, callee.Receiver, callee.Name)
				}
			} else if callee.Package != "" && callee.Package != sym.Package {
				calleeName = fmt.Sprintf("%s.%s", callee.Package, callee.Name)
			}
			fmt.Printf("    %s%s()\n", prefix, calleeName)
		}
		fmt.Println()
	} else {
		for _, group := range res.Groups {
			fmt.Printf("Package %s\n", group.Package)
			for _, callee := range group.Callees {
				fmt.Println("----------------------------------------")
				fmt.Printf("%-10s: %s %s\n", "Callee", callee.Kind, callee.Name)
				if callee.Receiver != "" {
					fmt.Printf("%-10s: %s\n", "Receiver", callee.Receiver)
				}
				fmt.Printf("%-10s: %s\n", "File", callee.FilePath)
				fmt.Printf("%-10s: %d\n", "Line", callee.StartLine)
				
				if callee.Documentation != "" {
					doc := strings.TrimSpace(callee.Documentation)
					lines := strings.Split(doc, "\n")
					if len(lines) > 0 {
						fmt.Printf("%-10s: %s\n", "Doc", lines[0])
					}
				}
			}
			fmt.Println()
		}
	}

	fmt.Println("========================================")
	fmt.Printf("Total Callees : %d\n", res.Stats.TotalCallees)
	fmt.Printf("Files         : %d\n", res.Stats.TotalFiles)
	fmt.Printf("Packages      : %d\n", res.Stats.TotalPkgs)
	
	if res.Stats.LookupTimeMs < 1.0 {
		fmt.Printf("Lookup Time   : %.0f µs\n", res.Stats.LookupTimeMs*1000)
	} else {
		fmt.Printf("Lookup Time   : %.2f ms\n", res.Stats.LookupTimeMs)
	}
	
	fmt.Println("========================================")
}

func PrintReferenceResult(res *ReferenceResult, opts ReferenceOptions, jsonOutput bool) {
	if jsonOutput {
		printJSON(res)
		return
	}
	if res.MultipleFound && !opts.Interactive {
		printMultiple(res.Candidates)
		return
	}

	sym := res.Symbol
	if sym == nil {
		fmt.Println("Target symbol not found.")
		return
	}

	if opts.TreeFormat {
		printReferenceTree(res)
		return
	}

	fmt.Print("Target Symbol\n\n")
	fmt.Printf("Symbol\n%s\n", sym.Name)
	fmt.Printf("Kind\n%s\n", sym.Kind)
	if sym.Receiver != "" {
		fmt.Printf("Receiver\n%s\n", sym.Receiver)
	}
	fmt.Printf("Package\n%s\n", sym.Package)
	fmt.Printf("File\n%s\n", sym.FilePath)
	fmt.Printf("Line\n%d\n", sym.StartLine)
	if sym.Documentation != "" {
		doc := strings.TrimSpace(sym.Documentation)
		doc = strings.ReplaceAll(doc, "\n", " ")
		doc = strings.ReplaceAll(doc, "\r", "")
		if len(doc) > 80 {
			doc = doc[:77] + "..."
		}
		fmt.Printf("Documentation\n%s\n", doc)
	}

	fmt.Print("\n──────────────────────────────────────\n\n")
	fmt.Print("References\n\n")

	if len(res.Groups) == 0 {
		fmt.Print("None\n\n")
	} else {
		for i, group := range res.Groups {
			fmt.Printf("Package %s\n\n", group.Package)
			
			for _, ref := range group.References {
				fmt.Print("----------------------------------------\n\n")
				fmt.Print("Reference\n\n")
				if ref.Relationship != "" {
					fmt.Printf("%s\n\n", ref.Relationship)
				} else {
					fmt.Printf("%s\n\n", ref.Name)
				}
				
				fmt.Print("File\n\n")
				fmt.Printf("%s\n\n", ref.File)
				
				fmt.Print("Line\n\n")
				fmt.Printf("%d\n\n", ref.Line)
				
				fmt.Print("Context\n\n")
				fmt.Printf("%s\n\n", ref.Context)
			}
			
			if i < len(res.Groups)-1 {
				fmt.Print("──────────────────────────────────────\n\n")
			}
		}
	}

	fmt.Print("──────────────────────────────────────\n\n")
	fmt.Print("Statistics\n\n")
	
	fmt.Printf("%-18s: %d\n", "Total References", res.Stats.TotalReferences)
	fmt.Printf("%-18s: %d\n", "Files", res.Stats.TotalFiles)
	fmt.Printf("%-18s: %d\n", "Packages", res.Stats.TotalPkgs)
	
	if res.LookupTimeMs < 1.0 {
		fmt.Printf("%-18s: %.0f µs\n", "Lookup Time", res.LookupTimeMs*1000)
	} else {
		fmt.Printf("%-18s: %.2f ms\n", "Lookup Time", res.LookupTimeMs)
	}
}

func printReferenceTree(res *ReferenceResult) {
	targetLabel := formatImpactSymbolLabel(res.Symbol)
	fmt.Println(targetLabel)
	
	if len(res.Groups) == 0 {
		return
	}
	
	var allRefs []Reference
	for _, g := range res.Groups {
		allRefs = append(allRefs, g.References...)
	}
	
	for i, ref := range allRefs {
		isLast := i == len(allRefs)-1
		branch := "├── "
		if isLast {
			branch = "└── "
		}
		
		rel := ref.Relationship
		if rel == "" {
			rel = ref.Name
		}
		if rel == "reference" || rel == "" {
			rel = ref.File // fallback
		} else {
			if !strings.Contains(rel, ".") {
				rel = ref.Package + "." + rel
			}
		}
		
		fmt.Printf("%s%s\n", branch, rel)
	}
}

func PrintImplementsResult(res *ImplementsResult, opts ImplementsOptions, jsonOutput bool) {
	if jsonOutput {
		printJSON(res)
		return
	}
	if res.MultipleFound && !opts.Interactive {
		printMultiple(res.Candidates)
		return
	}
	if res.Symbol == nil {
		fmt.Println("No symbol found.")
		return
	}

	if opts.TreeFormat {
		printImplementsTree(res)
		return
	}

	sym := res.Symbol

	fmt.Print("Target Interface\n\n")
	fmt.Printf("%s\n", sym.Name)
	fmt.Printf("Package: %s\n\n", sym.Package)

	fmt.Print("Required Methods\n\n")
	if len(res.RequiredMethods) == 0 {
		fmt.Println("None")
	} else {
		for _, m := range res.RequiredMethods {
			fmt.Printf("• %s\n", m)
		}
	}
	fmt.Println()

	fmt.Print("Implementations\n\n")
	if len(res.Implementations) == 0 {
		fmt.Println("No implementations found.")
		return
	}

	for _, impl := range res.Implementations {
		fmt.Printf("✓ %s.%s\n", impl.Package, impl.Struct)
		if len(impl.ImplementedMethods) > 0 {
			fmt.Println("    implements:")
			for _, m := range impl.ImplementedMethods {
				fmt.Printf("      ✓ %s()\n", m)
			}
		}
		fmt.Println()
	}
}

func printImplementsTree(res *ImplementsResult) {
	fmt.Println(res.Symbol.Name)
	
	for i, impl := range res.Implementations {
		prefix := "├──"
		if i == len(res.Implementations)-1 {
			prefix = "└──"
		}
		fmt.Printf("%s %s\n", prefix, impl.Struct)
		
		for j, m := range impl.ImplementedMethods {
			childPrefix := "│   ├──"
			if i == len(res.Implementations)-1 {
				childPrefix = "    ├──"
			}
			if j == len(impl.ImplementedMethods)-1 {
				if i == len(res.Implementations)-1 {
					childPrefix = "    └──"
				} else {
					childPrefix = "│   └──"
				}
			}
			fmt.Printf("%s %s()\n", childPrefix, m)
		}
	}
}

func PrintInterfaceResult(res *InterfaceResult, opts InterfaceOptions, jsonOutput bool) {
	if jsonOutput {
		printJSON(res)
		return
	}

	if res.MultipleFound && !opts.Interactive {
		printMultiple(res.Candidates)
		return
	}
	if res.Symbol == nil {
		fmt.Println("No symbol found.")
		return
	}

	if opts.TreeFormat {
		fmt.Printf("%s\n", res.Symbol.Name)
		if len(res.Interfaces) == 0 {
			// Do nothing or maybe print a blank branch if needed, but wait
			return
		}
		for i, iface := range res.Interfaces {
			prefix := "├──"
			if i == len(res.Interfaces)-1 {
				prefix = "└──"
			}
			fmt.Printf("%s %s\n", prefix, iface.Name)

			for j, m := range iface.ImplementedMethods {
				methodPrefix := "│   ├──"
				if i == len(res.Interfaces)-1 {
					methodPrefix = "    ├──"
				}
				if j == len(iface.ImplementedMethods)-1 {
					if i == len(res.Interfaces)-1 {
						methodPrefix = "    └──"
					} else {
						methodPrefix = "│   └──"
					}
				}
				fmt.Printf("%s %s()\n", methodPrefix, m)
			}
			if i != len(res.Interfaces)-1 {
				fmt.Println("│")
			}
		}
		return
	}

	sym := res.Symbol
	fmt.Print("Target Type\n\n")
	fmt.Printf("%s\n", sym.Name)
	fmt.Printf("Kind: %s\n", sym.Kind)
	fmt.Printf("Package: %s\n", sym.Package)
	fmt.Printf("File: %s\n", sym.FilePath)
	fmt.Printf("Line: %d\n", sym.StartLine)
	if sym.Documentation != "" {
		fmt.Printf("Documentation: %s\n", sym.Documentation)
	}

	fmt.Print("\n──────────────────────────────────────\n\n")
	fmt.Print("Implemented Interfaces\n\n")

	for i, iface := range res.Interfaces {
		fmt.Printf("Package %s\n\n", iface.Package)
		fmt.Print("----------------------------------------\n\n")
		fmt.Print("Interface\n\n")
		fmt.Printf("%s\n\n", iface.Name)

		fmt.Print("Required Methods\n\n")
		if len(iface.RequiredMethods) == 0 {
			fmt.Print("None\n\n")
		} else {
			for _, m := range iface.RequiredMethods {
				fmt.Printf("• %s\n", m)
			}
			fmt.Println()
		}

		fmt.Print("Implemented Methods\n\n")
		if len(iface.ImplementedMethods) == 0 {
			fmt.Print("None\n\n")
		} else {
			for _, m := range iface.ImplementedMethods {
				fmt.Printf("✓ %s()\n", m)
			}
			fmt.Println()
		}

		fmt.Print("Missing Methods\n\n")
		if len(iface.MissingMethods) == 0 {
			fmt.Print("None\n\n")
		} else {
			for _, m := range iface.MissingMethods {
				fmt.Printf("✗ %s()\n", m)
			}
			fmt.Println()
		}

		fmt.Print("Pointer Receiver\n\n")
		if iface.PointerReceiver {
			fmt.Print("Yes\n\n")
		} else {
			fmt.Print("No\n\n")
		}

		if i < len(res.Interfaces)-1 {
			fmt.Print("──────────────────────────────────────\n\n")
		}
	}

	if len(res.Interfaces) == 0 {
		fmt.Print("None\n\n")
	}

	fmt.Print("──────────────────────────────────────\n\n")
	fmt.Print("Statistics\n\n")

	fmt.Printf("%-20s %d\n", "Total Interfaces", res.Stats.TotalInterfaces)
	fmt.Printf("%-20s %d\n", "Packages", res.Stats.TotalPackages)
	
	if res.LookupTimeMs < 1.0 {
		fmt.Printf("%-20s %.0f µs\n", "Lookup Time", res.LookupTimeMs*1000)
	} else {
		fmt.Printf("%-20s %.2f ms\n", "Lookup Time", res.LookupTimeMs)
	}
}
