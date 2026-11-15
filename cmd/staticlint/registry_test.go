package main

import (
	"strings"
	"testing"

	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func TestAnalyzerRegistry(t *testing.T) {
	seen := make(map[string]bool, len(analyzers))
	counts := map[string]int{"SA": 0, "ST": 0}
	for _, analyzer := range analyzers {
		if seen[analyzer.Name] {
			t.Fatalf("duplicate analyzer %q", analyzer.Name)
		}
		seen[analyzer.Name] = true
		for _, family := range []string{"SA", "ST"} {
			if strings.HasPrefix(analyzer.Name, family) {
				counts[family]++
			}
		}
	}
	if counts["SA"] < 80 {
		t.Errorf("registered %d SA analyzers; want the complete Staticcheck SA family", counts["SA"])
	}
	if counts["ST"] == 0 {
		t.Error("no Staticcheck non-SA analyzers are registered")
	}
	for _, analyzer := range staticcheck.Analyzers {
		if !seen[analyzer.Analyzer.Name] {
			t.Errorf("Staticcheck analyzer %q is missing", analyzer.Analyzer.Name)
		}
	}
	for _, analyzer := range stylecheck.Analyzers {
		if !seen[analyzer.Analyzer.Name] {
			t.Errorf("Staticcheck style analyzer %q is missing", analyzer.Analyzer.Name)
		}
	}
	for _, name := range []string{"osexit", "nilerr", "forcetypeassert", "printf", "shadow"} {
		if !seen[name] {
			t.Errorf("required analyzer %q is not registered", name)
		}
	}
}
