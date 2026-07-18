package graph

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
)

// findCallersOf returns functions/methods that invoke selected, using the
// existing lexical call-graph scan (no AST, no repository rescan).
// fileCache is shared across lookups so Impacts stays O(N) over unique symbols.
func (e *Engine) findCallersOf(selected *analyzer.Symbol, fileCache map[string][]string) []CallerNode {
	if selected == nil {
		return nil
	}
	if e.fileCache == nil {
		e.fileCache = make(map[string][]string)
	}

	getFileLines := func(filePath string) []string {
		if lines, ok := e.fileCache[filePath]; ok {
			return lines
		}
		var lines []string
		if content, err := os.ReadFile(filePath); err == nil {
			lines = strings.Split(string(content), "\n")
		}
		e.fileCache[filePath] = lines
		return lines
	}

	targetPkg := selected.Package
	var found []CallerNode
	seen := make(map[string]bool)

	for file, deps := range e.repo.Deps.FileDeps {
		importsTarget := false
		for _, dep := range deps {
			if strings.HasSuffix(dep, targetPkg) {
				importsTarget = true
				break
			}
		}

		var filePkg string
		for _, sym := range e.repo.Symbols {
			if filepath.ToSlash(sym.FilePath) == filepath.ToSlash(file) {
				filePkg = sym.Package
				break
			}
		}
		if filePkg == targetPkg {
			importsTarget = true
		}

		if !importsTarget {
			continue
		}

		for _, sym := range e.repo.Symbols {
			if sym.Kind != "Function" && sym.Kind != "Method" {
				continue
			}
			if filepath.ToSlash(sym.FilePath) != filepath.ToSlash(file) {
				continue
			}

			key := symbolKey(sym.Package, sym.Name, sym.Receiver)
			if seen[key] {
				continue
			}

			isSelf := sym.Name == selected.Name &&
				sym.Package == selected.Package &&
				sym.Receiver == selected.Receiver

			lines := getFileLines(sym.FilePath)
			if len(lines) == 0 {
				continue
			}

			start := sym.StartLine - 1
			if start < 0 {
				start = 0
			}
			end := sym.EndLine
			if end > len(lines) {
				end = len(lines)
			}
			if start > end {
				start = end
			}

			if isSelf {
				start = sym.StartLine
				if start >= end {
					continue
				}
			}

			body := strings.Join(lines[start:end], "\n")
			braceIdx := strings.Index(body, "{")
			if braceIdx != -1 {
				body = body[braceIdx:]
			}
			body = stripLineComments(body)

			if !containsToken(body, selected.Name) {
				continue
			}

			seen[key] = true
			found = append(found, CallerNode{
				Symbol:      sym,
				IsRecursive: isSelf,
			})
		}
	}

	return found
}

func stripLineComments(body string) string {
	var cleanBody strings.Builder
	inLineComment := false
	for i := 0; i < len(body); i++ {
		if inLineComment {
			if body[i] == '\n' {
				inLineComment = false
				cleanBody.WriteByte(body[i])
			}
			continue
		}
		if i+1 < len(body) && body[i] == '/' && body[i+1] == '/' {
			inLineComment = true
			i++
			continue
		}
		cleanBody.WriteByte(body[i])
	}
	return cleanBody.String()
}

func containsToken(body, token string) bool {
	idx := strings.Index(body, token)
	for idx != -1 {
		preOK, postOK := true, true
		if idx > 0 {
			preChar := body[idx-1]
			if isIdentChar(preChar) {
				preOK = false
			}
		}
		endIdx := idx + len(token)
		if endIdx < len(body) {
			postChar := body[endIdx]
			if isIdentChar(postChar) {
				postOK = false
			}
		}
		if preOK && postOK {
			return true
		}
		nextIdx := strings.Index(body[idx+1:], token)
		if nextIdx == -1 {
			return false
		}
		idx += 1 + nextIdx
	}
	return false
}

func isIdentChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

func symbolKey(pkg, name, receiver string) string {
	if receiver != "" {
		return pkg + "." + receiver + "." + name
	}
	return pkg + "." + name
}

func (e *Engine) lookupExactSymbol(pkg, name, receiver string) *analyzer.Symbol {
	indices, ok := e.repo.ByName[name]
	if !ok {
		return nil
	}
	for _, idx := range indices {
		s := &e.repo.Symbols[idx]
		if s.Package == pkg && s.Receiver == receiver {
			return s
		}
	}
	// Fall back to package+name when receiver was unknown.
	for _, idx := range indices {
		s := &e.repo.Symbols[idx]
		if s.Package == pkg {
			return s
		}
	}
	return nil
}
