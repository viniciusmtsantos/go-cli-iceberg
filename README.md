# O Iceberg da CLI do Go

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
> ANALISA LOGS → Cria relatório de latência, status e endpoints a partir do access.log

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

> Explore: `go help testflag`

---

### Bloco 4 — `go run`
> Rodando um pacote externo diretamente com `go run`

```bash
# Rodar o analisador de vulnerabilidades em todas as dependências do projeto em memoria
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

---

### Bloco 5 — `go build` + `go install`

```bash
# Compila o projeto para um binário local (sem instalar globalmente)
go build -o hexa_mundial ./cmd/logscope
./hexa_mundial -input testdata/access.log
```

```bash
# Compila e instala globalmente (disponível em $GOPATH/bin)
go install ./cmd/logscope
```

---

## CAMADA 2 - Qualidade, Dependências e Documentação

### Bloco 7 — `go doc`

**📖 Stdlib Docs**
```bash
# Documentação offline para a função fmt.Printf
go doc fmt.Printf
```

**🔍 Seu Próprio Código**
```bash
# Documentação do pacote entry
go doc ./pkg/entry
```

```bash
# Documentação da struct LogEntry
go doc ./pkg/entry LogEntry
```

---

### Bloco 8 — `go fmt`

**O atalho e o motor**
- `go fmt` (atalho) → `gofmt -l -w`
- `gofmt` (motor) → `gofmt [flags] [path]`

```bash
# Listar arquivos que precisam de formatação
gofmt -l -d .
```

```bash
# Aplica formatação no projeto inteiro
go fmt ./...
```

> Explore `gofmt -help`
---

### Bloco 9.0 — `go fix`

```bash
# Atualiza e corrige partes do código automaticamente
go fix ./... 
```

**Quando Usar**
- Remover APIs depreciadas
- Refatorar partes do código para padrões conhecidos

**Atenção**
- Não vai resolver todos os seus problemas, mas é uma ferramenta útil

---

### Bloco 10 — `go env`

**Configuração Global**
```bash
# Configura valor pra GOFLAGS (trimpath remove caminhos do sistema dos binários durante build)
go env -w GOFLAGS=-trimpath
```

```bash
go env GOFLAGS
```

```bash
go env -u GOFLAGS
```

> Explore: `go help environment`

---

### Bloco 11 — `go list`

```bash
# Listar os pacotes do projeto
go list ./...
```

```bash
# Listar dependências do projeto indicando versão mais nova disponível para elas
go list -m -u all
```

---

### Bloco 12 — `go test -cover`

**Cobertura de Testes**
- Testes passam, mas quanto do código está exercitado?

```bash
# Cobertura geral do projeto
go test -short -cover ./internal/...
```

```bash
# Gera arquivo de cobertura para análise com `go tool cover`
go test -short -coverprofile=coverage.out ./internal/...
```

```bash
# Cobertura por função
go tool cover -func=coverage.out
```

```bash
# Abre um navegador com visualização de cobertura (linhas vermelhas = não cobertas, verdes = cobertas)
go tool cover -html=coverage.out
```

> Explore: `go help testflag`

---

### Bloco 13 — Limpeza `go clean`

```bash
# Limpa o cache de testes (força reexecução dos testes na próxima vez)
go clean -testcache
```

**LIMPA TUDO**
```bash
# Limpa o cache de build (compilação do zero na próxima vez)
go clean -cache
```

---

## Camada 3 — Geração, Build avançado e Foco

### Bloco 14 — `go generate`

**Diretiva**
- `//go:generate` → comando a ser rodado quando `go generate` for executado

```bash
# Fácil manutenção: novo level (TRACE)? Roda `go generate` Crie um novo tipo e execute
go generate ./pkg/entry/
```

---

### Bloco 15 — `go build -tags` e `go build -ldflags`

**Build Tags (Compilação condicional)**
- `//go:build json` → incluir apenas quando flag ativa

```bash
go build -o logscope ./cmd/logscope
./logscope -input testdata/access.log  
```

```bash
go build -tags json -o logscope-json ./cmd/logscope
./logscope-json -input testdata/access.log 
```

**Injetar Metadados (ldflags)**
- Injetar metadados em tempo de build (versão, commit, timestamp)

```bash
go build -ldflags "-X main.version=1.2.0 -X main.commit=7aa0090" -o logscope-linker ./cmd/logscope
```

```bash
go version -m ./logscope-linker
```

---

### Bloco 16 — `GOOS` e `go build` para cross compilation

```bash
# Compila para Windows
GOOS=windows go build -o logscope.exe ./cmd/logscope
```

```bash
# Verificar o binário gerado
file logscope.exe
```

---

## CAMADA 4 — Performance, observabilidade e... bug?

### Bloco 18 — Benchmarks

```bash
# Rodar todos os benchmarks do parser (latência, throughput, memória)
go test -bench=. -benchmem ./internal/parser/
```

**Interpretando Resultados** (ex: `BenchmarkParseReader-12`)
1. **Iterações** (6032278): Go rodou 6+ milhões de vezes em 1 segundo
2. **Latência** (`X ns/op`): X nanossegundos por linha processada
3. **Memory** (`Y B/op, Z allocs/op`): Y bytes e Z alocações por operação
4. **Paralelismo** (sufixo `-12`): usou todos os 12 núcleos disponíveis

> Explore: `go help testflag`

---

### Bloco 18.1 — Erro Comum: Não pré-alocar memória

```bash
# latencies: make([]time.Duration, 0, len(chunk)) para latencies: make([]time.Duration, 0)
go test -run=^$ -bench=BenchmarkProcessSafe -benchmem ./internal/processor/
```

- Sem saber tamanho final, Go para o tempo TODO para pedir mais memória

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