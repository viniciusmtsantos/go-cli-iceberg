# O Iceberg da CLI do Go

> **Antes de começar:** rode `./reset.sh` no terminal (não mostrar para o público).

## Abertura — `go help`

- `go help` → lista TUDO: comandos + tópicos conceituais do toolchain
- `go help <command>` → documentação completa, offline, no terminal
- `go help <topic>` → conceitos: `modules`, `buildconstraint`, `testflag`

- Mais que apresentação de comandos (cansativa) → convite a explorar
- `go help` é seu mapa. Cada bloco terminará sugerindo: *"explore com `go help <x>`"*
- Vocês vão sair daqui motivados a conhecer sozinhos o resto

---

## CAMADA 1 — Let's Go

### Bloco 0 — `go version`

```bash
# Versão do Go instalada
go version
```

```bash
# Extrai metadados embutidos no executável analisado pra mostrar informações de compilação
go version -m $(which docker)
```

---

### Bloco 1 — Logscope: Analisador de Logs HTTP
- Lê logs HTTP → Analisa e cria relatório de latência, status e endpoints a partir do access.log

```bash
# Rodando o projeto indicando o pacote main
go run ./cmd/logscope -input testdata/access.log
```

### Bloco 2 — `go mod init` || `go get` || `go mod tidy`

```bash
# Inicializa um novo módulo Go
go mod init github.com/gopher/logscope
```
```bash
# Adiciona uma dependência (ex: color para colorir o output)
go get github.com/fatih/color@latest
```
```bash
# Faz a faxina no nosso módulo
go mod tidy
```

---

### Bloco 3 — `go test`

```bash
# Rodar todos os testes do projeto com verbose
go test -v ./...
```


```bash
# Rodar testes rápidos (ignora testes longos) com verbose
go test -v -short ./...
```

> Dica de flags: `go help testflag`

---

### Bloco 4 — `go run`

Usando `govulncheck` para varrer dependências procurando CVEs públicas

```bash
# Rodar o analisador de vulnerabilidades em todas as dependências do projeto em memoria
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

---

### Bloco 5 — `go build` + `go install`

```bash
go build -o hexa_mundial ./cmd/logscope
./hexa_mundial -input testdata/access.log
```

```bash
# Compila e instala globalmente (disponível em $GOPATH/bin)
go install -n ./cmd/logscope
```

---

## CAMADA 2 - Qualidade, Dependências e Documentação

### Bloco 7 — `go doc`

**🎯 Documentação Local — Offline**
- Como descobrir o que `ParseReader` faz sem abrir o navegador?
- `go doc fmt.Printf` → parâmetros, tipos, comportamento, no terminal

**📖 Stdlib Docs**
```bash
go doc fmt.Printf
```

**🔍 Seu Próprio Código**
```bash
go doc ./pkg/entry
go doc ./pkg/entry LogEntry
```

> Próximo: `go help doc`

---

### Bloco 8 — `go fmt`

**🎯 Formatação Canônica — Zero Debate**
- `go fmt` (atalho) → formata o projeto (~rápido, sem flags)
- `gofmt` (motor) → tool raiz (flags: `-l` listar, `-d` diff)

**✅ Aplicar Formatação**
```bash
go fmt ./...
```

```bash
gofmt -l -d .
```

> Próximo: `go help fmt`

---

### Bloco 9.0 — `go fix`

**🔧 Quando Usar**
- Remover deprecated APIs
- Refatoração de padrão conhecido

**💡 Exemplo Real**
```bash
go fix ./... 
```

- Nem sempre vai ser algo que vai resolver todos os problemas do nosso código, mas pode ser uma ferramenta poderosa

> Próximo: `go help fix`

---

### Bloco 10 — `go env`

**🔧 Superpoder: Configuração Global**
```bash
go env -w GOFLAGS=-trimpath
go env GOFLAGS
go env -u GOFLAGS  # desfazer
```

> Próximo: `go help environment`

---

### Bloco 11 — `go list`

**🎯 Listar Pacotes e Módulos**
- Você sabe exatamente o que está carregando?
- Dependências atualizadas?

**📋 Listar Pacotes do Projeto**
```bash
go list ./...
```

**🔍 Auditar Dependências**
```bash
go list -m all
go list -m -u all  # -u: versão mais nova disponível
```

Resultado: `[v...]` ao lado de deps = alerta de desatualização

> Próximo: `go help list`

---

### Bloco 12 — `go test -cover`

**🎯 Cobertura de Testes — Saber o que está sendo testado**
- Testes passam, mas quanto do código está exercitado?
- Percentual + análise por função + HTML visual

**📊 Coverage Simples**
```bash
go test -short -cover ./internal/...
```

**🔍 Por Função**
```bash
go test -short -coverprofile=coverage.out ./internal/...
```

**🎨 Visual HTML**
```bash
go tool cover -func=coverage.out
go tool cover -html=coverage.out
```

Resultado: Navegador abre com linhas vermelhas (não cobertas) e verdes (cobertas)

> Próximo: `go help testflag`

---

### Bloco 13 — Faxina `go clean`

**🧹 Remover Deps Não Usadas**
```bash
go mod tidy
cat go.mod
```

Resultado: `fatih/color` removido (nunca foi importado de verdade)

**💾 Cache de Build**
```bash
go test -short ./internal/...
go clean -testcache
go test -short ./internal/...
```

**🔥 Nuclear Options**
```bash
go clean -cache        # objetos compilados (~recompila tudo do zero)
```

> Próximo: `go help clean`

---

## ⚙️ Camada 3 — Build avançado e geração

### Bloco 14 — `go generate`

**🎯 Automação de Geração de Código**
- Tipos com constantes (Level: DEBUG, INFO, WARN, ERROR, FATAL)
- Método `String()` gerado automaticamente
- `stringer` → converte números em texto (eficiente)

**📝 Diretiva Mágica**
- `//go:generate` → comentário especial (não executa automaticamente)
- `go generate` procura por diretivas e executa

**🔧 Demonstração**
```bash
cat pkg/entry/entry.go      # veja a diretiva //go:generate
cat pkg/entry/level_string.go
```

**♻️ Regenerar**
```bash
rm pkg/entry/level_string.go
go generate ./pkg/entry/
cat pkg/entry/level_string.go
```

**💡 Por que Importa**
- Elimina código braçal (enum conversion)
- Criação automatizada com `stringer`
- Performance: mapeamento de memória, muito mais rápido que validação manual
- Fácil manutenção: novo level (TRACE)? Roda `go generate` novamente

> Próximo: `go help generate`

---

### Bloco 15 — Build Tags + `-ldflags`

**🎯 Customizar Build Sem Magic Strings**
- Dois modos de output: texto e JSON (qual incluir?)
- Injetar metadata em build time (versão, commit, timestamp)

**🏷️ Build Tags (Conditional Compilation)**
```bash
cat internal/report/report.go        # //go:build !json
cat internal/report/report_json.go   # //go:build json
```

- `//go:build json` → incluir apenas quando flag ativa
- Compilador inclui/exclui arquivos (não if/else em tempo de execução)

**🔨 Build com e sem Tag**
```bash
go build -o logscope ./cmd/logscope
./logscope -input testdata/access.log   # texto

go build -tags json -o logscope-json ./cmd/logscope
./logscope-json -input testdata/access.log  # JSON
```

**💉 Injetar Metadata (ldflags)**
```bash
go build -o logscope ./cmd/logscope
./logscope -version
go version -m ./logscope
```

- Injeção no linker (imutável)

**🔒 Versão no Binário**
```bash
go build \
  -ldflags "-X main.version=1.2.0 -X main.commit=$(git rev-parse --short HEAD)" \
  -o logscope-linker \
  ./cmd/logscope

./logscope-linker -version
go version -m ./logscope-linker
```

**💡 Vantagens**
- Versão gravada no binário (imutável, auditável)

> Próximo: `go help build`

---

### Bloco 16 — Cross-compilation

**🎯 Compilar para Qualquer Plataforma**
- Go compila para qualquer OS/Arch sem Docker, sem VM

**🌍 Compilar Cruzado**
```bash
GOOS=windows go build -o logscope.exe ./cmd/logscope
ls -lh logscope.exe
file logscope.exe

GOOS=linux GOARCH=arm64 go build -o logscope-arm64 ./cmd/logscope
ls -lh logscope-arm64
file logscope-arm64
```

> Próximo: `go help build`

---

## 🔬 Camada 4 — Testing profundo e observabilidade

### Bloco 18 — Benchmarks

**🎯 Medir Performance Real**
- Testes garantem que código funciona
- Benchmarks provam que **aguenta o tranco em produção**
- Resultado do Go: 3 visões simultâneas (escala, latência, custo de memória)

**⚡ Demonstração**
```bash
go test -bench=. -benchmem ./internal/parser/
```

**📊 Interpretando Resultado** (ex: `BenchmarkParseReader-12`)
1. **Iterações** (6032278): Go rodou 6+ milhões de vezes em 1 segundo (não é lucky guess — prova estatística)
2. **Latência** (`193 ns/op`): 193 nanossegundos por linha processada (~1 milionésimo de segundo)
3. **Memory** (`112 B/op, 1 allocs/op`): 112 bytes + 1 alocação = quase sem lixo pro GC
4. **Paralelismo** (sufixo `-12`): usou todos os 12 núcleos disponíveis

**🎯 Por que o Hot Path?**
- Fatiar strings + converter datas = **hot path** (caminho quente)
- Gargalo se lento/memory wasteful

**⚙️ Comparando Estratégias**
```bash
go test -run=^$ -bench=BenchmarkProcessSafe -benchmem ./internal/processor/
```

Resultado: 1 worker vs 8 workers
- **Vitória (ns/op cai)**: paralelismo funciona
- **Lei dos rendimentos decrescentes**: 8x workers ≠ 8x speedup (overhead de Goroutines)
- **Troca inteligente (allocs/op)**: mais alocações = mais segurança (race-free com isolamento)

> Próximo: `go help testflag`

---

### Bloco 18.1 — Live Coding: Otimização Sênior vs Júnior

**🎯 Demonstração Prática: De Lento para Rápido**

**⚠️ Erro Comum 1: Não pré-alocar memória**
```bash
# Altere: latencies: make([]time.Duration, 0, len(chunk))
# Para:   latencies: make([]time.Duration, 0)
go test -run=^$ -bench=BenchmarkProcessSafe -benchmem ./internal/processor/
```

Resultado: `allocs/op` explode de 116 para milhares
- Sem saber tamanho final, Go para o tempo TODO para pedir mais memória
- Salva o GC de um infarto: reverta!

**⚡ Otimização Real: Remover Cálculo do Hot Loop**
```bash
# 4 mudanças cirúrgicas:
# 1. struct partialStats: map[string]int → map[entry.Level]int
# 2. Criação: byLevel: make(map[entry.Level]int, 5)
# 3. Hot Loop: p.byLevel[e.Level.String()]++ → p.byLevel[e.Level]++
# 4. Merge: stats.ByLevel[k] += v → stats.ByLevel[k.String()] += v

go test -run=^$ -bench=BenchmarkProcessSafe -benchmem ./internal/processor/
```

Resultado: `ns/op` cai ~10%
- Tirou conversão String do loop (centenas de milhares de iterações)
- Jogou para merge final (apenas 5 vezes)
- CPU clean, allocs unchanged

**💡 Ganho Real**
- 1 worker: 6.6M → 5.7M ns (~900K nanosegundos economizados)
- Produção: bilhões de logs/dia = menos servidores = $ economizado

> Próximo: `go help testflag`

---

### Bloco 19 — `go test -race`

**🎯 Detectar Race Conditions em Tempo de Execução**
- Temos dois processors: `ProcessSafe` (correto) e `ProcessNaive` (com bug)
- Race detector instrumenta cada acesso à memória

**🔍 Sem Proteção (Crash Genérico)**
```bash
go test -run TestNaiveRace ./internal/processor/
```

Resultado: `fatal error: concurrent map writes` — morre feio, sem saber onde

**🚨 Com Race Detector**
```bash
go test -race -run TestNaiveRace ./internal/processor/
```

Resultado: `WARNING: DATA RACE`
- **Linha exata**: Goroutine 13 leu X, Goroutine 15 escreveu X
- **Stack trace**: Linha 45 (`naive.go`), ponto de origem (linha 39)
- Não é mágica — compilador instrumenta cada acesso

**✅ Validar Código Correto**
```bash
go test -race -run TestProcessSafe_DeterministicWithRace ./internal/processor/
```

Resultado: Silêncio (clean, race-free)

**💡 Em Produção**
- Race detector custa 5-10x mais CPU (não use em prod)
- Use em CI/CD (bloqueia build se race detectada)

> Próximo: `go help testflag`

---

### Bloco 20 — `go test -fuzz`

**🎯 Fuzzing — Encontrar Bugs com Entradas Aleatórias**
- Cobertura testa entradas que **VOCÊ imaginou**
- Fuzzing testa entradas que **VOCÊ NÃO imaginou**
- Produção é criativa — encontra coisas malucas

**📖 Estrutura**
```bash
cat internal/parser/parser_fuzz_test.go
```

- `f.Add()` → Seed Corpus (exemplos válidos + inválidos)
- `f.Fuzz()` → Go mutação infinita (nulos, emojis, strings cortadas)
- **Regra**: ParseReader não pode Panic (erro é OK, crash é crime)

**🚀 Soltar o Monstro**
```bash
go test -fuzz=FuzzParseReader -fuzztime=10s ./internal/parser/
```

**📊 Interpretando Output**
- `gathering baseline` → lê sementes, aprende padrões
- `execs/sec` → **100.000+ cenários em 10 segundos** (QA humana não consegue)
- `new interesting` → mutação descobriu caminho novo (salva para testar sempre)
- `FAIL - context deadline exceeded` → vitória! (relógio estourou, sem crash encontrado)

**⚠️ Dica de Trincheira**
```bash
go test -fuzz=FuzzParseReader -fuzztime=5m ./internal/parser/  # CI/CD: 5 min
```

- Sem `-fuzztime` = roda infinito (GitHub Actions explode sua conta!)
- Cada entrada descoberta fica no corpus (nunca regride)

> Próximo: `go help testflag`

---

### Bloco 21 — `go test -shuffle` + `go test -count`

**🔄 Rodar Múltiplas Vezes**
```bash
go test -short -count=3 ./internal/...
```

Resultado: Suite inteira 3 vezes
- Testa flaky tests (tempo de CPU, scheduling)
- `-count=1` é forma idiomática de disable cache

> Próximo: `go help testflag`

---

### Bloco 22 — `go tool` e o Poderoso `pprof`

**🎯 Profiling — Raio-X do Programa em Execução**
- `go tool` → canivete suíço de ferramentas baixo nível
- `pprof` → CPU/Memória profiler oficial do Go

**💾 Preparação (Dados Pesados)**
```bash
./logscope -input testdata/access.log -cpuprofile cpu.prof
go tool pprof -no_browser -http=:8080 cpu.prof
```

**(No navegador) → Menu "View" → "Flame Graph"**

**🎨 Flame Graph**
- Largura do retângulo = tempo de CPU consumido
- Altura = profundidade de stack
- Bateu olho, achou gargalo em 5 segundos

**🧠 Memória**
```bash
./logscope -input testdata/access.log -memprofile mem.prof
go tool pprof -no_browser -http=:8081 mem.prof
```

Resultado: Heap Profile (foto da RAM post-GC)
- Vazamento de memória? Função está esquecendo liberar dados? **Essa tela aponta**

> Próximo: `go help tool`

--

### Bloco 24 — `go bug` (NOVO - Escape Hatch)

**🎯 Quando Tudo Falha — Report de Bugs**
- Encontrou um bug que não consegue reproduzir?
- Comportamento estranho que não faz sentido?
- `go bug` é sua saída de emergência

**🚨 Quando Usar**
- Suspeita de bug no próprio Go (raro, mas acontece)
- Comportamento reproduzível que parece impossível
- Precisa reportar para time do Go

**📋 Demonstração**
```bash
go bug
```

**🎓 Closure: O Iceberg Completo**
- Começamos em `go help` (porta de entrada)
- Percorremos superfície → dev rápido → compilação → distribuição
- Camada 2: qualidade, dependências, documentação
- Camada 3: builds customizados, cross-compile, reprodutibilidade
- Camada 4: performance, concorrência, observabilidade profunda
- E quando tudo falha: `go bug`

## Resumo

| Comando | O que entrega |
|---------|--------------|
| `go help` | A porta de entrada para tudo no toolchain |
| `go version -m` | Audita dependências de qualquer binário Go |
| `go get` + `go mod tidy` | Gerencia dependências com segurança criptográfica |
| `go run` | Executa sem compilar (dev rápido, sem artefatos) |
| `go build` | Compila para produção (nativo, sem runtime) |
| `go install` | Compila + instala globalmente em `$GOPATH/bin` |
| `go doc` | Documentação offline — stdlib e seu próprio código |
| `go fmt` | Formatação canônica — zero debate de estilo |
| `go vet` | Bugs que o compilador não vê — inspeciona a AST |
| `go fix` | Atualiza código automaticamente (breaking changes do Go) |
| `go env -w` | Configuração Go persistente sem variáveis de shell |
| `go list -m -u all` | Radar de dependências desatualizadas |
| `go mod why` | Rastreia a origem de qualquer dep no grafo |
| `go test -cover` | Cobertura — sabe o que está sendo testado |
| `go clean` | Gestão cirúrgica do cache de build e módulos |
| `go generate` | Automação de geração de código |
| `go build -ldflags` + `-tags` | Build customizado sem magic |
| `go test -race` | Detecta corridas de dados em tempo de execução |
| `go test -fuzz` | Encontra bugs com entradas aleatórias |
| `go test -shuffle` | Detecta dependências ocultas entre testes |
| `go test -bench` | Mede performance real (latência, throughput, memória) |
| `go tool pprof` | Profiling de CPU e memória |
| `go bug` | Reporta bugs para a comunidade Go |