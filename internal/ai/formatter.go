package ai

import (
	"fmt"
)

// PrintAskResult formats the AskResult struct for the CLI standard output.
func PrintAskResult(res *AskResult) {
	fmt.Printf("Question\n\n%s\n", res.Question)
	fmt.Printf("\nRelevant Symbol\n\n%s\n", res.TargetSymbol)
	fmt.Printf("\nSummary\n\n%s\n", res.Summary)
	fmt.Printf("\nDetailed Explanation\n\n%s\n", res.Explanation)

	if len(res.RelatedFiles) > 0 {
		fmt.Printf("\nRelated Files\n\n")
		for _, f := range res.RelatedFiles {
			fmt.Printf("• %s\n", f)
		}
	} else {
		fmt.Printf("\nRelated Files\n\nNone\n")
	}

	fmt.Printf("\nConfidence: %s\n", res.Confidence)
	if res.LookupTimeMs < 1.0 {
		fmt.Printf("Lookup Time: %.0f µs\n", res.LookupTimeMs*1000)
	} else {
		fmt.Printf("Lookup Time: %.2f ms\n", res.LookupTimeMs)
	}
}
