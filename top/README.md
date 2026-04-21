# Roteiro — O Início do Iceberg da CLI do Go

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

## Bloco 0 — O projeto

> Sem terminal. Abre o editor e mostra o código.

- _"`goprobe` — uma sonda do ambiente de desenvolvimento. Criado para essa apresentação porque precisava de um projeto que exigisse de nós exatamente os comandos que quero mostrar."_
- O que ele faz: verifica se Go, Git, Docker e kubectl estão instalados; exibe informações do SO
- Estrutura em três arquivos-chave:
  - `main.go` → entrypoint com os subcomandos
  - `internal/checker/` → lógica + testes + benchmarks
  - `internal/sysinfo/` → três implementações, uma por plataforma — spoiler do que vem
- _"O projeto está aqui. Sem dependências, sem binário, sem testes rodados. Vamos resolver isso agora."_

---

## Bloco 1 — `go version`

- Roda `go version` → versão do runtime
- _"Mas olha o que `-m` e `-json` fazem juntos — descobrimos isso no `go help version`."_
- Roda `go version -m -json $(which docker)` → lista os módulos Go embutidos no binário do Docker
- _"Útil para auditar dependências de um binário em produção sem ter o código-fonte."_

```bash
go version
go version -m -json $(which docker)
```

---

## Bloco 2 — `go mod init` + `go.mod`

- _"Antes de baixar qualquer coisa, o projeto precisa de um módulo."_
- Mostra o `go.mod` limpo — só `module` e `go version`, sem dependências ainda
- _"Esse arquivo é o contrato do projeto. O nome do módulo é o caminho de import — funciona mesmo sem estar no GitHub."_

```bash
cat go.mod
```

---

## Bloco 3 — `go get`

- _"Vou adicionar a lib `fatih/color` para colorir o terminal do goprobe."_
- Roda `go get github.com/fatih/color@latest` → `go.mod` ganha o `require`, `go.sum` é criado
- Aponta no output: `upgraded go 1.22 => 1.25.0` — _"a diretiva `go` no `go.mod` não é a versão instalada, é a versão mínima exigida pelo grafo de dependências. O `fatih/color` exige Go ≥ 1.25 — por isso atualizou. Você ainda compila com o 1.26."_
- _"O `go.sum` é o lockfile: hash criptográfico de cada módulo. Ninguém substitui uma dep sem o Go perceber."_
- Menciona: `-u` atualiza tudo, `-u=patch` só patch — `go help get` lista essas flags

```bash
go get github.com/fatih/color@latest
cat go.mod
cat go.sum | head -5
```

---

## Bloco 4 — `go mod tidy`

- _"O `go get` adicionou a dep direta. O tidy resolve as indiretas e remove o que não usa."_
- Roda `go mod tidy` → registra `go-colorable`, `go-isatty`, `golang.org/x/sys`
- _"Quem nunca falou: 'roda um tidy pra ver se resolve'?"_

```bash
go mod tidy
cat go.mod
```

---

## Bloco 5 — `go test`

- _"Módulo configurado, dependências resolvidas. Antes de rodar qualquer coisa, quero saber se o código está de pé — `go test` é o orquestrador nativo, sem pytest, sem jest."_
- Suite completa → todos passam
- Verbose → cada teste com PASS/FAIL e tempo
- Benchmarks → as colunas: iterações | `ns/op` = latência | `B/op` e `allocs/op` = pressão no GC
- _"Verde. Agora sim posso rodar. `go help testflag` lista todas as flags de benchmark."_

```bash
go test ./...
go test -v ./internal/checker/
go test -bench=. ./internal/checker/
go test -bench=BenchmarkCheckAll -benchmem ./internal/checker/
```

---

## Bloco 6 — `go run`

- _"Testes passando. Agora quero ver o projeto rodando — mas sem compilar ainda. `go run` executa direto na memória."_
- Roda `go run . version` e `go run . tools`
- _"Agora uma flag pouco conhecida — está no `go help run`: sufixo de versão `@latest`."_
- Roda `go run golang.org/x/vuln/cmd/govulncheck@latest .`
- _"O Go ignora o `go.mod` atual e roda em modo isolado. Ferramenta externa sem contaminar suas dependências."_

```bash
go run . version
go run . tools
go run golang.org/x/vuln/cmd/govulncheck@latest .
```

---

## Bloco 7 — Build Constraints

- _"O goprobe lê informações do sistema. Mas Linux, macOS e Windows têm APIs diferentes. Como o Go resolve isso?"_
- Abre os três arquivos → cada um com `//go:build <os>` na linha 1
- _"Mesmo pacote, três implementações. O compilador escolhe qual incluir. `go help buildconstraint` lista todos os valores válidos para GOOS e GOARCH."_
- Cross-compile ao vivo:

```bash
cat internal/sysinfo/sysinfo_linux.go    # //go:build linux  — lê /proc
cat internal/sysinfo/sysinfo_darwin.go   # //go:build darwin — usa sysctl
cat internal/sysinfo/sysinfo_windows.go  # //go:build windows — usa wmic
GOOS=windows go build -o goprobe.exe .
ls -lh goprobe.exe
rm goprobe.exe
```

---

## Bloco 8 — `go build`

- _"Agora compilo para a minha plataforma."_
- Roda `go build -o goprobe .` — a flag `-o` nomeia o binário, `go help build` lista várias outras: `-race`, `-ldflags`, `-tags`
- Demonstra o binário ao vivo
- _"Nativo. Sem runtime, sem JVM. Copia para qualquer Linux amd64 e roda."_
- Fecha o ciclo: `go version -m ./goprobe` — voltamos ao Bloco 1, agora inspecionando o que compilamos

```bash
go build -o goprobe .
ls -lh goprobe
./goprobe tools
./goprobe sysinfo
go version -m ./goprobe
```

---

## Bloco 9 — `go install`

- _"`go build` gera o binário aqui na pasta. `go install` compila e move para `$GOPATH/bin` — disponível globalmente."_
- Antes: `go env GOPATH` → mostra onde fica esse diretório na máquina — muita gente nunca viu
- A flag `-n` imprime os passos sem executar — está em `go help install`
- _"Olha o que aparece: mkdir temporário, compilação, link, mv para `$GOPATH/bin`. O toolchain sem segredos."_
- Instala de verdade e prova que chegou lá — sem depender do `$PATH` estar configurado

```bash
go env GOPATH
go install -n .
go install .
ls -lh $(go env GOPATH)/bin/goprobe
$(go env GOPATH)/bin/goprobe version
```