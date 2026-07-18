package graph

import (
	"testing"
)

func TestImpactsPackageFiltering(t *testing.T) {
	engine := createImpactFixture(t)
	res, err := engine.Impacts("Init", ImpactOptions{PackageFilter: "cli"})
	if err != nil {
		t.Fatal(err)
	}
	
	for _, c := range append(res.DirectCallers, res.IndirectCallers...) {
		if c.Package != "cli" {
			t.Errorf("expected only cli package callers, got %s", c.Package)
		}
	}
}

func TestImpactsMax(t *testing.T) {
	engine := createImpactFixture(t)
	res, err := engine.Impacts("Init", ImpactOptions{Max: 1})
	if err != nil {
		t.Fatal(err)
	}
	
	total := len(res.DirectCallers) + len(res.IndirectCallers)
	if total > 1 {
		t.Errorf("expected at most 1 caller, got %d", total)
	}
}

func TestImpactsSorting(t *testing.T) {
	engine := createImpactFixture(t)
	resName, err := engine.Impacts("Init", ImpactOptions{SortBy: "name"})
	if err != nil {
		t.Fatal(err)
	}
	
	resLine, err := engine.Impacts("Init", ImpactOptions{SortBy: "line"})
	if err != nil {
		t.Fatal(err)
	}
	
	if len(resName.IndirectCallers) > 0 && len(resLine.IndirectCallers) > 0 {
		// Just a smoke test to ensure sorting options don't panic or drop callers
		if len(resName.IndirectCallers) != len(resLine.IndirectCallers) {
			t.Errorf("sorting options should not change number of callers")
		}
	}
}
