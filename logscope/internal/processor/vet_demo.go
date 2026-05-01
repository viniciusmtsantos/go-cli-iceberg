// Package processor — vet_demo.go
//
// Este arquivo contém bugs que o `go vet` detecta automaticamente.
// Ele existe para a seção "go vet" da apresentação.
//
// Demonstração:
//
//	go vet ./internal/processor/
//
// Após a demo, corrija os bugs (veja os comentários "FIX:") e rode novamente.
package processor

import (
	"fmt"
	"sync"
)

// badCounter demonstra o erro clássico de copiar um sync.Mutex por valor.
//
// go vet reporta:
//
//	"passes lock by value: processor.badCounter contains sync.Mutex"
//
// Acontece porque o receiver por valor (badCounter) faz uma CÓPIA da struct,
// incluindo o mutex — e cada cópia é um mutex diferente e independente,
// quebrando completamente a sincronização.
//
// FIX: mudar o receiver de `c badCounter` para `c *badCounter`
type badCounter struct {
	mu    sync.Mutex
	count int
}

// BUG: receiver por valor copia o mutex. go vet detecta isso.
func (c badCounter) inc() { // ← deveria ser (c *badCounter)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
}

// badFormat demonstra um erro de formato que o go vet detecta.
//
// go vet reporta:
//
//	"fmt.Sprintf format %s has arg n of wrong type int"
//
// FIX: trocar %s por %d
func badFormat(n int) string {
	return fmt.Sprintf("total de entradas: %s", n) // ← FIX: trocar %s por %d
}

// Evita erros de "declared and not used" para os tipos/funções acima.
var (
	_ = badCounter{}
	_ = badFormat
)
