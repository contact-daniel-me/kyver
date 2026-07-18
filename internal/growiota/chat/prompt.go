package chat

import (
	"fmt"
	"strings"

	"github.com/contact-daniel-me/kyver/internal/retrieval"
)

func BuildSystemPrompt(bundle *retrieval.RetrievalBundle) string {
	var sb strings.Builder

	sb.WriteString("You are Kyver, an AI-powered software engineering platform.\n")
	sb.WriteString("You are answering a developer's question about their codebase.\n")
	sb.WriteString("IMPORTANT INSTRUCTIONS:\n")
	sb.WriteString("1. You must ONLY use the provided context to answer the question.\n")
	sb.WriteString("2. If the answer is not in the context, state that you do not know based on the current context.\n")
	sb.WriteString("3. ALWAYS cite your sources (files, symbols) at the end of your response.\n\n")

	if bundle != nil && bundle.Confidence > 0 {
		sb.WriteString("=== RETRIEVED CONTEXT ===\n")
		
		if len(bundle.TopSymbols) > 0 {
			sb.WriteString("\nSYMBOLS:\n")
			for _, sym := range bundle.TopSymbols {
				sb.WriteString(fmt.Sprintf("- %s (%s) in %s\n", sym.Name, sym.Kind, sym.FilePath))
				if sym.Signature != "" {
					sb.WriteString(fmt.Sprintf("  Signature: %s\n", sym.Signature))
				}
				if sym.Documentation != "" {
					sb.WriteString(fmt.Sprintf("  Doc: %s\n", sym.Documentation))
				}
			}
		}

		if len(bundle.TopFiles) > 0 {
			sb.WriteString("\nFILES:\n")
			for _, f := range bundle.TopFiles {
				sb.WriteString(fmt.Sprintf("- %s\n", f))
			}
		}

		if len(bundle.TopPackages) > 0 {
			sb.WriteString("\nPACKAGES:\n")
			for _, p := range bundle.TopPackages {
				sb.WriteString(fmt.Sprintf("- %s\n", p))
			}
		}

		if len(bundle.TopContexts) > 0 {
			sb.WriteString("\nDEEP STRUCTURAL CONTEXT:\n")
			for _, ctx := range bundle.TopContexts {
				sb.WriteString(fmt.Sprintf("Target: %s (%s)\n", ctx.Target, ctx.Type))
				if len(ctx.Dependencies) > 0 {
					sb.WriteString(fmt.Sprintf("  Dependencies: %s\n", strings.Join(ctx.Dependencies, ", ")))
				}
				if len(ctx.Dependents) > 0 {
					sb.WriteString(fmt.Sprintf("  Dependents: %s\n", strings.Join(ctx.Dependents, ", ")))
				}
			}
		}
		sb.WriteString("==========================\n\n")
	}

	return sb.String()
}
