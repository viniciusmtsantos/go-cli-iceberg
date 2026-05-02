# logscope — O Iceberg da CLI do Go

> **Antes de começar:** rode `./reset.sh` no terminal (não mostrar para o público).

---

## 🎯 Entenda o Projeto em 2 Minutos

**Logscope é um analisador de HTTP access logs** — te mostra estatísticas sobre quem acessou sua API:
- Qual foi a latência média? (e p50, p95, p99)
- Quantas requisições falharam vs sucesso?
- Qual endpoint foi mais popular?

### Os Dois Modos

**🔄 Modo 1: Gerar dados**
```bash
go run ./cmd/logscope -gen -lines 1000 -output testdata/access.log
```
Cria 1000 linhas de log **realistas** (fake) com timestamps, métodos HTTP, paths, status codes, latências.

**📊 Modo 2: Processar e analisar**
```bash
go run ./cmd/logscope -input testdata/access.log
```
Lê o arquivo, faz math nos dados, mostra um relatório com barras visuais:
```
Entries analyzed:  1000
Latency:  avg=147ms  p50=42ms  p95=592ms  p99=1.7s
Status Codes:  ████ 200 (625)  ██ 404 (243)  █ 500 (132)
Top Endpoints:  /api/users (256)  /api/products (128)  /health (107)
```

**Pipeline (1 linha):**
```bash
go run ./cmd/logscope -gen -lines 1000 | go run ./cmd/logscope
```

### Por que esse projeto?

Não é um projeto "real" — é **educacional**. Serve pra demonstrar comandos Go que você realmente usa em projetos:
- `go test` → tem dois processors (seguro vs com race condition intencional)
- `go fmt` → mostra formatação quebrada
- `go vet` → bug de format string fica óbvio: `%!s(int=1000)`
- `go run`, `go build`, `go install`, `go generate`
- Benchmarks, profiling, tracing, build tags, CI commands

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
- _"A estrutura do projeto segue as convenções do módulo Go — `go help modules` documenta o sistema de módulos."_
- _"O projeto está aqui. Sem binário, sem artefatos. Vamos resolver isso agora."_

---

### Bloco 1 — `go version`

- Roda `go version` → versão do runtime
- _"Mas olha o que `-m` e `-json` fazem juntos — descobrimos isso no `go help version`."_

```bash
go version
go version $(which docker)
```

> _"Podemos ver o binario como uma caixa preta que só vai ser executada, mas com version -m a gente consegue revelar algumas informações do nosso executavel que podem ser uteis para auditoria."_

```
go version -m -json $(which docker)

/usr/bin/docker: go1.24.5 -- A versão exata do compilador Go que foi usada para gerar este binário
        path    github.com/docker/cli/cmd/docker 
        - O caminho do pacote principal
        build   -buildmode=pie 
        - Position Independent Executable. É uma flag de segurança que carrega o binário em locais de memória aleatórios toda vez que ele é executado, dificultando a exploração de vulnerabilidades em memória
        build   -compiler=gc 
        - Indica qual compilador foi usado. gc é o compilador padrão
        build   -ldflags=" -X \"github.com/docker/cli/cli/version.GitCommit=980b856\" -X \"github.com/docker/cli/cli/version.BuildTime=2025-07-25T11:34:09Z\" -X 
        \"github.com/docker/cli/cli/version.Version=28.3.3\" -X \"github.com/docker/cli/cli/version.PlatformName=Docker Engine - Community\""
        - vamos ver estes carinhas aqui mais pro fundo do iceberg
        build   -tags=pkcs11
        build   DefaultGODEBUG=asynctimerchan=1,gotestjsonbuildtext=1,gotypesalias=0,httplaxcontentlength=1,httpmuxgo121=1,httpservecontentkeepheaders=1,multipathtcp=0,netedns0=0,panicnil=1,randseednop=0,rsa1024min=0,tls10server=1,tls3des=1,tlsmlkem=0,tlsrsakex=1,tlsunsafeekm=1,winreadlinkvolume=0,winsymlink=0,x509keypairleaf=0,x509negativeserial=1,x509rsacrt=0,x509usepolicies=0
        build   CGO_ENABLED=1
        build   GOARCH=amd64
        build   GOOS=linux
        build   GOAMD64=v1
```

> _"Útil para auditar dependências de um binário sem ter o código-fonte. Vamos guardar para depois que compilarmos o logscope."_

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
> _"Não vou chamar `go mod tidy` agora — vamos usar essa dep nos próximos blocos. `go help modules` documenta o go.mod."_

---

### Bloco 3 — `go test` (primeira tentativa)

> _"Módulo configurado. Antes de rodar qualquer coisa, quero saber se o código está de pé."_

```bash
go test ./...
```

> _"O que você espera? Testes passam ou falham? Bem, o `go test` não roda os testes cegamente — ele passa o código pelo `go vet` primeiro. O go vet faz uma análise estática do código para detectar problemas comuns. Neste caso, falha porque há um bug detectado: `internal/report/report.go:30` — `%s` tentando formatar um `int`."_

> _"Isso é exatamente o que queremos: a suite de qualidade avisando que tem coisa para resolver antes de continuar."_

> _"Aqui tem um detalhe: no Bloco 19, vamos ter um teste (`TestNaiveRace`) que **crash propositalmente** quando roda sem `-race`. Se rodarmos `go test ./...` agora sem proteção, ele bate nesse teste e morre. Por isso usamos `-short` — é um flag que os testes podem consultar para pular comportamentos destrutivos durante development. Vamos ver isso funcionando:"_

```bash
go test -short ./...
```

> _"Agora passa o `-short`: o TestNaiveRace pula, mas **o `go vet` ainda roda** — e aquele bug em `report.go` ainda aparece. Esse é o ponto: `go vet` é independente do `-short`."_
> _"Camada 2: primeiro `go fmt`, depois `go vet` manual. Quando resolver os dois bugs intencionais, o `go test -short ./...` vai passar silenciosamente. `go help test` explica como o `go test` executa o vet internamente. `go help testflag` documenta `-short` e `-run`."_

---

### Bloco 4 — `go run` (executar sem compilar)

> _"Os testes detectaram problemas, mas o código ainda **compila e roda**. Não é um erro fatal — é um aviso. Vamos testar o programa na prática antes de corrigir."_

```bash
go run ./cmd/logscope -gen -lines 1000 -output testdata/access.log
go run ./cmd/logscope -input testdata/access.log
```

> _"`go run` compila e executa na memória — sem deixar um binário no disco. Primeiro comando: gera 1000 linhas de log de teste e salva em `testdata/access.log`. Segundo: lê esse arquivo e mostra o relatório."_
> _"O programa funciona normalmente. A saída mostra `Entries analyzed: %!s(int=1000)` — essa bagunça é o bug de format string que vimos no Bloco 3 em ação. Em produção, isso corromperia o log."_

> _"Agora vamos verificar vulnerabilidades conhecidas nas dependências:"_

```bash
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

> _"Dois pontos importantes aqui: (1) `go run` com módulo externo — o Go baixa, compila e executa em um ambiente isolado, sem afetar seu `go.mod`. (2) O comando varre todas as suas dependências procurando por CVEs públicas. Se encontrasse algo grave, falharia com exit code 1 — bloquearia o build em CI. Aqui, sem vulnerabilidades no seu código."_

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

> _"Parâmetros, tipos, comportamento. No terminal, offline. `go help doc` documenta as flags do go doc."_
> _"Agora no nosso próprio código:"_

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

> _"Silêncio. Zero configuração, zero debate — o `go fmt` tem uma opinião e ela é definitiva. Em CI: `gofmt -l . | grep .` retorna exit 1 se qualquer arquivo estiver fora do padrão. `go help fmt` documenta as opções."_

---

### Bloco 9 — `go vet`

> _"Formatação resolvida. Mas o projeto ainda tem um bug que o `go fmt` não vê. Vamos chamar o caçador:"_

```bash
go vet ./...
```

> _"Bug em `internal/report/report.go`: a linha `Entries analyzed` usa `%s` para formatar `stats.Total`, que é um `int`. O log vai imprimir `%!s(int=1234)` em produção — exatamente o que vimos no Bloco 3."_
> _"É o mesmo bug que `go test` detectou no Bloco 3 — mas agora estamos investigando de propósito."_

> _"`go vet` usa a mesma `go/ast` que o compilador usa internamente. Ele inspecionou a árvore sintática e cruzou os tipos dos argumentos com os verbos do format string. `go help vet` lista os analisadores disponíveis."_
> _"Corrige: `%s` → `%d` na linha `Entries analyzed`."_

```bash
# edita internal/report/report.go: troca %s por %d em fmt.Fprintf(tw, "  Entries analyzed:..."
go vet ./...
```

> _"Silêncio. O projeto está limpo. Agora o `go test` vai passar:"_

```bash
go test -short ./...
go test -v ./internal/parser/
```

> _"Todos os testes passando. O flag `-short` pula testes que crasham propositalmente — veremos um deles no Bloco 19."_

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
go env -w GOFLAGS=-trimpath
go env GOENV
cat $(go env GOENV)
```

> _"`go env -w` escreveu numa config global do Go — persiste entre projetos, terminais e reboots. Sem `.bashrc`, sem `.zshrc`, sem configurar de novo no próximo setup."_

```bash
go env -u GOFLAGS
```

> _"`go help environment` lista todas as variáveis de ambiente do Go."_

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

> _"O JSON é para automação. Num pipeline de CI: se alguma dep estiver N versões atrás, falha o build. `go help list` documenta todos os formatos de saída."_

---

### Bloco 12 — `go mod why`

> _"O `go list` mostrou `golang.org/x/sys` na lista. Você não adicionou essa dep. De onde ela veio?"_

```bash
go mod graph | grep sys
```

> _"O `go mod graph` mostra o grafo completo: `logscope` → `golang.org/x/sys` e a cadeia toda: `fatih/color` → `go-isatty` → `golang.org/x/sys`. Quando aparece uma lib desconhecida no `go.sum`, esse é o caminho de investigação."_

```bash
go mod why golang.org/x/sys
```

> _"Diz que a main module não precisa — porque nenhum `import` do nosso código usa `golang.org/x/sys` diretamente. Mas ela está no `go.mod`. Por quê? Porque adicionamos `fatih/color` com `go get` mas não usamos no código. O `go mod tidy` vai limpar isso no Bloco 14. `go help mod` documenta todos os subcomandos de módulo."_

---

### Bloco 13 — `go test -cover`

> _"Os testes passam. Mas quanto do código está sendo testado?"_

```bash
go test -short -cover ./internal/...
```

> _"Tem o percentual de cobertura. Mas onde estão as linhas descobertas? Para saber, precisa de um perfil:"_

```bash
go test -short -coverprofile=coverage.out ./internal/...
go tool cover -func=coverage.out
```

> _"Por função. Dá para ver exatamente qual caminho de código nunca foi exercitado. E tem o HTML:"_

```bash
go tool cover -html=coverage.out
```

> _"`go tool cover` já vem no toolchain — sem instalar nada. Linhas vermelhas = não cobertas. É a forma mais visual de entender o que falta testar. `go help testflag` e `go help cover` documentam as opções de cobertura."_

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
go test -short ./internal/...
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

> _"Apagou os módulos baixados. Libera espaço considerável em projetos grandes. Use quando suspeitar de uma dep corrompida ou quiser garantir um build 100% limpo. `go help clean` documenta todos os flags de limpeza."_

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
go test -run=^$ -bench=BenchmarkProcessSafe -benchmem ./internal/processor/
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
go test -short -count=3 ./internal/...
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

> _"Heap profile capturado após o GC — mostra onde o logscope aloca memória. `go help tool` lista todos os sub-tools disponíveis."_

---

### Bloco 23 — `go tool trace`

> _"O pprof mostra onde o tempo vai. O trace mostra **como** o tempo está distribuído entre goroutines."_

```bash
logscope -input access.log -trace trace.out
go tool trace trace.out
```

> _"Você vê goroutines nascendo e morrendo, o scheduler do Go distribuindo trabalho entre os cores, o GC pausando tudo. Nenhuma lib externa — está no toolchain. `go help tool` documenta o trace."_

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
