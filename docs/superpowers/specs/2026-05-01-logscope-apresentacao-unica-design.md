# Design: logscope como projeto único da apresentação

**Data:** 2026-05-01
**Repositório:** viniciusmtsantos/go-cli-iceberg
**Status:** aprovado

---

## Problema e objetivo

O repositório atualmente tem três projetos separados (`top/`, `layer2/`, `logscope/`) para demonstrar diferentes camadas do toolchain Go. O objetivo é **consolidar tudo no `logscope`**, tornando-o o único projeto usado na apresentação da Casa de Backend.

O `logscope` é um analisador de logs HTTP com geração de dados, processamento concorrente, build tags, ldflags, profiling e tracing — estrutura rica o suficiente para demonstrar todos os comandos do toolchain Go em profundidade.

---

## Abordagem

**Abordagem A — README único + `reset.sh` com patches Python.**

Um único `README.md` na raiz do `logscope/` serve como roteiro completo da apresentação. O `reset.sh` usa Python inline (igual ao `layer2/reset.sh`) para introduzir bugs intencionais no código antes de apresentar, e imprime a sequência de comandos. Os bugs ficam em arquivos limpos no git e são introduzidos apenas na hora da demo.

As pastas `top/` e `layer2/` continuam no repositório como referência histórica.

---

## Estrutura de arquivos

### Arquivos criados

```
logscope/
├── README.md                              ← roteiro completo (novo)
├── reset.sh                               ← restaura estado inicial (novo)
└── internal/parser/
    └── parser_fuzz_test.go                ← fuzz test para demo go test -fuzz (novo)
```

### Arquivos existentes (sem alterações no git)

Os bugs são introduzidos apenas em tempo de apresentação pelo `reset.sh`:

- `internal/processor/processor.go` — recebe bug de `go fmt` (indentação quebrada em `ProcessSafe`)
- `internal/parser/parser.go` — recebe bug de `go vet` (verb `%d` recebendo string)

---

## Estrutura do README (blocos do iceberg)

### 🌊 Abertura e Superfície

| Bloco | Comando(s) | O que demonstra |
|-------|-----------|-----------------|
| Abertura | `go help` | Meta-comando — não precisa de terminal |
| 0 | — | Apresentação do `logscope` e sua estrutura |
| 1 | `go version`, `go version -m ./logscope` | Versão do runtime + auditoria de binário |
| 2 | `go.mod`, `go get`, `go mod tidy` | Módulos, dependências, lockfile |
| 3 | `go test ./...`, `go test -v` | Testes nativos |
| 4 | `go run ./cmd/logscope`, `go run <tool>@latest` | Execução sem build, ferramenta externa isolada |
| 5 | `go build -o logscope .` | Compilação, binário nativo |
| 6 | `go install .`, `go env GOPATH` | Instalação global |

### 🔎 Camada 2 — Qualidade e dependências

| Bloco | Comando(s) | O que demonstra |
|-------|-----------|-----------------|
| 7 | `go doc`, `go doc ./internal/parser ParseReader` | Documentação offline e do próprio código |
| 8 | `gofmt -l .`, `gofmt -d`, `go fmt ./...` | Formatação canônica com bug intencional |
| 9 | `go vet ./...` | Bugs que o compilador não vê |
| 10 | `go env`, `go env -w`, `go env -u` | Configuração persistente do Go |
| 11 | `go list ./...`, `go list -m -u all` | Radar de dependências desatualizadas |
| 12 | `go mod why` | Rastreia origem de dependências |
| 13 | `go test -cover`, `go tool cover -func`, `go tool cover -html` | Cobertura de testes |
| 14 | `go clean -testcache`, `go clean -cache`, `go clean -modcache` | Gestão cirúrgica do cache |

### ⚙️ Camada 3 — Geração e build avançado (novos blocos)

| Bloco | Comando(s) | O que demonstra |
|-------|-----------|-----------------|
| 15 | `go generate ./pkg/entry/` | Geração de código com stringer |
| 16 | `go build -tags json`, `go build -ldflags "-X main.version=..."` | Build tags + injeção de variáveis em tempo de build |
| 17 | `GOOS=windows go build -o logscope.exe`, `go build -trimpath` | Cross-compilation + builds reproduzíveis |

### 🧪 Camada 4 — Testing profundo (novos blocos)

| Bloco | Comando(s) | O que demonstra |
|-------|-----------|-----------------|
| 18 | `go test -bench=.`, `go test -benchmem` | Benchmarks com latência e pressão no GC |
| 19 | `go test -race ./...` com `-naive` | Detector de corridas de dados |
| 20 | `go test -fuzz=FuzzParseReader` | Fuzzing com entradas aleatórias |
| 21 | `go test -shuffle=on`, `go test -count=3` | Ordem aleatória de testes + bypass de cache |

### 📊 Camada 5 — Observabilidade

| Bloco | Comando(s) | O que demonstra |
|-------|-----------|-----------------|
| 22 | `logscope -cpuprofile cpu.prof`, `go tool pprof -http=:8080 cpu.prof` | Profiling de CPU |
| 23 | `logscope -trace trace.out`, `go tool trace trace.out` | Tracing de execução |
| 24 | `go tool nm ./logscope \| head -20` | Inspeção de símbolos do binário |

---

## Tabela de resumo final (para o README)

| Comando | O que entrega |
|---------|--------------|
| `go help` | A porta de entrada para tudo no toolchain |
| `go version -m` | Audita dependências de qualquer binário Go |
| `go get` + `go mod tidy` | Gerencia dependências com segurança criptográfica |
| `go doc` | Documentação offline — stdlib e seu próprio código |
| `go fmt` | Formatação canônica — zero debate de estilo |
| `go vet` | Bugs que o compilador não vê |
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

---

## Bugs intencionais (introduzidos pelo `reset.sh`)

### Bug 1 — `go fmt` em `internal/processor/processor.go`

Quebrar a indentação da função `ProcessSafe`. O código compila e os testes passam, mas `gofmt -l .` reporta o arquivo e `gofmt -d processor.go` mostra o diff.

### Bug 2 — `go vet` em `internal/parser/parser.go`

Inverter o verb de formato num `fmt.Printf` de log de erro: usar `%d` para uma string e `%s` para um error. O código compila, mas `go vet ./...` detecta o tipo incompatível.

---

## `reset.sh` — comportamento

1. Remove artefatos: `logscope`, `logscope.exe`, `logscope-json`, `access.log`, `cpu.prof`, `mem.prof`, `trace.out`, `coverage.out`
2. Executa `go clean -testcache`
3. Reverte `go.mod` ao estado sem dependências externas (caso `go get` tenha sido usado durante demo)
4. Introduz bug de `go fmt` em `processor.go` via Python inline
5. Introduz bug de `go vet` em `parser.go` via Python inline
6. Imprime checklist da sequência de apresentação

---

## Fuzz test — `internal/parser/parser_fuzz_test.go`

```go
func FuzzParseReader(f *testing.F) {
    // corpus seeds de linhas válidas
    f.Add("2026-04-26T10:00:00Z INFO GET /api/users 200 145 1234\n")
    f.Fuzz(func(t *testing.T, data string) {
        // o parser não deve entrar em pânico com nenhuma entrada
        _, _ = parser.ParseReader(strings.NewReader(data))
    })
}
```

O fuzz test demonstra `go test -fuzz=FuzzParseReader -fuzztime=10s` ao vivo, mostrando o Go encontrando entradas inesperadas.

---

## Considerações

- O `logscope` **não tem dependências externas** (só stdlib). Para o demo de `go get` + `go mod tidy` (Bloco 2), o apresentador adiciona `github.com/fatih/color@latest` ao vivo e depois remove. O `reset.sh` garante que `go.mod` volta ao estado limpo.
- O tema central é **`go help`** — cada bloco menciona que a flag demonstrada está documentada em `go help <comando>`.
- Os blocos podem ser apresentados em sessões separadas ou na ordem do iceberg dependendo do tempo disponível.
