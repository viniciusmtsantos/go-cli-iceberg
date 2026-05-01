# logscope — Projeto único da apresentação

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transformar o `logscope` no único projeto da apresentação "iceberg da CLI do Go", com README roteiro completo (24 blocos, 4 camadas), `reset.sh` e fuzz test.

**Architecture:** O `vet_demo.go` já existente é ajustado para ter bugs reais detectáveis pelo `go vet`. Um fuzz test é adicionado ao parser. O `reset.sh` usa Python inline para injetar o bug de `go fmt` em `processor.go` e restaura `vet_demo.go` e `go.mod` via `git checkout`. O `README.md` é o roteiro completo da apresentação, seguindo o estilo dos projetos `top/` e `layer2/`.

**Tech Stack:** Go 1.25+, bash, python3 (inline no reset.sh)

---

## Mapa de arquivos

| Operação | Arquivo | Responsabilidade |
|----------|---------|-----------------|
| Modify | `logscope/internal/processor/vet_demo.go` | Ativar bug de `go vet` em `badFormat` (`%d` → `%s`) |
| Create | `logscope/internal/parser/parser_fuzz_test.go` | Fuzz test para demo de `go test -fuzz` |
| Create | `logscope/reset.sh` | Restaura estado inicial: bugs + artefatos + go.mod |
| Create | `logscope/README.md` | Roteiro completo da apresentação (24 blocos) |

---

## Task 1: Ativar bug real em `vet_demo.go`

**Files:**
- Modify: `logscope/internal/processor/vet_demo.go:49`

O `badFormat` atualmente usa `%d` (correto), mas o comentário documenta que o bug é `%s` para `int`.
Mudar para `%s` torna o bug real — `go vet ./...` vai reportar `fmt.Sprintf format %s has arg n of wrong type int`.

- [ ] **Step 1: Aplicar a mudança**

Em `logscope/internal/processor/vet_demo.go`, linha 49, mudar:
```go
// antes (correto — sem bug de vet):
return fmt.Sprintf("total de entradas: %d", n) // ← deveria ser %d

// depois (bug intencional para go vet detectar):
return fmt.Sprintf("total de entradas: %s", n) // ← FIX: trocar %s por %d
```

- [ ] **Step 2: Verificar que `go vet` detecta os dois bugs**

```bash
cd logscope
go vet ./...
```

Saída esperada (duas linhas, qualquer ordem):
```
# github.com/gopher/logscope/internal/processor
internal/processor/vet_demo.go:35:6: copylocks: func (c badCounter) inc() passes lock by value: github.com/gopher/logscope/internal/processor.badCounter contains sync.Mutex
internal/processor/vet_demo.go:49:29: fmt.Sprintf format %s has arg n of wrong type int
```

- [ ] **Step 3: Verificar que os testes ainda passam (o vet_demo não é executado pelos testes)**

```bash
go test ./...
```

Saída esperada: todos os testes passam (vet_demo.go não tem testes).

- [ ] **Step 4: Commit**

```bash
cd ..
git add logscope/internal/processor/vet_demo.go
git commit -m "fix(vet_demo): ativar bug real em badFormat (%d → %s) para demo go vet

Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>"
```

---

## Task 2: Criar fuzz test

**Files:**
- Create: `logscope/internal/parser/parser_fuzz_test.go`

- [ ] **Step 1: Criar o arquivo**

Conteúdo completo de `logscope/internal/parser/parser_fuzz_test.go`:

```go
package parser_test

import (
	"strings"
	"testing"

	"github.com/gopher/logscope/internal/parser"
)

// FuzzParseReader verifies that the parser never panics on arbitrary input.
//
// Run with:
//
//	go test -fuzz=FuzzParseReader -fuzztime=10s ./internal/parser/
//
// The fuzzer mutates the seed corpus — inserting random bytes, truncating lines,
// injecting unicode — and reports any input that causes a panic.
func FuzzParseReader(f *testing.F) {
	// Seed corpus — representative log lines
	f.Add("2026-04-26T10:00:00Z INFO GET /api/users 200 145 1234\n")
	f.Add("2026-04-26T10:00:01Z WARN POST /api/auth/login 401 22 0\n")
	f.Add("2026-04-26T10:00:02Z ERROR DELETE /api/orders/99 500 310 0\n")
	f.Add("malformed line that should be skipped\n")
	f.Add("") // empty input — must return empty slice, not panic

	f.Fuzz(func(t *testing.T, data string) {
		// The parser must never panic regardless of input.
		// Errors are expected and acceptable; panics are bugs.
		_, _ = parser.ParseReader(strings.NewReader(data))
	})
}
```

- [ ] **Step 2: Verificar que os testes existentes ainda passam**

```bash
cd logscope
go test ./internal/parser/
```

Saída esperada: `ok  github.com/gopher/logscope/internal/parser`

- [ ] **Step 3: Rodar o fuzz por 5 segundos para confirmar que não há panic**

```bash
go test -fuzz=FuzzParseReader -fuzztime=5s ./internal/parser/
```

Saída esperada (sem falhas):
```
fuzz: elapsed: 5s, execs: ..., new interesting: ..., coverage: ...
PASS
ok  github.com/gopher/logscope/internal/parser  5.XXXs
```

- [ ] **Step 4: Commit**

```bash
cd ..
git add logscope/internal/parser/parser_fuzz_test.go
git commit -m "test(parser): adiciona FuzzParseReader para demo go test -fuzz

Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>"
```

---

## Task 3: Criar `reset.sh`

**Files:**
- Create: `logscope/reset.sh`

O `reset.sh` prepara o `logscope` para a apresentação:
1. Remove binários e artefatos
2. Restaura `vet_demo.go` (caso tenha sido corrigido durante demo anterior)
3. Restaura `go.mod` ao estado sem deps externas
4. Injeta o bug de `go fmt` em `processor.go` (função `Avg` sem indentação)
5. Imprime a sequência completa da apresentação

- [ ] **Step 1: Criar o arquivo `logscope/reset.sh`**

Conteúdo completo:

```bash
#!/usr/bin/env bash
# reset.sh — restaura o logscope ao estado inicial para treino da apresentação.
# Introduz propositalmente: (1) formatação quebrada em processor.go e (2) bugs de vet em vet_demo.go.
set -euo pipefail

echo "[reset] removendo binários e arquivos gerados..."
rm -f logscope logscope.exe logscope-json logscope-arm64 logscope-safe

echo "[reset] removendo artefatos de profiling e análise..."
rm -f cpu.prof mem.prof trace.out coverage.out access.log

echo "[reset] limpando cache de testes..."
go clean -testcache 2>/dev/null || true

echo "[reset] restaurando go.mod ao estado sem dependências externas..."
cat > go.mod << 'EOF'
module github.com/gopher/logscope

go 1.25.0
EOF
rm -f go.sum

echo "[reset] restaurando vet_demo.go ao estado com bugs intencionais..."
git checkout HEAD -- internal/processor/vet_demo.go

echo "[reset] introduzindo formatação quebrada em processor.go (função Avg)..."
python3 - << 'PYEOF'
content = open("internal/processor/processor.go").read()
old = '''// Avg returns the mean latency across all entries.
func (s *Stats) Avg() time.Duration {
\tif len(s.Latencies) == 0 {
\t\treturn 0
\t}
\tvar total time.Duration
\tfor _, l := range s.Latencies {
\t\ttotal += l
\t}
\treturn total / time.Duration(len(s.Latencies))
}'''
new = '''// Avg returns the mean latency across all entries.
func (s *Stats) Avg() time.Duration {
if len(s.Latencies) == 0 {return 0}
var total time.Duration
for _,l := range s.Latencies {total += l}
return total/time.Duration(len(s.Latencies))
}'''
result = content.replace(old, new)
if result == content:
    print("AVISO: substituição do fmt bug não encontrou o trecho esperado — processor.go pode já estar com bug ou formatação diferente")
else:
    open("internal/processor/processor.go", "w").write(result)
    print("ok")
PYEOF

echo ""
echo "✓ Pronto! Estado inicial restaurado com dois bugs intencionais:"
echo ""
echo "  ① processor.go → Avg() sem indentação (go fmt vai corrigir)"
echo "  ② vet_demo.go  → badFormat com %s para int + badCounter com receiver por valor (go vet vai detectar)"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "SEQUÊNCIA DA APRESENTAÇÃO"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🌊 SUPERFÍCIE"
echo "  1.  go help / go help <cmd>                         # abertura"
echo "  2.  go version"
echo "  3.  go version -m -json \$(which docker)"
echo "  4.  cat go.mod"
echo "  5.  go get github.com/fatih/color@latest"
echo "  6.  cat go.mod && cat go.sum | head -5"
echo "  7.  go test ./..."
echo "  8.  go test -v ./internal/parser/"
echo "  9.  go run ./cmd/logscope -gen -lines 1000 -output access.log"
echo "  10. go run ./cmd/logscope -input access.log"
echo "  11. go run golang.org/x/vuln/cmd/govulncheck@latest ."
echo "  12. go build -o logscope ./cmd/logscope && ls -lh logscope"
echo "  13. ./logscope -gen -lines 100 | ./logscope"
echo "  14. go version -m ./logscope"
echo "  15. go env GOPATH && go install -n ./cmd/logscope"
echo "  16. go install ./cmd/logscope"
echo "  17. \$(go env GOPATH)/bin/logscope -version"
echo ""
echo "🔎 CAMADA 2 — QUALIDADE"
echo "  18. go doc fmt.Printf"
echo "  19. go doc ./pkg/entry LogEntry"
echo "  20. go doc ./internal/parser ParseReader"
echo "  21. go doc ./internal/processor ProcessSafe"
echo "  22. gofmt -l ."
echo "  23. gofmt -d internal/processor/processor.go"
echo "  24. go fmt ./... && gofmt -l ."
echo "  25. go vet ./..."
echo "  26. # corrige vet_demo.go: badCounter receiver + %s→%d"
echo "  27. go vet ./..."
echo "  28. go env && go env -json GOPATH GOCACHE GOMODCACHE"
echo "  29. go env -w GOTELEMETRY=off && cat \$(go env GOENV)"
echo "  30. go env -u GOTELEMETRY"
echo "  31. go list ./... && go list -m -u all | grep '\\['"
echo "  32. go mod why golang.org/x/sys && go mod why github.com/fatih/color"
echo "  33. go test -cover ./internal/..."
echo "  34. go test -coverprofile=coverage.out ./internal/... && go tool cover -func=coverage.out"
echo "  35. go tool cover -html=coverage.out"
echo "  36. go mod tidy && cat go.mod    # remove color (não usada no código)"
echo "  37. go clean -testcache && go test ./internal/..."
echo "  38. go clean -cache && go clean -modcache"
echo ""
echo "⚙️  CAMADA 3 — BUILD AVANÇADO E GERAÇÃO"
echo "  39. cat pkg/entry/entry.go      # mostra //go:generate"
echo "  40. rm pkg/entry/level_string.go && go generate ./pkg/entry/ && cat pkg/entry/level_string.go"
echo "  41. cat internal/report/report.go && cat internal/report/report_json.go"
echo "  42. go build -tags json -o logscope-json ./cmd/logscope"
echo "  43. ./logscope-json -gen -lines 100 | ./logscope-json"
echo "  44. go build -ldflags \"-X main.version=1.2.0 -X main.commit=\$(git rev-parse --short HEAD)\" -o logscope ./cmd/logscope"
echo "  45. ./logscope -version && go version -m ./logscope"
echo "  46. GOOS=windows go build -o logscope.exe ./cmd/logscope && file logscope.exe"
echo "  47. GOOS=linux GOARCH=arm64 go build -o logscope-arm64 ./cmd/logscope && file logscope-arm64"
echo "  48. go build -trimpath -o logscope-safe ./cmd/logscope"
echo "  49. strings ./logscope-safe | grep \$(pwd) || echo '✓ nenhum path local'"
echo ""
echo "🔬 CAMADA 4 — TESTING PROFUNDO E OBSERVABILIDADE"
echo "  50. go test -bench=. -benchmem ./internal/parser/"
echo "  51. go test -bench=BenchmarkProcessSafe -benchmem ./internal/processor/"
echo "  52. go test -run TestNaiveRace ./internal/processor/"
echo "  53. go test -race -run TestNaiveRace ./internal/processor/"
echo "  54. go test -race -run TestProcessSafe_DeterministicWithRace ./internal/processor/"
echo "  55. go test -fuzz=FuzzParseReader -fuzztime=10s ./internal/parser/"
echo "  56. go test -shuffle=on -v ./internal/parser/"
echo "  57. go test -count=3 ./internal/..."
echo "  58. logscope -gen -lines 200000 -output access.log"
echo "  59. logscope -input access.log -cpuprofile cpu.prof && go tool pprof -http=:8080 cpu.prof"
echo "  60. logscope -input access.log -memprofile mem.prof && go tool pprof -http=:8081 mem.prof"
echo "  61. logscope -input access.log -trace trace.out && go tool trace trace.out"
echo "  62. go tool nm ./logscope | head -20"
echo "  63. go tool nm ./logscope | wc -l"
echo ""
```

- [ ] **Step 2: Tornar o script executável**

```bash
cd logscope
chmod +x reset.sh
```

- [ ] **Step 3: Testar o reset.sh — verificar que injeta o bug de fmt**

```bash
cd logscope
./reset.sh
gofmt -l .
```

Saída esperada: `internal/processor/processor.go` aparece na lista (bug injetado com sucesso).

- [ ] **Step 4: Verificar que o bug de vet também está presente**

```bash
go vet ./...
```

Saída esperada: dois erros em `vet_demo.go` (copylocks + format string).

- [ ] **Step 5: Verificar que o código compila mesmo com os bugs**

```bash
go build ./...
```

Saída esperada: sem erros (bugs são de fmt/vet, não de compilação).

- [ ] **Step 6: Reverter processor.go ao estado limpo**

```bash
git checkout HEAD -- internal/processor/processor.go
```

- [ ] **Step 7: Commit**

```bash
cd ..
git add logscope/reset.sh
git commit -m "feat(logscope): adiciona reset.sh para preparar estado da apresentação

Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>"
```

---

## Task 4: Criar `README.md`

**Files:**
- Create: `logscope/README.md`

Este é o roteiro completo da apresentação. Segue o estilo dos projetos `top/` e `layer2/`:
narrador em `> _"itálico"_`, comandos em blocos `bash`, referências ao `go help` em cada bloco.

- [ ] **Step 1: Criar o arquivo `logscope/README.md`** com o seguinte conteúdo:

```markdown
# logscope — O Iceberg da CLI do Go

> **Antes de começar:** rode `./reset.sh` no terminal (não mostrar para o público).

---

## Abertura — `go help`

> Sem terminal. Só fala.

- _"Antes de qualquer comando, quero te dar um atalho que a maioria ignora."_
- `go help` → lista todos os comandos **e** tópicos de conceito do toolchain
- `go help <command>` → doc completa de qualquer comando, no terminal
- `go help <topic>` → conceitos como `modules`, `buildconstraint`, `testflag`
- _"Cada flag que aparecer daqui pra frente tem `go help` atrás. Guarda isso."_

---

## 🌊 Superfície — comandos do dia a dia

### Bloco 0 — O projeto

> Sem terminal. Abre o editor e mostra o código.

- _"`logscope` — um analisador de logs HTTP estruturados. Construído para essa apresentação porque exige de nós exatamente os comandos que quero mostrar."_
- O que ele faz: lê logs de acesso HTTP, calcula percentis de latência, distribuição de status e top endpoints
- Dois modos: `-gen` gera dados sintéticos; sem flags, analisa de stdin ou arquivo
- Pipeline-friendly: `logscope -gen -lines 50000 | logscope`
- Estrutura em três camadas:
  - `cmd/logscope/` → entrypoint com flags, profiling e tracing
  - `internal/` → parser, processor (safe + naive), report
  - `pkg/entry/` → tipos compartilhados + código gerado por `go generate`
- _"O projeto está aqui. Sem binário, sem artefatos. Vamos resolver isso agora."_

---

### Bloco 1 — `go version`

- Roda `go version` → versão do runtime
- _"Mas olha o que `-m` e `-json` fazem juntos — descobrimos isso no `go help version`."_

```bash
go version
go version -m -json $(which docker)
```

> _"Útil para auditar dependências de um binário em produção sem ter o código-fonte. Vamos guardar para depois que compilarmos o logscope."_

---

### Bloco 2 — `go.mod` + `go get` + `go mod tidy`

> _"Antes de rodar qualquer coisa, o projeto precisa de um módulo."_

```bash
cat go.mod
```

> _"Esse arquivo é o contrato do projeto. Só `module` e `go version` — sem dependências ainda. O nome do módulo é o caminho de import, mesmo sem estar no GitHub."_

> _"Vou adicionar a lib `fatih/color` para colorir o terminal."_

```bash
go get github.com/fatih/color@latest
cat go.mod
cat go.sum | head -5
```

> _"O `go.sum` é o lockfile: hash criptográfico de cada módulo. Ninguém substitui uma dep sem o Go perceber."_
> _"Olha a diretiva `go` no `go.mod` — ela pode ter subido. Isso não é a versão instalada, é a versão mínima que o grafo de dependências exige."_
> _"Não vou chamar `go mod tidy` agora — vamos usar essa dep nos próximos blocos."_

---

### Bloco 3 — `go test`

> _"Módulo configurado. Antes de rodar qualquer coisa, quero saber se o código está de pé."_

```bash
go test ./...
go test -v ./internal/parser/
```

> _"`go test` é o orquestrador nativo — sem pytest, sem jest. `-v` mostra cada teste com PASS/FAIL e tempo. `go help testflag` lista todas as flags disponíveis."_

---

### Bloco 4 — `go run`

> _"Testes passando. Agora quero ver o projeto rodando — mas sem compilar ainda."_

```bash
go run ./cmd/logscope -gen -lines 1000 -output access.log
go run ./cmd/logscope -input access.log
```

> _"`go run` executa direto na memória, sem gerar binário. Agora uma flag pouco conhecida — está em `go help run`:"_

```bash
go run golang.org/x/vuln/cmd/govulncheck@latest .
```

> _"O Go ignora o `go.mod` atual e roda em modo isolado. Ferramenta externa sem contaminar suas dependências."_

---

### Bloco 5 — `go build`

> _"Vamos compilar para a minha plataforma."_

```bash
go build -o logscope ./cmd/logscope
ls -lh logscope
./logscope -gen -lines 100 | ./logscope
```

> _"Nativo. Sem runtime, sem JVM. A flag `-o` nomeia o binário — `go help build` lista várias outras: `-race`, `-ldflags`, `-tags`."_
> _"E agora podemos fechar o ciclo do Bloco 1:"_

```bash
go version -m ./logscope
```

> _"Inspecionamos o binário que compilamos — os mesmos metadados que vimos no Docker."_

---

### Bloco 6 — `go install`

> _"`go build` gera o binário aqui na pasta. `go install` compila e move para `$GOPATH/bin` — disponível globalmente."_

```bash
go env GOPATH
go install -n ./cmd/logscope
```

> _"O `-n` imprime os passos sem executar: mkdir temporário, compilação, link, mv para `$GOPATH/bin`. O toolchain sem segredos. Está em `go help install`."_

```bash
go install ./cmd/logscope
ls -lh $(go env GOPATH)/bin/logscope
$(go env GOPATH)/bin/logscope -version
```

---

## 🔎 Camada 2 — Qualidade e dependências

### Bloco 7 — `go doc`

> _"Antes de rodar qualquer coisa — como você descobriria o que `ParseReader` faz sem abrir o navegador?"_

```bash
go doc fmt.Printf
```

> _"Parâmetros, tipos, comportamento. No terminal, offline. Agora no nosso próprio código:"_

```bash
go doc ./pkg/entry
go doc ./pkg/entry LogEntry
go doc ./internal/parser ParseReader
go doc ./internal/processor ProcessSafe
```

> _"Viu os exemplos de uso nos comentários? O `go doc` renderiza isso. É a mesma documentação do `pkg.go.dev` — só que local e instantânea."_

---

### Bloco 8 — `go fmt`

> _"Abre o `internal/processor/processor.go` no editor. A função `Avg` está sem indentação. O código compila. O binário funciona. Mas o `go fmt` não aprova."_

```bash
gofmt -l .
```

> _"Listou o `processor.go`. Não é um erro — é uma diferença em relação ao padrão canônico do Go. Antes de corrigir, veja exatamente o que vai mudar:"_

```bash
gofmt -d internal/processor/processor.go
```

> _"Verde é o que vai entrar, vermelho é o que vai sair. Agora aplica:"_

```bash
go fmt ./...
gofmt -l .
```

> _"Silêncio. Zero configuração, zero debate — o `go fmt` tem uma opinião e ela é definitiva. Em CI: `gofmt -l . | grep .` retorna exit 1 se qualquer arquivo estiver fora do padrão."_

---

### Bloco 9 — `go vet`

> _"Formatação resolvida. Mas o projeto ainda tem bugs. Não são de sintaxe — o compilador não detecta. Vamos chamar o caçador:"_

```bash
go vet ./...
```

> _"Dois bugs em `vet_demo.go`:"_
> _"1. `badCounter.inc()` com receiver por valor — cada chamada copia o `sync.Mutex`, quebrando a sincronização."_
> _"2. `badFormat` usa `%s` para formatar um `int` — o log vai imprimir lixo em produção."_

> _"`go vet` usa a mesma `go/ast` que o nosso parser usa internamente. Ele inspecionou a árvore sintática e cruzou os tipos dos argumentos com os verbos do format string."_
> _"Corrige: receiver `badCounter` → `*badCounter`, `%s` → `%d`."_

```bash
# edita vet_demo.go: linha do receiver (badCounter → *badCounter) e linha do Sprintf (%s → %d)
go vet ./...
```

> _"Silêncio. O projeto está limpo."_

---

### Bloco 10 — `go env`

> _"Agora um desvio: o logscope não tem um comando `env`, mas o toolchain tem — e ele é mais poderoso do que parece."_

```bash
go env
go env GOPATH GOCACHE GOMODCACHE
go env -json GOPATH GOCACHE GOMODCACHE
```

> _"`GOMODCACHE` é onde ficam os módulos baixados. `GOCACHE` é o cache de builds. Ambos fora do projeto — múltiplos projetos compartilham o mesmo cache."_
> _"Mas tem um superpoder aqui que quase ninguém conhece:"_

```bash
go env -w GOTELEMETRY=off
go env GOENV
cat $(go env GOENV)
```

> _"`go env -w` escreveu numa config global do Go — persiste entre projetos, terminais e reboots. Sem `.bashrc`, sem `.zshrc`, sem configurar de novo no próximo setup."_

```bash
go env -u GOTELEMETRY
```

---

### Bloco 11 — `go list`

> _"Projeto limpo, ambiente configurado. Agora: você sabe exatamente o que está carregando como dependência?"_

```bash
go list ./...
go list -m all
go list -m -u all
```

> _"Viu o `[v...]` ao lado de `fatih/color`? É a versão mais nova disponível. Radar de desatualização sem sair do terminal."_

```bash
go list -m -u all | grep '\['
go list -m -u -json all
```

> _"O JSON é para automação. Num pipeline de CI: se alguma dep estiver N versões atrás, falha o build."_

---

### Bloco 12 — `go mod why`

> _"O `go list` mostrou `golang.org/x/sys` na lista. Você não adicionou essa dep. De onde ela veio?"_

```bash
go mod why golang.org/x/sys
```

> _"Imprimiu o caminho mais curto no grafo: `logscope` → `fatih/color` → `go-isatty` → `golang.org/x/sys`. Quatro saltos. Você vê a rota inteira."_

```bash
go mod graph | grep sys
go mod why github.com/fatih/color
```

> _"Essencial em projetos grandes — quando aparece uma lib desconhecida no `go.sum`, esse comando te diz quem a trouxe."_

---

### Bloco 13 — `go test -cover`

> _"Os testes passam. Mas quanto do código está sendo testado?"_

```bash
go test -cover ./internal/...
```

> _"Tem o percentual de cobertura. Mas onde estão as linhas descobertas? Para saber, precisa de um perfil:"_

```bash
go test -coverprofile=coverage.out ./internal/...
go tool cover -func=coverage.out
```

> _"Por função. Dá para ver exatamente qual caminho de código nunca foi exercitado. E tem o HTML:"_

```bash
go tool cover -html=coverage.out
```

> _"`go tool cover` já vem no toolchain — sem instalar nada. Linhas vermelhas = não cobertas. É a forma mais visual de entender o que falta testar."_

---

### Bloco 14 — `go clean`

> _"Antes de fechar essa camada: dependências e cache."_

> _"Adicionamos `fatih/color` nos Blocos 2–12. Mas o logscope não usa essa dep no código. O `go mod tidy` detecta isso:"_

```bash
go mod tidy
cat go.mod
```

> _"Removeu automaticamente. Esse é o comportamento correto: `go mod tidy` limpa tudo que não é importado de verdade."_

> _"Agora o cache:"_

```bash
du -sh $(go env GOCACHE)
du -sh $(go env GOMODCACHE)
go clean -testcache
go test ./internal/...
```

> _"Com `-testcache` limpo, todos os testes rodam de novo — mesmo sem mudança de código. Útil quando você suspeita que um resultado está sendo servido do cache."_

```bash
go clean -cache
```

> _"`-cache` limpa os objetos compilados — o próximo build recompila tudo do zero. E o nuclear:"_

```bash
du -sh $(go env GOMODCACHE)
go clean -modcache
du -sh $(go env GOMODCACHE)
```

> _"Apagou os módulos baixados. Libera espaço considerável em projetos grandes. Use quando suspeitar de uma dep corrompida ou quiser garantir um build 100% limpo."_

---

## ⚙️ Camada 3 — Build avançado e geração

### Bloco 15 — `go generate`

> _"O logscope tem um tipo `Level` com constantes: DEBUG, INFO, WARN, ERROR, FATAL. Como o método `String()` é gerado?"_

```bash
cat pkg/entry/entry.go
```

> _"O `//go:generate` é um comentário especial — não é executado automaticamente. `go generate` lê todos os arquivos do pacote e executa as diretivas encontradas."_

```bash
cat pkg/entry/level_string.go
```

> _"Esse arquivo foi gerado pelo `stringer` — nunca edite manualmente. Vamos apagá-lo e regenerar:"_

```bash
rm pkg/entry/level_string.go
go generate ./pkg/entry/
cat pkg/entry/level_string.go
```

> _"O `stringer` varreu o tipo `Level` e gerou o `String()` automaticamente — com bounds check incluído. `go help generate` documenta o mecanismo. `go generate ./...` roda todas as diretivas do projeto."_

---

### Bloco 16 — Build tags + `-ldflags`

> _"O logscope tem dois modos de output: texto e JSON. Como o compilador escolhe qual incluir?"_

```bash
cat internal/report/report.go      # //go:build !json
cat internal/report/report_json.go # //go:build json
```

> _"A linha `//go:build json` faz o compilador incluir esse arquivo apenas quando a tag `json` estiver ativa. `go help buildconstraint` lista todos os valores válidos."_

```bash
go build -o logscope ./cmd/logscope
./logscope -gen -lines 100 | ./logscope

go build -tags json -o logscope-json ./cmd/logscope
./logscope-json -gen -lines 100 | ./logscope-json
```

> _"Mesmo código-fonte, comportamento diferente. Sem if/else — o compilador inclui ou exclui arquivos inteiros."_

> _"Agora injetando variáveis em tempo de build via `-ldflags`:"_

```bash
go build \
  -ldflags "-X main.version=1.2.0 -X main.commit=$(git rev-parse --short HEAD)" \
  -o logscope \
  ./cmd/logscope

./logscope -version
go version -m ./logscope
```

> _"O `-X` reescreve variáveis `var` no binário. Zero `os.Getenv`, zero arquivo de config — a versão está gravada no binário. `go help build` documenta as `-ldflags`."_

---

### Bloco 17 — Cross-compilation + `-trimpath`

> _"O Go compila para qualquer plataforma de qualquer plataforma. Sem Docker, sem VM."_

```bash
GOOS=windows go build -o logscope.exe ./cmd/logscope
ls -lh logscope.exe
file logscope.exe

GOOS=linux GOARCH=arm64 go build -o logscope-arm64 ./cmd/logscope
ls -lh logscope-arm64
file logscope-arm64
```

> _"Dois binários para plataformas que nem temos aqui. `go help environment` lista todos os valores válidos de GOOS e GOARCH."_

> _"Agora um detalhe de segurança que poucos conhecem: o path local pode vazar no binário."_

```bash
strings ./logscope | grep $(pwd) | head -3
go build -trimpath -o logscope-safe ./cmd/logscope
strings ./logscope-safe | grep $(pwd) || echo "✓ nenhum path local no binário"
```

> _"`-trimpath` remove todos os paths absolutos do binário — builds reproduzíveis, sem leak de estrutura interna do servidor. `go help build` documenta essa flag."_

---

## 🔬 Camada 4 — Testing profundo e observabilidade

### Bloco 18 — Benchmarks

> _"Os testes passam. Mas quão rápido é o parser? Quão rápido é o processor?"_

```bash
go test -bench=. -benchmem ./internal/parser/
```

> _"Três colunas que importam: iterações | `ns/op` (latência por operação) | `B/op` e `allocs/op` (pressão no GC)."_

> _"Agora comparando workers no processor:"_

```bash
go test -bench=BenchmarkProcessSafe -benchmem ./internal/processor/
```

> _"ProcessSafe com 1 worker vs 8 workers — você vê o speedup à medida que adicionamos goroutines. `go help testflag` documenta todas as flags de benchmark."_

---

### Bloco 19 — `go test -race`

> _"O logscope tem dois processors: `ProcessSafe` (correto) e `ProcessNaive` (com corrida de dados intencional). Vamos ver a diferença."_

```bash
go test -run TestNaiveRace ./internal/processor/
```

> _"Pode ou não crashar — corrida de dados é undefined behaviour, resultado não-determinístico. Agora com o detector:"_

```bash
go test -race -run TestNaiveRace ./internal/processor/
```

> _"O detector imprimiu os goroutines em conflito, com stack traces. Nenhuma mágica — o compilador instrumenta cada acesso à memória em tempo de build."_

> _"Agora confirmamos que `ProcessSafe` é livre de corridas:"_

```bash
go test -race -run TestProcessSafe_DeterministicWithRace ./internal/processor/
```

> _"Silêncio. `go help testflag` documenta o `-race`."_

---

### Bloco 20 — `go test -fuzz`

> _"Cobertura de testes garante que o código funciona para as entradas que você imaginou. Fuzzing garante que não quebra para as que você não imaginou."_

```bash
go test -fuzz=FuzzParseReader -fuzztime=10s ./internal/parser/
```

> _"O fuzzer começa com o corpus que fornecemos e muta as entradas — bytes aleatórios, unicode, tamanhos extremos. Se o parser entrar em pânico, ele encontrou um bug."_

```bash
ls testdata/fuzz/FuzzParseReader/ 2>/dev/null && echo "corpus salvo" || echo "nenhuma falha encontrada — corpus ainda vazio"
```

> _"Cada entrada interessante descoberta é salva no corpus e vira parte do `go test` normal — bugs encontrados pelo fuzzer nunca regridem. `go help testflag` documenta `-fuzz`, `-fuzztime`, `-fuzzminimizetime`."_

---

### Bloco 21 — `go test -shuffle` + `go test -count`

> _"Dois flags menos conhecidos que revelam problemas ocultos nos testes."_

```bash
go test -shuffle=on -v ./internal/parser/
```

> _"Ordem dos testes aleatorizada. Se um teste só passa quando roda depois de outro — dependência de estado — `-shuffle=on` vai expor isso. `go help testflag` documenta o flag."_

```bash
go test -count=3 ./internal/...
```

> _"Roda a suíte 3 vezes seguidas. Útil para testes flaky. `-count=1` é a forma idiomática de desabilitar o cache sem limpar o `testcache`."_

---

### Bloco 22 — `go tool pprof`

> _"Geramos dados. Agora: onde o logscope passa o tempo?"_

```bash
logscope -gen -lines 200000 -output access.log
logscope -input access.log -cpuprofile cpu.prof
```

> _"O perfil está gravado. Abrindo o visualizador:"_

```bash
go tool pprof -http=:8080 cpu.prof
```

> _"O `go tool pprof` abre um servidor local com visualizações interativas: flame graph, top funções, grafo de chamadas. Já vem no toolchain — sem instalar nada."_

> _"Para memória:"_

```bash
logscope -input access.log -memprofile mem.prof
go tool pprof -http=:8081 mem.prof
```

> _"Heap profile capturado após o GC — mostra onde o logscope aloca memória."_

---

### Bloco 23 — `go tool trace`

> _"O pprof mostra onde o tempo vai. O trace mostra **como** o tempo está distribuído entre goroutines."_

```bash
logscope -input access.log -trace trace.out
go tool trace trace.out
```

> _"Você vê goroutines nascendo e morrendo, o scheduler do Go distribuindo trabalho entre os cores, o GC pausando tudo. Nenhuma lib externa — está no toolchain."_

---

### Bloco 24 — `go tool nm`

> _"Último nível do iceberg. O que está dentro do binário compilado?"_

```bash
go tool nm ./logscope | head -20
go tool nm ./logscope | grep -i main
go tool nm ./logscope | wc -l
```

> _"`go tool nm` lista todos os símbolos: funções, variáveis, tipos — com endereço e tamanho. É o que o linker enxerga."_

> _"Útil para verificar se uma função foi incluída no binário, debugar tree-shaking, entender o que está ocupando espaço."_

```bash
go tool nm ./logscope | sort -k2 -rn | head -10
```

> _"Os maiores símbolos por tamanho. `go help tool` documenta todos os sub-tools disponíveis."_

---

## Resumo

| Comando | O que entrega |
|---------|--------------|
| `go help` | A porta de entrada para tudo no toolchain |
| `go version -m` | Audita dependências de qualquer binário Go |
| `go get` + `go mod tidy` | Gerencia dependências com segurança criptográfica |
| `go doc` | Documentação offline — stdlib e seu próprio código |
| `go fmt` | Formatação canônica — zero debate de estilo |
| `go vet` | Bugs que o compilador não vê — inspeciona a AST |
| `go env -w` | Configuração Go persistente sem variáveis de shell |
| `go list -m -u all` | Radar de dependências desatualizadas |
| `go mod why` | Rastreia a origem de qualquer dep no grafo |
| `go test -cover` | Cobertura — sabe o que está sendo testado |
| `go generate` | Automação de geração de código |
| `-ldflags` + `-tags` | Build customizado sem magic |
| `-trimpath` | Builds reproduzíveis, sem paths locais |
| `go test -race` | Detecta corridas de dados em tempo de execução |
| `go test -fuzz` | Encontra bugs com entradas aleatórias |
| `go test -shuffle` | Detecta dependências ocultas entre testes |
| `go tool pprof` | Profiling de CPU e memória |
| `go tool trace` | Visualiza goroutines e scheduler |
| `go tool nm` | Inspeciona símbolos compilados no binário |
| `go clean` | Gestão cirúrgica do cache de build e módulos |

> **Esse é o iceberg. A maioria conhece a ponta. Você agora conhece o fundo.**
```

- [ ] **Step 2: Verificar que o README tem mais de 300 linhas**

```bash
cd logscope
wc -l README.md
```

Saída esperada: número > 300.

- [ ] **Step 3: Commit**

```bash
cd ..
git add logscope/README.md
git commit -m "docs(logscope): adiciona roteiro completo da apresentação (24 blocos, 4 camadas)

Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>"
```

---

## Self-Review

### Spec coverage
- ✅ README.md com 24 blocos e 4 camadas (🌊 🔎 ⚙️ 🔬)
- ✅ `reset.sh` com bugs de fmt (processor.go) e vet (vet_demo.go) + restaura go.mod + imprime sequência
- ✅ Fuzz test para demo de `go test -fuzz`
- ✅ `vet_demo.go` com bugs reais detectáveis pelo `go vet`
- ✅ `go help` referenciado em cada bloco
- ✅ Fluxo de `go get` (Bloco 2) → `go list/mod why` (Blocos 11-12) → `go mod tidy` (Bloco 14)
- ✅ Iceberg: surface → quality → build avançado → testing profundo + observabilidade
- ✅ `top/` e `layer2/` não são removidos

### Type consistency
- Função `Avg()` em `processor.go` — mesmo nome usado no Python do reset.sh ✅
- `FuzzParseReader` — mesmo nome no fuzz test e na sequência do reset.sh ✅
- Paths de arquivos consistentes em todos os tasks ✅

### Placeholder scan
- Nenhum TBD, TODO ou "similar ao Task N" encontrado ✅
- Todos os steps têm código completo ✅
