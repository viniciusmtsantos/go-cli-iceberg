# O Iceberg da CLI do Go

## Abertura —  `go help`

- `go help` → Lista TUDO go pode oferecer: comandos + tópicos conceituais
- `go help <command>` → Execução
- `go help <topic>` → Conceitos

---

### Projeto Logscope: Nosso Playground
Recebe uma amostra de logs, processa e cria relatório sobre top endpoints, status de request e percentis de latência.

## CAMADA 1 - Let's Go

### `go version`

_Mostrar versão do Binário Go instalada na sua máquina_
```bash
go version
```

`version -m`  _Informar metadados embutidos no binário durante a compilação_
```bash
go version -m /usr/bin/docker
```

---

### `go mod`   `go list`   `go get`

_Inicializar novo módulo Go_
```bash
go mod init github.com/gopher/logscope
```

_Organizar o módulo. Veja o go.mod e go.sum antes e depois_
```bash
go mod tidy
```

_Listar dependências do projeto indicando última versão disponível para elas_
```bash
go list -m -u all
```

_Atualizar todas as dependências para a última versão disponível. Atenção às breaking changes_
```bash
go get -u ./...
```

**Explore  `go help modules`**

---

### `go test`

_Executar todos os testes do projeto com verbose_
```bash
go test -v ./internal/parser
```

**Explore  `go help testflag`**

---

### `go run`

**Executar meu projeto sem compilar um binário local**
_Executar o projeto indicando o pacote main_
```bash
go run ./cmd/logscope -input testdata/access.log
```

**Executar um pacote Go diretamente sem instalar**
_`govulncheck`  analisa vulnerabilidades nas dependências do projeto_
```bash
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

**Explore  `go help packages`**

---

### `go build`   `go install`

_Compilar o projeto para um binário local (sem instalar globalmente)_ 
```bash
go build ./cmd/logscope
```

_Executar o binário gerado_
```bash
./logscope -input testdata/access.log
```

_Compilar e instalar globalmente (disponível em $GOPATH/bin)_
```bash
go install ./cmd/logscope
```

**Explore  `go help buildmode`**

---

## CAMADA 2 - Ambiente, Limpeza, Documentação e Qualidade

### `go env`
**Mapa de configuração global**

_Configura valor global pra GOFLAGS (trimpath remove caminhos do sistema dos binários durante build)_
```bash
go env -w GOFLAGS=-trimpath
```

_Verificar configuração atual de GOFLAGS_
```bash
go env GOFLAGS
```

_Resetar GOFLAGS para valor padrão_
```bash
go env -u GOFLAGS
```

**Explore  `go help environment`**

---

### `go clean`
**Limpador oficial**

_Limpa o cache de testes (força reexecução dos testes na próxima vez)_
```bash
go clean -testcache
```

_Limpa o cache de build (compilação do zero na próxima vez)_
```bash
go clean -cache
```

**Explore  `go help cache`**

---

### `go doc`
**Documentador oficial**

_Consultar documentação da **standard lib**_
```bash
go doc fmt.Printf
```

_Consultar documentação no **seu código**_
```bash
go doc ./pkg/entry LogEntry
```

---

### `go fmt`
**Seu formatador oficial do go**
_`go fmt` (atalho) →  `gofmt -l -w [path]`_
_`gofmt` (motor) →  `gofmt [flags] [path]`_

_Listar arquivos que precisam de formatação_
```bash
gofmt -l -d .
```

_Aplicar formatação no projeto inteiro_
```bash
go fmt ./...
```

**Explore  `gofmt -help`**

---

### `go fix`
**Refatorar dependências depreciadas e trechos do código para padrões conhecidos**

_Formatar e refatorar partes do código automaticamente_
```bash
go fix ./... 
```

**Atenção**
- Não vai resolver todos os problemas do seu código, mas é uma ferramenta útil

---

### `go test -cover`  e  `go tool cover`
**Quanto do código tá coberto?**

_Informar cobertura de testes no pacote e gerar output a partir destas informações_ 
```bash
go test -cover -coverprofile=coverage.out ./internal/parser/...
```

_Informar cobertura de testes por função a partir do  coverage.out_
```bash
go tool cover -func=coverage.out
```

_Criar um HTML com visualização de cobertura a partir do coverage.out_
```bash
go tool cover -html=coverage.out
```

**Explore  `go help testfunc`**

---

## Camada 3 — Geração automática, Build customizado e Performance 

### `go generate`
**Automatizar geração de código**

Diretiva  `//go:generate`  _→ Executar comando definido quando `go generate` for executado_
```bash
go generate ./pkg/entry/
```

---

### `go build -tags`
**Compilação condicional**

Diretiva  `//go:build <tag>`  _→ inclui o arquivo apenas quando a tag está ativa no build_
_Padrão_
```bash
./logscope -input testdata/access.log
```

_Arquivo incluído e suporte ao relatorio em JSON_
```bash
go build -tags json -o logscope-json ./cmd/logscope
```

```bash
./logscope-json -input testdata/access.log 
```

---

### `go build -ldflags`
**Injeção de metadados**
Flag  `ldflags`  _→ Injetar metadados em tempo de build (versão, commit, timestamp, etc.)_
```bash
go build -ldflags "-X main.version=1.2.0 -X main.commit=7aa0090" -o logscope-linker ./cmd/logscope
```

_Verificar metadados embutidos no binário_
```bash
go version -m ./logscope-linker
```

**Explore  `go help buildconstraint`**

---

### `go test -bench`

_Rodar todos os benchmarks do parser (latência, throughput, memória)_
```bash
go test -bench=. -benchmem ./internal/parser/
```

**Interpretando Resultados**
1. **Iterações** (`V`): Go rodou V milhões de vezes em 1 segundo
2. **Latência** (`X ns/op`): X nanossegundos por linha processada
3. **Memoria** (`Y B/op, Z allocs/op`): Y bytes e Z alocações por operação
4. **Paralelismo** (sufixo `-12`): usou todos os 12 núcleos disponíveis

---

### `go test -run`
_Executar um teste específico_
```bash
go test -run TestNaiveRace -v ./internal/processor/
```

Temos o erro, mas poucas informações para investigar

---

### `go test -race`

**Race detector**
Flag  `-race`  _→ Executar o teste instrumentando cada acesso à memória e mostrando quais goroutines conflitaram_ 
```bash
go test -run TestNaiveRace -race ./internal/processor/
```

---

### `go test -short`

**Mudar comportamento padrão de execução dos testes**
Flag  `-short`  _→ Executar o teste de forma resumida, skipando algum teste_ 
```bash
go test -run TestNaiveRace -v -short ./internal/processor/
```

---

## CAMADA 4 — Profiling, Bug e Observabilidade

### `go tool pprof`
Flag  `pprof`  _→ Profiler de CPU/Memória oficial do Go_

**Preparação e amostragem do profiling de CPU**
_Registra onde o processador passou o tempo enquanto sua aplicação rodava._
```bash
./logscope -input testdata/access.log -cpuprofile cpu.prof
go tool pprof -http=:8080 cpu.prof
```

**Preparação e amostragem do profiling de memória**
_Registra quem alocou memória, quanto e onde no código._
```bash
./logscope -input testdata/access.log -memprofile mem.prof
go tool pprof -http=:8081 mem.prof
```

**Explore  `go tool pprof -help`**

---

### `go bug`

**Reportar Bugs para a turma da marmota**
_Comportamento estranho da lingageum ou do toolchain? Suspeita de bug no próprio Go?_
```bash
go bug
```

---

### `go work`
**Trabalhar com múltiplos módulos Go locais simultaneamente sem precisar publicá-los.**
_Inicializar um workspace Go no diretório atual_
```bash
go work init .
```

---

### `go telemetry`
**Gerencia o envio de dados de uso e erros das ferramentas Go para os times do Go.**

_Ver o status atual de telemetria_
```bash
go telemetry
```

_Coleta os dados normalmente, mas não envia nada_
```bash
go telemetry local
```