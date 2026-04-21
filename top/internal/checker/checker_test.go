package checker_test

import (
	"testing"

	"github.com/casadebackend/goprobe/internal/checker"
)

func TestCheck_GoIsInstalled(t *testing.T) {
	goTool := checker.Tool{Name: "Go", Command: "go", Args: []string{"version"}}
	result := checker.Check(goTool)

	if !result.Found {
		t.Fatal("Go deve estar instalado no ambiente de testes")
	}
	if result.Version == "" {
		t.Error("versão não deve ser vazia")
	}
	if result.Latency == 0 {
		t.Error("latência deve ser medida")
	}
}

func TestCheck_FerramentaInexistente(t *testing.T) {
	fake := checker.Tool{
		Name:    "fantasma",
		Command: "binario-que-nao-existe",
		Args:    []string{"--version"},
	}
	result := checker.Check(fake)

	if result.Found {
		t.Error("ferramenta inexistente não deve ser encontrada")
	}
	if result.Version != "" {
		t.Errorf("versão deve ser vazia, got %q", result.Version)
	}
}

func TestCheckAll_RetornaResultadoParaCadaFerramenta(t *testing.T) {
	tools := []checker.Tool{
		{Name: "Go", Command: "go", Args: []string{"version"}},
		{Name: "Fantasma", Command: "nao-existe", Args: []string{"--version"}},
	}

	results := checker.CheckAll(tools)

	if len(results) != len(tools) {
		t.Fatalf("esperado %d resultados, obteve %d", len(tools), len(results))
	}
	if !results[0].Found {
		t.Error("primeiro resultado (Go) deve ser encontrado")
	}
	if results[1].Found {
		t.Error("segundo resultado (fantasma) não deve ser encontrado")
	}
}

func BenchmarkCheck_Go(b *testing.B) {
	goTool := checker.Tool{Name: "Go", Command: "go", Args: []string{"version"}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.Check(goTool)
	}
}

func BenchmarkCheckAll_DefaultTools(b *testing.B) {
	tools := checker.DefaultTools
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.CheckAll(tools)
	}
}
