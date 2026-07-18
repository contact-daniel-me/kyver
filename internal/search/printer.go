package search

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"
)

func PrintPlain(q Query, results []Result, related []Result, references []string, d time.Duration) {
	fmt.Printf("Found %d results in %v.\n\n", len(results), d)

	groups := map[string][]Result{
		"Struct":    {},
		"Interface": {},
		"Function":  {},
		"Method":    {},
		"Constant":  {},
		"Variable":  {},
		"Package":   {},
	}

	for _, res := range results {
		kind := res.Kind
		if kind == "Type" {
			kind = "Struct" // Fallback mapping
		}
		if _, exists := groups[kind]; !exists {
			kind = "Struct" // Unknown goes to Struct for simplicity
		}
		groups[kind] = append(groups[kind], res)
	}

	order := []string{"Struct", "Interface", "Function", "Method", "Constant", "Variable", "Package"}

	fi, _ := os.Stdout.Stat()
	useColor := (fi.Mode() & os.ModeCharDevice) != 0

	colors := map[string]string{
		"Struct":    "\033[34m", // Blue
		"Interface": "\033[36m", // Cyan
		"Function":  "\033[32m", // Green
		"Method":    "\033[33m", // Yellow
		"Constant":  "\033[35m", // Magenta
		"Variable":  "\033[37m", // White
		"Package":   "\033[0m",
	}
	reset := "\033[0m"
	boldCyan := "\033[1;36m"

	for _, kind := range order {
		for _, res := range groups[kind] {
			fmt.Println("------------------------------------------------")
			
			kindDisplay := res.Kind
			if useColor && colors[kind] != "" {
				kindDisplay = colors[kind] + res.Kind + reset
			}
			
			// Highlight matched portion
			name := res.Name
			if q.Text != "" && useColor {
				re := regexp.MustCompile("(?i)(" + regexp.QuoteMeta(q.Text) + ")")
				name = re.ReplaceAllString(name, boldCyan+"$1"+reset)
			}
			
			fmt.Printf("%s: %s\n", kindDisplay, name)
			
			fmt.Println()
			if res.Receiver != "" {
				fmt.Printf("Receiver: %s\n\n", res.Receiver)
			}
			
			fmt.Printf("Package : %s\n\n", res.Package)
			
			if res.Signature != "" {
				fmt.Printf("Signature:\n%s\n\n", res.Signature)
			}
			
			if res.File != "" {
				fmt.Printf("Location: %s:%d\n\n", res.File, res.Line)
			}

			if res.Documentation != "" {
				fmt.Println("Documentation:")
				fmt.Println(res.Documentation)
				fmt.Println()
			}
			
			fmt.Printf("Score : %.0f\n", res.Score)
			fmt.Printf("Reason: %s\n", res.Reason)
		}
	}
	if len(results) > 0 {
		fmt.Println("------------------------------------------------")
	}

	if len(related) > 0 {
		fmt.Println("------------------------------------------------")
		relGroups := make(map[string][]Result)
		for _, r := range related {
			kind := r.Kind
			if kind == "Type" {
				kind = "Struct"
			}
			relGroups[kind] = append(relGroups[kind], r)
		}
		
		for _, kind := range order {
			if len(relGroups[kind]) > 0 {
				plural := kind + "s"
				if kind == "Package" {
					plural = "Packages"
				}
				fmt.Printf("Related %s\n\n", plural)
				for _, r := range relGroups[kind] {
					fmt.Println(r.Name)
				}
				fmt.Println()
			}
		}
	}

	if len(references) > 0 {
		fmt.Println("References")
		for _, r := range references {
			fmt.Println(r)
		}
		fmt.Println("------------------------------------------------")
	}

	// Result Summary
	fmt.Println("Results")
	fmt.Println()
	
	totalScore := 0.0
	highestScore := 0.0
	lowestScore := 100.0
	
	for _, kind := range order {
		count := len(groups[kind])
		if count > 0 {
			plural := kind + "s"
			if kind == "Package" {
				plural = "Packages"
			}
			fmt.Printf("%-13s: %d\n", plural, count)
			
			for _, r := range groups[kind] {
				totalScore += r.Score
				if r.Score > highestScore {
					highestScore = r.Score
				}
				if r.Score < lowestScore {
					lowestScore = r.Score
				}
			}
		}
	}
	
	if len(results) == 0 {
		lowestScore = 0
	}
	avgScore := 0.0
	if len(results) > 0 {
		avgScore = totalScore / float64(len(results))
	}
	
	fmt.Println()
	fmt.Printf("%-13s: %d\n", "Total", len(results))
	fmt.Println()
	fmt.Printf("%-13s: %.0f\n", "Average Score", avgScore)
	fmt.Printf("%-13s: %.0f\n", "Highest Score", highestScore)
	fmt.Printf("%-13s: %.0f\n", "Lowest Score", lowestScore)
	
	ms := float64(d.Microseconds()) / 1000.0
	fmt.Printf("%-13s: %.2f ms\n", "Execution", ms)
}

type JSONOutput struct {
	Query          string                 `json:"query"`
	ExecutionMs    float64                `json:"execution_ms"`
	Count          int                    `json:"count"`
	Results        []Result               `json:"results"`
	RelatedSymbols []Result               `json:"relatedSymbols,omitempty"`
	References     []string               `json:"references,omitempty"`
	Summary        map[string]interface{} `json:"summary"`
}

func PrintJSON(q Query, results []Result, related []Result, references []string, d time.Duration) error {
	ms := float64(d.Microseconds()) / 1000.0
	summary := make(map[string]interface{})
	
	totalScore := 0.0
	for _, r := range results {
		totalScore += r.Score
	}
	if len(results) > 0 {
		summary["average_score"] = totalScore / float64(len(results))
	} else {
		summary["average_score"] = 0
	}
	
	output := JSONOutput{
		Query:          q.Text,
		ExecutionMs:    ms,
		Count:          len(results),
		Results:        results,
		RelatedSymbols: related,
		References:     references,
		Summary:        summary,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
