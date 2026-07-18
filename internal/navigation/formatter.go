package navigation

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
)

func PrintGoto(res *GotoResult, open bool, jsonOutput bool) error {
	if jsonOutput {
		return printJSON(res)
	}

	if res.MultipleFound {
		printMultiple(res.Candidates)
		return nil
	}

	if open {
		fmt.Printf("%s:%d\n", res.Symbol.FilePath, res.Symbol.StartLine)
		return nil
	}

	fmt.Println("Definition Found")
	fmt.Println("File")
	fmt.Println(res.Symbol.FilePath)
	fmt.Println("Package")
	fmt.Println(res.Symbol.Package)
	fmt.Println("Line")
	fmt.Printf("%d\n", res.Symbol.StartLine)
	if res.Symbol.Signature != "" {
		fmt.Println("Signature")
		fmt.Println(res.Symbol.Signature)
	}
	if res.Symbol.Documentation != "" {
		fmt.Println("Documentation")
		fmt.Println(res.Symbol.Documentation)
	}
	return nil
}

func PrintPeek(res *PeekResult, jsonOutput bool) error {
	if jsonOutput {
		return printJSON(res)
	}

	if res.MultipleFound {
		printMultiple(res.Candidates)
		return nil
	}

	fileInfo, _ := os.Stdout.Stat()
	useColor := (fileInfo.Mode() & os.ModeCharDevice) != 0

	cyan := "\033[36m"
	reset := "\033[0m"
	if !useColor {
		cyan = ""
		reset = ""
	}

	fmt.Println("Peek Definition")
	fmt.Println()
	fmt.Printf("%-12s: %s%s %s.%s%s\n", "Symbol", cyan, res.Symbol.Kind, res.Symbol.Package, res.Symbol.Name, reset)
	fmt.Printf("%-12s: %s\n", "Package", res.Symbol.Package)
	fmt.Printf("%-12s: %s\n", "File", res.Symbol.FilePath)
	fmt.Printf("%-12s: %d\n", "Line", res.Symbol.StartLine)
	fmt.Printf("%-12s: %s\n\n", "Language", res.Symbol.Language)

	if res.Symbol.Documentation != "" {
		fmt.Println("Documentation")
		fmt.Println()
		fmt.Println(strings.TrimSpace(res.Symbol.Documentation))
		fmt.Println()
	}

	fmt.Println("====================================================")
	
	lines := strings.Split(res.Snippet, "\n")
	startLine := res.Symbol.StartLine
	for i, line := range lines {
		formattedLine := line
		if useColor {
			formattedLine = applySyntaxHighlighting(line)
		}
		fmt.Printf("%d | %s\n", startLine+i, formattedLine)
	}

	if res.Truncated {
		fmt.Println("\n──────────────────────────────")
		fmt.Printf("Preview truncated (%d additional lines)\n", res.TruncatedLines)
		fmt.Println()
		fmt.Println("Use:")
		fmt.Println()
		fmt.Printf("kyver goto %s\n\n", res.Symbol.Name)
		fmt.Println("to view the complete definition.")
		fmt.Println("──────────────────────────────")
	}

	fmt.Println("\n====================================================")
	
	if len(res.Methods) > 0 {
		fmt.Println("Methods")
		fmt.Println()
		for _, m := range res.Methods {
			fmt.Printf("• %s()\n", m.Name)
		}
		fmt.Println()
	}

	if len(res.Implements) > 0 {
		fmt.Println("Implements")
		fmt.Println()
		for _, impl := range res.Implements {
			fmt.Printf("• %s\n", impl.Name)
		}
		fmt.Println()
	}

	if len(res.ReferencedBy) > 0 {
		fmt.Println("Referenced By")
		fmt.Println()
		for _, ref := range res.ReferencedBy {
			fmt.Printf("• %s\n", ref)
		}
		fmt.Println()
	}

	if len(res.Dependencies) > 0 {
		fmt.Println("Dependencies")
		fmt.Println()
		for _, dep := range res.Dependencies {
			fmt.Printf("• %s\n", dep)
		}
		fmt.Println()
	}

	fmt.Println("Navigation Hints")
	fmt.Println()
	fmt.Println("Related Commands")
	fmt.Println()
	fmt.Printf("kyver goto %s\n", res.Symbol.Name)
	fmt.Printf("kyver context %s\n", res.Symbol.Name)
	fmt.Printf("kyver callers %s\n", res.Symbol.Name)
	fmt.Printf("kyver hierarchy %s\n", res.Symbol.Name)
	
	return nil
}

func applySyntaxHighlighting(line string) string {
	blue := "\033[34m"
	cyan := "\033[36m"
	green := "\033[32m"
	gray := "\033[90m"
	yellow := "\033[33m"
	reset := "\033[0m"

	// Very basic regex-free highlighting for the demo
	if strings.HasPrefix(strings.TrimSpace(line), "//") {
		return gray + line + reset
	}

	tokens := []string{"func", "type", "struct", "interface", "var", "const", "return", "if", "else", "for", "switch", "case"}
	highlighted := line
	for _, t := range tokens {
		// naive replacement, ensuring word boundaries would require regexp, but keeping it fast and simple
		// for now we just use a simple strings.Replace for keywords if they are surrounded by space
		highlighted = strings.ReplaceAll(highlighted, " "+t+" ", " "+blue+t+reset+" ")
		if strings.HasPrefix(highlighted, t+" ") {
			highlighted = blue + t + reset + highlighted[len(t):]
		}
	}

	// types
	types := []string{"string", "int", "bool", "float64", "error"}
	for _, t := range types {
		highlighted = strings.ReplaceAll(highlighted, " "+t+" ", " "+cyan+t+reset+" ")
		highlighted = strings.ReplaceAll(highlighted, " "+t+"\n", " "+cyan+t+reset+"\n")
		if strings.HasSuffix(highlighted, " "+t) {
			highlighted = highlighted[:len(highlighted)-len(t)-1] + " " + cyan + t + reset
		}
	}
	
	// simple function call/definition highlighting
	if strings.Contains(highlighted, "func ") {
		parts := strings.Split(highlighted, "(")
		if len(parts) > 1 {
			funcNameParts := strings.Split(parts[0], " ")
			funcName := funcNameParts[len(funcNameParts)-1]
			if funcName != "" {
				highlighted = strings.Replace(highlighted, funcName+"(", yellow+funcName+reset+"(", 1)
			}
		}
	}

	// strings (naive)
	if strings.Contains(highlighted, "\"") {
		parts := strings.Split(highlighted, "\"")
		for i := 1; i < len(parts); i += 2 {
			parts[i] = green + parts[i] + reset
		}
		highlighted = strings.Join(parts, "\"")
	}

	return highlighted
}

func PrintOutline(res *OutlineResult, jsonOutput bool) error {
	if jsonOutput {
		return printJSON(res)
	}

	fileInfo, _ := os.Stdout.Stat()
	useColor := (fileInfo.Mode() & os.ModeCharDevice) != 0

	cyan := "\033[36m"
	green := "\033[32m"
	gray := "\033[90m"
	reset := "\033[0m"
	if !useColor {
		cyan = ""
		green = ""
		gray = ""
		reset = ""
	}

	fmt.Println("File Outline")
	fmt.Println()
	fmt.Printf("%-12s: %s\n", "File", res.File)
	fmt.Printf("%-12s: %s\n", "Package", res.Package)
	fmt.Printf("%-12s: %s\n", "Language", res.Language)
	fmt.Println()

	if len(res.Imports) > 0 {
		fmt.Println("====================================================")
		fmt.Println("Imports")
		fmt.Println("====================================================")
		for _, imp := range res.Imports {
			fmt.Printf("• %s\n", imp)
		}
		fmt.Println()
	}

	for _, group := range res.Groups {
		fmt.Println("====================================================")
		fmt.Printf("%s\n", group.Kind)
		fmt.Println("====================================================")
		for _, sym := range group.Symbols {
			visibility := "[-]"
			if sym.Exported {
				visibility = "[+]"
				visibility = green + visibility + reset
			} else {
				visibility = gray + visibility + reset
			}
			
			lineStr := fmt.Sprintf("%4d", sym.StartLine)
			
			docPreview := ""
			if sym.Documentation != "" {
				lines := strings.Split(strings.TrimSpace(sym.Documentation), "\n")
				if len(lines) > 0 {
					docPreview = " // " + lines[0]
					if len(docPreview) > 60 {
						docPreview = docPreview[:57] + "..."
					}
					docPreview = gray + docPreview + reset
				}
			}
			
			fmt.Printf("%s %s %s%s%s%s\n", cyan+lineStr+reset, visibility, cyan, sym.Name, reset, docPreview)
		}
		fmt.Println()
	}

	fmt.Println("====================================================")
	fmt.Println("Statistics")
	fmt.Println("====================================================")
	fmt.Printf("%-15s: %d\n", "Total Symbols", res.Stats.TotalSymbols)
	fmt.Printf("%-15s: %d\n", "Total Imports", res.Stats.TotalImports)
	for kind, count := range res.Stats.Counts {
		fmt.Printf("%-15s: %d\n", kind, count)
	}

	return nil
}

func PrintHierarchy(res *HierarchyResult, jsonOutput bool) error {
	if jsonOutput {
		return printJSON(res)
	}

	if res.MultipleFound {
		printMultiple(res.Candidates)
		return nil
	}

	fmt.Println(res.Symbol.Name)
	for i, m := range res.Methods {
		prefix := "├── "
		if i == len(res.Methods)-1 && len(res.Related) == 0 && len(res.Dependencies) == 0 {
			prefix = "└── "
		}
		fmt.Printf("%s%s()\n", prefix, m.Name)
	}
	
	if len(res.Related) > 0 {
		fmt.Println("\nRelated")
		for _, r := range res.Related {
			fmt.Printf("- %s (%s)\n", r.Name, r.Kind)
		}
	}
	
	if len(res.Dependencies) > 0 {
		fmt.Println("\nDependencies")
		for _, d := range res.Dependencies {
			fmt.Printf("- %s\n", d)
		}
	}

	return nil
}

func printMultiple(candidates []analyzer.Symbol) {
	fmt.Println("Multiple definitions exist:")
	fmt.Println()
	for i, c := range candidates {
		// e.g. "1 Struct  repository.Engine"
		fmt.Printf("%d %s  %s.%s\n", i+1, c.Kind, c.Package, c.Name)
		fmt.Printf("Location : %s:%d\n", c.FilePath, c.StartLine)
		if c.Documentation != "" {
			fmt.Println("Documentation:")
			lines := strings.Split(strings.TrimSpace(c.Documentation), "\n")
			fmt.Println(lines[0])
		}
		fmt.Println()
	}
	fmt.Println("Use --select <number> to specify.")
}

func printJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
