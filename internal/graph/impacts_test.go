package graph

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/contact-daniel-me/kyver/internal/analyzer"
	"github.com/contact-daniel-me/kyver/internal/search"
)

// createImpactFixture builds a small multi-package call graph:
//
//	main → Execute → runInit → (*Engine).Init
//	TestInit → runInit
//	Neighbor (same package as Init) does NOT call Init
//	(*Engine).Init recursively calls itself
func createImpactFixture(t *testing.T) *Engine {
	t.Helper()
	tmp := t.TempDir()

	repoDir := filepath.Join(tmp, "repository")
	cliDir := filepath.Join(tmp, "cli")
	os.MkdirAll(repoDir, 0755)
	os.MkdirAll(cliDir, 0755)

	repoFile := filepath.Join(repoDir, "engine.go")
	os.WriteFile(repoFile, []byte(`package repository

type Engine struct{}

func (e *Engine) Init() {
	e.Init()
}

func Neighbor() {
	// does not call Init
}
`), 0644)

	cliFile := filepath.Join(cliDir, "cli.go")
	os.WriteFile(cliFile, []byte(`package cli

import "example.com/app/repository"

func runInit() {
	e := &repository.Engine{}
	e.Init()
}

func Execute() {
	runInit()
}

func main() {
	Execute()
}
`), 0644)

	testFile := filepath.Join(cliDir, "cli_test.go")
	os.WriteFile(testFile, []byte(`package cli

func TestInit() {
	runInit()
}
`), 0644)

	repo := &search.InMemoryRepository{
		Symbols: []analyzer.Symbol{
			{Name: "Engine", Kind: "Struct", Package: "repository", FilePath: repoFile, StartLine: 3, EndLine: 3, Exported: true},
			{Name: "Init", Kind: "Method", Package: "repository", FilePath: repoFile, StartLine: 5, EndLine: 7, Exported: true, Receiver: "*Engine"},
			{Name: "Neighbor", Kind: "Function", Package: "repository", FilePath: repoFile, StartLine: 9, EndLine: 11, Exported: true},
			{Name: "runInit", Kind: "Function", Package: "cli", FilePath: cliFile, StartLine: 5, EndLine: 8, Exported: false},
			{Name: "Execute", Kind: "Function", Package: "cli", FilePath: cliFile, StartLine: 10, EndLine: 12, Exported: true},
			{Name: "main", Kind: "Function", Package: "cli", FilePath: cliFile, StartLine: 14, EndLine: 16, Exported: false},
			{Name: "TestInit", Kind: "Function", Package: "cli", FilePath: testFile, StartLine: 3, EndLine: 5, Exported: true},
		},
		Deps: &analyzer.DependencyGraph{
			FileDeps: map[string][]string{
				repoFile: {"fmt"},
				cliFile:  {"example.com/app/repository"},
				testFile: {"example.com/app/repository"},
			},
		},
		ByName:     make(map[string][]int),
		ByPackage:  make(map[string][]int),
		ByKind:     make(map[string][]int),
		ByReceiver: make(map[string][]int),
		ByLanguage: make(map[string][]int),
		Exported:   []int{},
	}

	for i, sym := range repo.Symbols {
		repo.ByName[sym.Name] = append(repo.ByName[sym.Name], i)
		repo.ByPackage[sym.Package] = append(repo.ByPackage[sym.Package], i)
		if sym.Receiver != "" {
			repo.ByReceiver[sym.Receiver] = append(repo.ByReceiver[sym.Receiver], i)
		}
	}

	e := &Engine{
		repo:         repo,
		pkgImports:   make(map[string]map[string]bool),
		pkgImportsBy: make(map[string]map[string]bool),
	}
	e.buildPackageGraph()
	return e
}

func impactCallerNames(callers []ImpactCaller) map[string]bool {
	m := make(map[string]bool)
	for _, c := range callers {
		m[c.Name] = true
	}
	return m
}

func TestImpactsNoImpact(t *testing.T) {
	engine := createImpactFixture(t)
	res, err := engine.Impacts("Neighbor", ImpactOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.DirectCallers) != 0 || len(res.IndirectCallers) != 0 {
		t.Fatalf("expected zero impact for Neighbor, got direct=%d indirect=%d",
			len(res.DirectCallers), len(res.IndirectCallers))
	}
	if res.Severity != "Low" {
		t.Errorf("expected Low severity, got %s", res.Severity)
	}
}

func TestImpactsSingleCaller(t *testing.T) {
	engine := createImpactFixture(t)

	// Isolate: only runInit calls Init directly among non-transitive walk
	// when we stop at direct — Verify Init has runInit as a direct caller.
	res, err := engine.Impacts("Init", ImpactOptions{})
	if err != nil {
		t.Fatal(err)
	}

	direct := impactCallerNames(res.DirectCallers)
	if !direct["runInit"] {
		t.Fatalf("expected runInit as direct caller, got %+v", res.DirectCallers)
	}
	if direct["Neighbor"] {
		t.Fatal("Neighbor must not be marked impacted (same package, no call)")
	}
	if direct["Execute"] || direct["main"] {
		t.Fatal("Execute/main must be indirect, not direct")
	}

	for _, dc := range res.DirectCallers {
		if dc.Name == "runInit" {
			if len(dc.Chain) == 0 || !strings.Contains(dc.Chain[0], "directly calls") {
				t.Errorf("runInit why-chain: %v", dc.Chain)
			}
			if !strings.Contains(dc.Chain[0], "Init") {
				t.Errorf("runInit chain should mention Init: %v", dc.Chain)
			}
		}
	}
}

func TestImpactsTransitiveCallers(t *testing.T) {
	engine := createImpactFixture(t)
	res, err := engine.Impacts("Init", ImpactOptions{})
	if err != nil {
		t.Fatal(err)
	}

	indirect := impactCallerNames(res.IndirectCallers)
	if !indirect["Execute"] {
		t.Fatalf("expected Execute as indirect caller, got %+v", res.IndirectCallers)
	}
	if !indirect["main"] {
		t.Fatalf("expected main as indirect caller, got %+v", res.IndirectCallers)
	}

	for _, ic := range res.IndirectCallers {
		switch ic.Name {
		case "Execute":
			if len(ic.Chain) == 0 || !strings.Contains(ic.Chain[0], "runInit") {
				t.Errorf("Execute should explain calls runInit, got %v", ic.Chain)
			}
		case "main":
			if len(ic.Chain) == 0 || !strings.Contains(ic.Chain[0], "Execute") {
				t.Errorf("main should explain calls Execute, got %v", ic.Chain)
			}
		}
	}
}

func TestImpactsRecursiveCalls(t *testing.T) {
	engine := createImpactFixture(t)
	res, err := engine.Impacts("Init", ImpactOptions{})
	if err != nil {
		t.Fatal(err)
	}

	// Self-recursion must not appear as a caller.
	for _, c := range append(res.DirectCallers, res.IndirectCallers...) {
		if c.Name == "Init" {
			t.Fatal("recursive self-call must not appear as an impacted caller")
		}
	}
}

func TestImpactsPackageBoundaries(t *testing.T) {
	engine := createImpactFixture(t)
	res, err := engine.Impacts("Init", ImpactOptions{})
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range append(res.DirectCallers, res.IndirectCallers...) {
		if c.Name == "Neighbor" {
			t.Fatal("same-package Neighbor must not be impacted without a call edge")
		}
	}

	pkgSet := make(map[string]bool)
	for _, p := range res.AffectedPackages {
		pkgSet[p] = true
	}
	if !pkgSet["cli"] {
		t.Errorf("expected cli in affected packages, got %v", res.AffectedPackages)
	}
	// repository should not appear unless a caller lives there (none do, besides skipped recursion)
	if pkgSet["repository"] {
		t.Errorf("repository should not be an affected package from callers, got %v", res.AffectedPackages)
	}
}

func TestImpactsSeverityCalculation(t *testing.T) {
	tests := []struct {
		direct, indirect, pkgs, public int
		expected                       string
	}{
		{0, 0, 0, 0, "Low"},
		{0, 5, 2, 1, "Low"}, // no direct → Low
		{1, 0, 1, 1, "Medium"},
		{2, 4, 1, 0, "Medium"},
		{2, 5, 1, 0, "High"}, // transitive threshold
		{3, 0, 1, 0, "High"},
		{1, 0, 2, 0, "High"},
		{5, 0, 3, 1, "Critical"},
		{1, 10, 3, 1, "Critical"}, // transitive critical path
		{5, 0, 3, 0, "High"},      // missing public API → not Critical
	}

	for _, tt := range tests {
		got, _ := computeImpactSeverity(tt.direct, tt.indirect, tt.pkgs, tt.public)
		if got != tt.expected {
			t.Errorf("D=%d T=%d P=%d A=%d: want %s, got %s",
				tt.direct, tt.indirect, tt.pkgs, tt.public, tt.expected, got)
		}
	}
}

func TestImpactsJSONSchema(t *testing.T) {
	engine := createImpactFixture(t)
	res, err := engine.Impacts("Init", ImpactOptions{})
	if err != nil {
		t.Fatal(err)
	}

	b, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	jsonStr := string(b)

	required := []string{
		`"symbol":`,
		`"directCallers":`,
		`"indirectCallers":`,
		`"affectedFiles":`,
		`"affectedPackages":`,
		`"severity":`,
		`"stats":`,
		`"lookupTimeMs":`,
		`"tests":`,
	}
	for _, field := range required {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("missing JSON field %s in: %s", field, jsonStr)
		}
	}

	// Ensure lowerCamelCase — no snake_case impact keys.
	for _, bad := range []string{`"direct_callers"`, `"affected_files"`, `"lookup_time_ms"`} {
		if strings.Contains(jsonStr, bad) {
			t.Errorf("unexpected snake_case key %s", bad)
		}
	}
}

func TestImpactsTestsSection(t *testing.T) {
	engine := createImpactFixture(t)
	res, err := engine.Impacts("Init", ImpactOptions{})
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, tc := range res.Tests {
		if tc.Name == "TestInit" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected TestInit in Tests, got %+v", res.Tests)
	}
}

func TestImpactsTreeFormat(t *testing.T) {
	engine := createImpactFixture(t)
	res, err := engine.Impacts("Init", ImpactOptions{TreeFormat: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.DirectCallers) == 0 {
		t.Fatal("expected callers for tree")
	}
	// Smoke: tree printer should not panic.
	PrintImpactResult(res, ImpactOptions{TreeFormat: true}, false)
}
