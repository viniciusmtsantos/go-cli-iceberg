# logscope — O Iceberg da CLI do Go

> **Antes de começar:** rode `./reset.sh` no terminal (não mostrar para o público).

---

## Abertura — `go help`

> Sem terminal. Só fala.

- _"Antes de qualquer comando, quero te dar um atalho que a maioria ignora."_
- `go help` → lista todos os comandos **e** tópicos de conceito do toolchain
- `go help <command>` → doc completa de qualquer comando, no terminal
- `go help <topic>` → conceitos como `modules`, `buildconstraint`, `testflag`
- _"Podiamos terminar aqui, porque  mais do que uma apresentação sobre comandos de terminal, que tem uma grande tendencia de ser cansativa, é um convite a explorar o que o toolchain do Go tem a oferecer. O `go help`é seu mapa, aproveite ele."_

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

### Por que esse projeto?

Não é um projeto "real" — é **educacional**. Serve pra gente explorar alguns comandos juntos

---

### Bloco 1 — `go version`

- Roda `go version` → versão do seu binario

```bash
go version
```

> _"Mas olha `go help version`, podemos observar a versão do go que compilou o binário"_

```bash
go version $(which docker)
```

> _"Outra coisa, podemos ver o binario como uma caixa preta que só vai ser executada, mas com `-m` a gente consegue revelar algumas informações do nosso executavel que podem ser uteis para auditoria."_

```bash
go version -m -json $(which docker)

/usr/bin/docker: go1.24.5 -- A versão exata do compilador Go que foi usada para gerar este binário
        path    github.com/docker/cli/cmd/docker 
        # O caminho do pacote principal
        build   -buildmode=pie 
        # Position Independent Executable. É uma flag de segurança que carrega o binário em locais de memória aleatórios toda vez que ele é executado, dificultando a exploração de vulnerabilidades em memória
        build   -compiler=gc
        # Indica qual compilador foi usado. gc é o compilador padrão
        build   -ldflags=" -X \"github.com/docker/cli/cli/version.GitCommit=980b856\" -X \"github.com/docker/cli/cli/version.BuildTime=2025-07-25T11:34:09Z\" -X 
        \"github.com/docker/cli/cli/version.Version=28.3.3\" -X \"github.com/docker/cli/cli/version.PlatformName=Docker Engine - Community\""
        # vamos ver estes carinhas aqui mais pro fundo do iceberg
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

> _"Antes de rodar qualquer este cara, o projeto precisa de um módulo."_

```bash
cat go.mod
```

> _"Esse arquivo é o contrato do projeto. Só `module` e `go 1.24` — sem dependências ainda. O nome do módulo é o caminho de import, mesmo sem estar no GitHub."_

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
> _"Puts, deu erro D;"_
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

> _"`go run` compila e executa na memória — sem deixar um binário no disco. Primeiro comando: gera 1000 linhas de log de teste e salva em `testdata/access.log`. Segundo: lê esse arquivo que possui timestamp | method | path | status | latency | bytes e mostra o relatório ."_
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
go build ./cmd/logscope
go build -o new-logscope ./cmd/logscope
ls -lh logscope
./logscope -input testdata/access.log
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
> _"`go fmt` (O Atalho): Formatar o seu projeto rapidamente (go fmt ./...). Ele é simples, não aceita flags complexas e já vem configurado para sobrescrever os arquivos._

> _"`gofmt` (O Motor): É a ferramenta raiz. Aceita flags como -l (para apenas listar arquivos fora do padrão) ou -d (para mostrar o diff)."_

> _"A Diferença: O `go fmt` é um comando de conveniência que chama o `gofmt` por baixo dos panos com configurações fixas."_

> _"Abre o `internal/processor/processor.go` no editor. A função `Avg` está sem indentação. O código compila. O binário funciona. Mas o `go fmt` não aprova."_

```bash
gofmt -l -d .
```

> _"Listou o `processor.go`. Não é um erro — é uma diferença em relação ao padrão canônico do Go. Antes de corrigir, veja exatamente o que vai mudar:"_
> _"+ é o que vai entrar, - é o que vai sair. Agora aplica:"_

```bash
go fmt ./...
go help fmt
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
go test -short -v ./...
```

> _"Todos os testes passando. O flag `-short` pula testes que crasham propositalmente, O terminal vai mostrar explicitamente: SKIP: — veremos um deles no Bloco 19."_

---

### Bloco 10 — `go env`

> _"Agora um desvio: o logscope não tem um comando `env`, mas o toolchain tem — e ele é mais poderoso do que parece."_

```bash
go env
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
go test -short ./internal/...
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

> _"Vamos pensar que o pkg/entry é o nosso vocaulario central. Ela guarda tipos fundamentais da nossa aplicação, como os Niveis de Log. Essa estrutura permite que parser, processor e report conversem sem problema de dependencia. Neste caso, nós definimos o esses níveis de forma eficiente usando enum através do iota. Porém na hora de imprirmir, a gente quer ver INFO, não NIVEL 1. É aqui que o go generate entra"_

_"Ao inves de escrever uma função gigante na mão, a gente deixa um comentário magico no codigo //go:generate.... O go aciona a ferramenta stringer e escreve um arquivo novo inteiro para nós, convertendo estes numeros em texto automaticamente."_

_"Braçal e Manutenção: Isso elimina código braçal neste trecho e cria manutenção automatizada, se precisar criar um outro level de severidade (TRACE) basta rodar o comando novamente"_

_"Performance: o codigo gerado por debaixo dos panos usa mapeamento direto na memoria, sendo muito mais rapido do que validação manual"_

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

> _"Sabe quando o suporte avisa de um bug em produção e você pergunta: 'Qual versão está rodando lá?'. O Júnior lê a versão de um arquivo `.env` ou de um `config.json`. Mas arquivos externos podem ser alterados, esquecidos ou corrompidos."_

> _"O Sênior faz o binário nascer sabendo quem ele é. Ele usa o Linker."_

```bash
go build -o logscope ./cmd/logscope

./logscope -version
go version -m ./logscope

```

```bash
go build \
  -ldflags "-X main.version=1.2.0 -X main.commit=$(git rev-parse --short HEAD)" \
  -o logscope-linker \
  ./cmd/logscope

./logscope-linker -version
go version -m ./logscope-linker
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

> _"Ah, mas a gente compila tudo no Docker dentro da pasta `/app`, famoso FROM AS builder, meu path local não vaza de qualquer jeito. Pra que usar isso?'"_

> _"Porque o `-trimpath` não é só sobre esconder caminhos. É sobre **Builds Reproduzíveis**."_

> _"Ele garante que o binário gerado na sua máquina e o binário gerado no pipeline de CI tenham exatamente o mesmo hash criptográfico. É a prova matemática para a auditoria de que o que está rodando no Kubernetes é exatamente o que foi aprovado no Git."_

```bash
strings ./logscope | grep $(pwd) | head -3
go build -trimpath -o logscope-safe ./cmd/logscope
strings ./logscope-safe | grep $(pwd) || echo "nenhum path local no binário"
```

> _"`-trimpath` remove todos os paths absolutos do binário — builds reproduzíveis, sem leak de estrutura interna do servidor. `go help build` documenta essa flag."_

---

## 🔬 Camada 4 — Testing profundo e observabilidade

### Bloco 18 — Benchmarks

> _"Os testes passam. Mas quão rápido é o parser? Quão rápido é o processor?"_

### Bloco 18 — Benchmarks

> _"Os testes garantem que o código funciona. Mas ele aguenta o tranco em produção?"_

> _"O resultado do benchmark do Go é um dos mais ricos do mercado porque ele te dá três visões de uma vez: Escala, Latência e Custo de Memória."_

> _"O `-bench` ativa os testes de performance, rodando a função milhares de vezes. E o `-benchmem` é o nosso radar de lixo: ele mostra quanta memória estamos alocando e deixando para o Garbage Collector limpar."_

> _"E por que eu escolhi rodar isso logo no `internal/parser`? Porque fatiar strings e converter datas é o **'hot path'** (o caminho quente) da nossa aplicação. Se o nosso analisador de logs for lento e gastar muita memória, o gargalo fatalmente estará aqui."_

```bash
go test -bench=. -benchmem ./internal/parser/
```

> _"Três colunas que importam: iterações | `ns/op` (latência por operação) | `B/op` e `allocs/op` (pressão no GC)."_

> _"1. O Cabeçalho (O Setup):
"Primeiro, ele identifica a máquina. Um i7 com 12 núcleos. E reparem no sufixo -12 do lado do nome de cada teste: o Go nos avisa que usou todo o paralelismo da máquina."

> _"2. Iterações (A Inteligência do Go):
"A primeira coluna de números (6032278). O Go não roda a função uma vez e confia. Ele rodou nosso parser de uma linha mais de 6 milhões de vezes em 1 segundo até garantir que o resultado é estatisticamente perfeito."

> _"3. Latência (ns/op):
"A segunda coluna é a velocidade. 193 ns/op significa que levamos absurdos 193 nanossegundos para processar uma linha. Para vocês terem ideia, isso é menos de 1 milionésimo de segundo. E no teste Parallel, nós derrubamos esse tempo para 45 ns. Provando que o Go distribuiu a carga lindamente entre os núcleos."

> _"4. O 'Radar de Lixo' (B/op e allocs/op):
"Mas é aqui na direita que o Sênior olha de verdade. 112 B/op e 1 allocs/op. Isso quer dizer que para cada linha de texto processada, o Go precisou pedir memória para o sistema operacional apenas UMA vez."

"Por que isso é ouro? Porque memória em Go significa Garbage Collector trabalhando. Uma alocação significa que nosso parser não gera quase nenhum 'lixo'. Nosso programa pode processar gigabytes de log e a CPU não vai engasgar tentando limpar a memória. É latência baixa garantida em produção.""_

> _"Agora comparando workers no processor:"_

```bash
go test -run=^$ -bench=BenchmarkProcessSafe -benchmem ./internal/processor/
```

> _"ProcessSafe com 1 worker vs 8 workers — você vê o speedup à medida que adicionamos goroutines. `go help testflag` documenta todas as flags de benchmark."_

1. A Vitória (O Tempo Caiu):
"Olhem a coluna ns/op (nanossegundos). Com 1 worker, levamos cerca de 6.6 milhões de nanossegundos (6.6 milissegundos). Quando subimos para 8 workers, o tempo caiu para 3.4 milissegundos. Nós cortamos a latência do sistema quase pela metade. O paralelismo funcionou e a CPU foi bem aproveitada."

2. A Lei dos Rendimentos Decrescentes:
"Mas reparem com o olhar de um Sênior: nós multiplicamos os workers por 8, mas o programa não ficou 8 vezes mais rápido. Por quê? Porque criar Goroutines, orquestrar o trabalho e juntar tudo no final (fase de merge) custa tempo de CPU. É a prova de que jogar mais threads em um problema nem sempre escala perfeitamente."

3. O Preço Pago (A Troca Inteligente):
"Agora olhem a última coluna, as alocações (allocs/op). Com 1 worker, fizemos 25 alocações. Com 8 workers, saltamos para 116 alocações de memória. Por que o Go consumiu mais memória se os dados processados eram os mesmos?"

"Porque para o nosso código ser livre de 'Race Conditions' (Corridas de Dados) sem usar Mutexes (travas que deixariam o código lento), nós demos a cada worker o seu próprio mapa de memória isolado. Trocamos um pouco de consumo de RAM por velocidade e segurança extrema. Isso é arquitetura de verdade."


### Bloco 18.1 — Live Coding: Otimização Sênior vs Júnior

> _"Falar de performance é fácil, quero ver no código. Vamos quebrar nosso processador de propósito para ver o erro que mais derruba aplicações Go em produção: esquecer de pré-alocar memória."_

> **Ação (Sem terminal):** Abra `internal/processor/processor.go` na função `ProcessSafe`.
> Altere a criação do slice `latencies` dentro da Goroutine, removendo o `len(chunk)`:
> De: `latencies: make([]time.Duration, 0, len(chunk)),`
> Para: `latencies: make([]time.Duration, 0),`

```bash
go test -run=^$ -bench=BenchmarkProcessSafe -benchmem ./internal/processor/
```

> _"Olhem o estrago na coluna de alocações (`allocs/op`)! Passamos de 116 para milhares de alocações. Sem saber o tamanho final da lista, o Go precisa pausar a execução o tempo todo para pedir mais memória ao SO e copiar os dados. Volte o `len(chunk)` e você salva o Garbage Collector de um infarto."_

> **Ação (Sem terminal):** Desfaça a alteração das latências. Agora vamos otimizar a CPU no Hot Loop.

> _"Nosso código está rápido, mas tem um gargalo invisível no 'Hot Loop' (o laço que processa as linhas). Nós estamos usando `e.Level.String()` como chave do mapa. Calcular hash de texto gasta muita CPU. Nós já temos o nível como um número inteiro (`entry.Level`). Vamos otimizar isso cirurgicamente."_

> **Ação (Sem terminal):** No `internal/processor/processor.go`, faça as 4 mudanças rápidas:
> 1. Na struct `partialStats`: mude `byLevel map[string]int` para `byLevel map[entry.Level]int`.
> 2. Na criação do mapa local: mude para `byLevel: make(map[entry.Level]int, 5)`.
> 3. No Hot Loop: mude `p.byLevel[e.Level.String()]++` para `p.byLevel[e.Level]++`.
> 4. No Merge final das estatísticas: mude `stats.ByLevel[k] += v` para `stats.ByLevel[k.String()] += v`.

```bash
go test -run=^$ -bench=BenchmarkProcessSafe -benchmem ./internal/processor/
```

> _"Olhem o tempo `ns/op` cair na hora! Nós tiramos a conversão de String de dentro do loop que roda centenas de milhares de vezes, e jogamos para a fase final de merge que roda apenas 5 vezes. Trocar texto por número no mapa aliviou a CPU. Otimização real com 4 linhas de código."_

1. O Ganho Real (Apontando para o 1 Worker):
"Olhem a primeira linha, o nosso teste de 1 worker. Nós caímos de 6.6 milhões de nanossegundos para 5.7 milhões. Nós raspamos quase 1 milhão de nanossegundos de latência por operação com apenas 4 linhas de código alteradas. Numa escala de bilhões de logs por dia, isso significa menos servidores no Kubernetes e dinheiro economizado no fim do mês."

2. Por que ficou mais rápido? (O Diagnóstico):
"E reparem na coluna da direita: nós continuamos fazendo as mesmas 25 alocações de memória. Ou seja, o nosso ganho não foi de memória, foi 100% de CPU."

"Sabe qual era o peso que a CPU estava carregando? Calcular o Hash de strings no mapa centenas de milhares de vezes. Texto é pesado. Número inteiro é nativo, direto e barato. Nós tiramos a tradução de texto do 'Hot Loop' e jogamos para o final. Otimização cirúrgica, provada matematicamente pelo próprio Go."

---

### Bloco 19 — `go test -race`

> _"O logscope tem dois processors: `ProcessSafe` (correto) e `ProcessNaive` (com um bug escondido). Vamos rodar o teste do Naive da forma comum."_

> _"E aqui já podemos usar mais uma flag interessante. O `-run` é o seu bisturi. Por padrão, o Go executa todas as funções que começam com 'Test' na pasta. Com o `-run`, você aplica um filtro de Expressão Regular (Regex)."_

> _"Ao digitar `-run TestNaiveRace`, você diz ao compilador: 'Ignore a suíte inteira. Foque APENAS no teste que contém este nome'. Isso é essencial para isolar falhas específicas e limpar o palco, garantindo que o crash de um teste não polua a visualização dos outros que estão passando."_

```bash
go test -run TestNaiveRace ./internal/processor/
```

> _"Boom! Ele crasha feio com `fatal error: concurrent map writes`. O Go tem uma proteção nativa que mata o programa na hora se duas goroutines escreverem num mapa ao mesmo tempo. Isso é ótimo para evitar corrupção de dados."_

> _"Mas olhem o problema: ele morreu e cuspiu um erro genérico do sistema. Ele não te diz ONDE no seu código isso aconteceu. Como você conserta algo que não sabe onde está? É para isso que existe o detector:"_

```bash
go test -race -run TestNaiveRace ./internal/processor/
```

> _"O detector imprimiu os goroutines em conflito, com stack traces. Nenhuma mágica — o compilador instrumenta cada acesso à memória em tempo de build."_

> _"Olhem a diferença! O compilador instrumentou a memória. Em vez de apenas morrer com um erro genérico de mapa, o `-race` pegou a corrida de dados no ato e imprimiu: 'WARNING: DATA RACE'."_

> _"E ele age como um detetive perfeito. Ele mostra a linha exata onde a Goroutine 13 leu a variável e a linha onde a Goroutine 15 escreveu nela (linha 45 do nosso `naive.go`). E não para por aí: ele ainda nos diz exatamente em qual linha de código essa Goroutine nasceu (linha 39, onde declaramos o nosso `go func`). O Go não só avisa que o código vai falhar em produção, ele te dá o mapa do tesouro exato para consertar."_

> _"Agora confirmamos que `ProcessSafe` é livre de corridas:"_

```bash
go test -race -run TestProcessSafe_DeterministicWithRace ./internal/processor/
```

> _"Silêncio. `go help testflag` documenta o `-race`."_

---

### Bloco 20 — `go test -fuzz`

> _"Cobertura de testes garante que o código funciona para as entradas que você imaginou. Fuzzing garante que não quebra para as que você não imaginou."_

### Bloco 20 — `go test -fuzz`

> _"A cobertura de testes comum garante que o código funciona para as entradas que VOCÊ imaginou. Mas em produção, o usuário digita coisas imprevisíveis. O Fuzzing garante que o sistema não quebre para as entradas que você NÃO imaginou. Vamos ver como se escreve um."_

> **Ação (Sem terminal):** Abra o arquivo `internal/parser/parser_fuzz_test.go` no editor.

> _"Olhem a estrutura. O Fuzzer não é cego, ele é inteligente. Primeiro, usamos `f.Add()` para dar alguns exemplos de logs válidos e inválidos. Isso se chama 'Seed Corpus' (Sementes). Nós dizemos ao Go: 'Esses são os formatos esperados, comece mutando a partir daqui'."_

> _"Depois vem o loop `f.Fuzz`. O Go vai gerar aquela variável `data` infinitamente, enfiando bytes nulos, emojis, cortando a string no meio. E qual a única regra do jogo ali dentro? Nosso `ParseReader` não pode dar **Panic**. Retornar um erro de 'linha mal formatada' é aceitável, mas o programa crashar é proibido. Vamos soltar o monstro:"_

```bash
go test -fuzz=FuzzParseReader -fuzztime=10s ./internal/parser/
```

> _"Deixem o comando rodar e olhem a mágica no terminal. O aviso `gathering baseline` é o Go lendo nossas sementes. Logo em seguida, ele ativa todos os núcleos do processador e começa o bombardeio."_

> _"Reparem no número de `execs`. Em apenas 10 segundos, o Go gerou e atirou mais de **100.000 cenários caóticos** contra o nosso código. É uma média absurda de testes por segundo. Nenhuma equipe de QA humana conseguiria fazer isso."_

> _"Ele marca ali `new interesting`. O algoritmo percebeu que uma mutação louca forçou nosso código a entrar num `if` obscuro, e salvou essa string bizarra na pasta do projeto para sempre testá-la no futuro."_

> _"No final, ele dá um 'FAIL - context deadline exceeded'. E isso é uma vitória! Isso significa que o relógio de 10 segundos estourou e o Go puxou a tomada sem encontrar nenhum erro fatal. O código sobreviveu ao bombardeio."_

> _"Dica de Trincheira: Sempre use o `-fuzztime`. Se você rodar sem ele, o teste não para nunca. Na sua máquina, use uns 10 segundos. No pipeline de CI/CD, coloque uns `-fuzztime=5m` (5 minutos). Se você rodar no GitHub Actions sem definir o tempo, ele vai rodar infinitamente e o seu chefe vai te ligar para perguntar por que a conta da nuvem explodiu!"_

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

### Bloco 22 — `go tool` e o poderoso `pprof`

> _"Chegamos na camada de Observabilidade profunda. Mas antes do comando, o que é esse `go tool`? O Go tem os comandos de dia a dia (`build`, `test`, `run`), mas ele também esconde um canivete suíço de baixo nível sob o comando `go tool`. Se você rodá-lo sozinho, verá dezenas de ferramentas internas que o próprio compilador usa."_

> _"E a joia da coroa dessa lista é o `pprof`. Ele é o Profiler oficial do Go. Ele tira um 'raio-X' do seu programa em execução, mapeando cirurgicamente em qual linha de código a CPU está suando ou onde a memória está sendo engolida. Vamos gerar uma carga pesada de 200 mil logs e gravar esse raio-X de CPU:"_

```bash
./logscope -gen -lines 200000 -output testdata/access.log
./logscope -input testdata/access.log -cpuprofile cpu.prof
```

> _"O perfil de CPU foi salvo no arquivo `cpu.prof`. Mas ler esse binário cru é impossível. É aí que a ferramenta brilha. Vamos pedir pro `go tool pprof` ler isso e subir um servidor web interativo:"_

```bash
$ go tool pprof -no_browser -http=:8080 cpu.prof
```

> **Ação (No Navegador):** A interface web vai abrir. Vá no menu "View" no topo esquerdo e clique em **"Flame Graph"**.

> _"Isso aqui já vem de fábrica, sem instalar nenhum software externo! O que um Sênior procura aqui? O 'Flame Graph' (Grafo de Chamas). Ele mostra a pilha de execução de forma visual. Quanto mais largo o retângulo, mais tempo de CPU a função está consumindo. Você bate o olho e acha o gargalo exato do sistema em 5 segundos."_

> _"E não é só CPU. Quer saber onde a RAM está indo?"_

```bash
./logscope -input testdata/access.log -memprofile mem.prof
go tool pprof -no_browser -http=:8081 mem.prof
```

> _"Esse é o Heap Profile. Ele tira uma foto da memória RAM logo após o Garbage Collector passar. Se você tem um vazamento de memória (Memory Leak) em produção travando o Kubernetes, é essa tela que vai te salvar, apontando qual função está esquecendo de liberar os dados."_

---

### Bloco 23 — `go tool trace`

> _"Se o `pprof` é um raio-X estático que mostra ONDE a sua CPU está gastando tempo, o `trace` é uma ressonância magnética em vídeo que mostra COMO as coisas acontecem ao longo do tempo."_

> _"Imagine a seguinte dor de cabeça: seu programa em Go está lento, mas o servidor está com 80% de CPU livre. Você roda o `pprof` e ele não aponta nada de errado. O que está acontecendo? O seu programa não está pesado, ele está **travado**. Goroutines bloqueadas esperando canais, ou o Garbage Collector pausando a aplicação. O `pprof` não enxerga isso, mas o `trace` sim."_

```bash
./logscope -input testdata/access.log -trace trace.out
```

> _"Geramos nosso arquivo de rastreio. Agora, vamos abrir a interface do trace:"_

```bash
go tool trace trace.out
```

> **Ação (No Navegador):** O navegador vai abrir uma página de links simples. Clique no primeiro link: **"View trace"**. *(Dica: Pressione 'W' para dar zoom e 'A/D' para navegar para os lados).*

> _"Isso aqui é o coração do Go batendo. Cada linha verde pontilhada que você vê ali como 'Proc 0, Proc 1, Proc 2' representa um núcleo lógico do seu processador. E as barrinhas coloridas são as suas Goroutines sendo executadas."_

> _"O Sênior usa isso aqui para entender o Scheduler do Go visualmente. Você consegue clicar em uma Goroutine e ver o milissegundo exato em que ela nasceu, quanto tempo ela ficou na fila esperando para rodar, quando ela travou esperando o disco, e o exato momento em que o Garbage Collector decretou um 'Stop The World' e parou tudo para limpar a memória."_

> _"Entender concorrência lendo código é difícil. Entender concorrência olhando para o `trace` é como ter visão de raio-X na arquitetura do sistema."_

---

### Bloco 24 — `go tool nm` (A Caixa Preta)

> _"Chegamos ao fundo do iceberg. Nós compilamos, nós otimizamos a CPU, nós vimos o trace. Mas e o binário final? Nosso programa tem poucos arquivos, mas o binário gerado tem megabytes de tamanho. O que exatamente o compilador enfiou lá dentro?"_

> _"Sabe quando você importa uma biblioteca gigante na AWS, mas só usa uma função pequena? Como ter certeza de que o compilador do Go fez o 'Dead Code Elimination' e não levou o peso morto para produção? O Sênior não supõe, ele prova. E para abrir a caixa preta do binário, usamos o `go tool nm`."_

```bash
go tool nm ./logscope | wc -l
```

> _"Nós escrevemos algumas dezenas de funções, mas o comando listou milhares de símbolos! O linker trouxe o runtime do Go, o Garbage Collector, o agendador de concorrência... Tudo está embutido aí dentro."_

> _"Mas e se o binário estiver absurdamente grande e eu precisar reduzir custos no meu container Docker? Quem é o culpado pelo peso? O `nm` tem uma flag chamada `-size`. Vamos listar as funções mais gordas do nosso executável:"_

```bash
go tool nm -size ./logscope | sort -k2 -rn | head -10
```

> _"Olhem isso! O comando ordenou as funções exatas pelo tamanho em bytes que elas ocupam no disco. Se uma dependência pesada vazou para o seu executável, o nome dela vai aparecer piscando aqui no topo. É assim que você audita o que vai para produção no nível atômico."_

```bash
go tool nm -size ./logscope | sort -k2 -rn | head -10
```

> _"Parece grego, não é? Mas vamos ler isso com olhos de Engenheiro de Software. A segunda coluna é o tamanho em bytes. E olha quem são os maiores vilões de peso do nosso binário:"_

> _"Vemos ali o `runtime.mheap_` ocupando 93KB. Isso é o coração do alocador de memória e do Garbage Collector do Go. Ele está embutido no nosso arquivo."_

> _"Vemos também funções marcadas com a letra 'T' (de Text/Código Executável), como o `time.parse` e o `fmt.printValue`. Por que elas estão aí? Porque nós usamos manipulação de datas e `Printf` no nosso código! O compilador vasculhou nosso projeto, pegou cirurgicamente apenas o que usamos da biblioteca padrão (stdlib) e soldou dentro do executável."_

> _"E é com essa tela que encerramos o nosso Iceberg."_

> _"O Go não precisa que você instale uma Máquina Virtual Java (JVM) ou um Node.js no servidor de produção. Ele empacota o coletor de lixo, o agendador de concorrência, o formatador de texto e o seu código em um único arquivo binário, autossuficiente e otimizado no nível do byte."_

> _"Esse é o iceberg da CLI do Go. A maioria das pessoas conhece apenas a ponta. Vocês agora conhecem o fundo. Muito obrigado!"_

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
