package analyzer

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/contact-daniel-me/kyver/internal/indexer"
)

type GoAnalyzer struct{}

func (a *GoAnalyzer) Language() string {
	return "Go"
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (a *GoAnalyzer) Analyze(fullPath string, file indexer.FileMetadata) (*AnalysisResult, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fullPath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", file.Path, err)
	}

	result := &AnalysisResult{
		PackageName: node.Name.Name,
		Symbols:     []Symbol{},
		Imports:     []string{},
		Metrics:     FileMetrics{},
		FilePath:    file.Path,
	}

	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, "\"")
		result.Imports = append(result.Imports, path)
	}
	result.Metrics.Imports = len(result.Imports)

	var lastLine int
	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		pos := fset.Position(n.End()).Line
		if pos > lastLine {
			lastLine = pos
		}
		return true
	})
	result.Metrics.LOC = lastLine

	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			sym := a.parseFuncDecl(fset, d, file.Path, result.PackageName)
			result.Symbols = append(result.Symbols, sym)

			if sym.Kind == "Method" {
				result.Metrics.Methods++
			} else {
				result.Metrics.Functions++
			}

			if sym.Exported {
				result.Metrics.Exported++
			}

			lines := sym.EndLine - sym.StartLine
			result.Metrics.TotalFuncLines += lines
			if lines > result.Metrics.MaxFuncLines {
				result.Metrics.MaxFuncLines = lines
			}

		case *ast.GenDecl:
			syms := a.parseGenDecl(fset, d, file.Path, result.PackageName)
			for _, sym := range syms {
				result.Symbols = append(result.Symbols, sym)
				if sym.Kind == "Struct" {
					result.Metrics.Structs++
				} else if sym.Kind == "Interface" {
					result.Metrics.Interfaces++
				} else if sym.Kind == "Constant" {
					result.Metrics.Constants++
				} else if sym.Kind == "Variable" {
					result.Metrics.Variables++
				} else if sym.Kind == "Type" {
					result.Metrics.Aliases++
				}
				if sym.Exported {
					result.Metrics.Exported++
				}
			}
		}
	}

	for _, cg := range node.Comments {
		result.Metrics.CommentsCount += len(cg.List)
	}

	return result, nil
}

func (a *GoAnalyzer) parseFuncDecl(fset *token.FileSet, d *ast.FuncDecl, filePath, pkgName string) Symbol {
	name := d.Name.Name
	exported := ast.IsExported(name)

	startLine := fset.Position(d.Pos()).Line
	endLine := fset.Position(d.End()).Line

	kind := "Function"
	receiver := ""
	if d.Recv != nil && len(d.Recv.List) > 0 {
		kind = "Method"
		switch t := d.Recv.List[0].Type.(type) {
		case *ast.Ident:
			receiver = t.Name
		case *ast.StarExpr:
			if ident, ok := t.X.(*ast.Ident); ok {
				receiver = "*" + ident.Name
			}
		}
	}

	doc := ""
	if d.Doc != nil {
		doc = d.Doc.Text()
	}

	return Symbol{
		ID:            generateID(),
		Name:          name,
		Kind:          kind,
		Language:      "Go",
		Package:       pkgName,
		Receiver:      receiver,
		FilePath:      filePath,
		StartLine:     startLine,
		EndLine:       endLine,
		Exported:      exported,
		Documentation: strings.TrimSpace(doc),
		Signature:     "func " + name, // basic signature
	}
}

func (a *GoAnalyzer) parseGenDecl(fset *token.FileSet, d *ast.GenDecl, filePath, pkgName string) []Symbol {
	var symbols []Symbol

	kindMap := map[token.Token]string{
		token.TYPE:  "Type",
		token.CONST: "Constant",
		token.VAR:   "Variable",
	}

	baseKind := kindMap[d.Tok]

	doc := ""
	if d.Doc != nil {
		doc = d.Doc.Text()
	}

	for _, spec := range d.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			name := s.Name.Name
			kind := baseKind

			switch s.Type.(type) {
			case *ast.StructType:
				kind = "Struct"
			case *ast.InterfaceType:
				kind = "Interface"
			}

			symDoc := doc
			if s.Doc != nil {
				symDoc = s.Doc.Text()
			}

			symbols = append(symbols, Symbol{
				ID:            generateID(),
				Name:          name,
				Kind:          kind,
				Language:      "Go",
				Package:       pkgName,
				FilePath:      filePath,
				StartLine:     fset.Position(s.Pos()).Line,
				EndLine:       fset.Position(s.End()).Line,
				Exported:      ast.IsExported(name),
				Documentation: strings.TrimSpace(symDoc),
			})

		case *ast.ValueSpec:
			for _, name := range s.Names {
				symDoc := doc
				if s.Doc != nil {
					symDoc = s.Doc.Text()
				}

				symbols = append(symbols, Symbol{
					ID:            generateID(),
					Name:          name.Name,
					Kind:          baseKind,
					Language:      "Go",
					Package:       pkgName,
					FilePath:      filePath,
					StartLine:     fset.Position(s.Pos()).Line,
					EndLine:       fset.Position(s.End()).Line,
					Exported:      ast.IsExported(name.Name),
					Documentation: strings.TrimSpace(symDoc),
				})
			}
		}
	}

	return symbols
}
