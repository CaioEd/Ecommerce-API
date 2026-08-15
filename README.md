# Ecommerce API

API REST em **Go + Gin** com **GORM**, seguindo **Layered Architecture**.
Simulação de um e-commerce.

## Arquitetura em camadas

O fluxo de uma requisição segue sempre a mesma direção, cada camada
dependendo apenas da camada imediatamente abaixo:

```
HTTP  ->  Handler  ->     Service    ->       Repository  ->  Banco (GORM)
          (apresentação)  (regra de negócio)  (acesso a dados)
```

| Camada       | Pacote                  | Responsabilidade                                         |
|--------------|-------------------------|----------------------------------------------------------|
| Apresentação | `internal/handler`      | Recebe/valida requisições HTTP e monta respostas         |
| Negócio      | `internal/service`      | Regras de negócio (hash de senha, e-mail único, etc.)    |
| Dados        | `internal/repository`   | Persistência via GORM                                    |
| Domínio      | `internal/model`        | Entidades (tabelas)                                      |
| Transporte   | `internal/dto`          | Objetos de request/response                              |
| Infra        | `internal/database`     | Conexão e migrações                                      |
| Rotas        | `internal/router`       | Registro dos endpoints                                   |
| Config       | `config`                | Carregamento de variáveis de ambiente                    |

## Pré-requisitos

- Go 1.22+
- Docker + Docker Compose (para subir o PostgreSQL)
- `make`

### Windows

O Makefile funciona nos três sistemas. No Windows há três caminhos, todos suportados:

| Ambiente            | Como instalar o `make`                              |
|---------------------|-----------------------------------------------------|
| PowerShell / cmd    | `winget install GnuWin32.Make` ou `choco install make` |
| Git Bash            | já vem com os utilitários Unix; instale o `make` via `choco` |
| WSL                 | `sudo apt install make` (comporta-se como Linux)    |

O Makefile detecta o ambiente e ajusta os comandos: em `cmd.exe` usa `copy`/`rd`,
em shells POSIX usa `cp`/`rm`. O binário do `make build` sai como `bin/api.exe`
no Windows.

Uma ressalva: no Git Bash, `make db-shell` abre um programa interativo e pode
precisar de `winpty` — se o `psql` travar sem mostrar o prompt, rode
`winpty docker compose exec postgres psql -U postgres -d ecommerce`.

## Como rodar

```bash
make dev
```

Esse único comando cria o `.env`, sobe o Postgres no Docker, espera o banco
aceitar conexões e inicia a API. O servidor sobe em `http://localhost:8080` e o
`AutoMigrate` cria a tabela `users` no boot.

Se preferir passo a passo:

```bash
make env     # cria o .env a partir do .env.example
make deps    # instala as dependências do Go
make db-up   # sobe o Postgres e aguarda ficar pronto
make run     # inicia a API
```

## Comandos do Makefile

| Comando        | Descrição                                                  |
|----------------|------------------------------------------------------------|
| `make help`    | Lista todos os comandos disponíveis (padrão)               |
| `make env`     | Cria o `.env` a partir do `.env.example` (não sobrescreve) |
| `make deps`    | Instala/sincroniza as dependências do Go                   |
| `make db-up`   | Sobe o Postgres via Docker e aguarda ficar pronto          |
| `make db-down` | Para o Postgres mantendo os dados                          |
| `make db-reset`| Destrói o volume e recria o banco do zero                  |
| `make db-logs` | Acompanha os logs do Postgres                              |
| `make db-shell`| Abre um `psql` dentro do container                         |
| `make run`     | Sobe a API                                                 |
| `make dev`     | Fluxo completo: `.env` + banco + API                       |
| `make build`   | Compila o binário em `bin/api`                             |
| `make fmt`     | Formata o código                                           |
| `make vet`     | Roda a análise estática                                    |
| `make test`    | Executa todos os testes unitários                          |
| `make test-v`  | Idem, com a saída detalhada de cada teste                  |
| `make test-one`| Roda um teste específico                                   |
| `make cover`   | Relatório de cobertura por função                          |
| `make test-api`| Roda a collection do Postman via Newman                    |
| `make clean`   | Remove os artefatos de build                               |

## Testes unitários

```bash
make test        # roda tudo
make test-v      # tudo, mostrando cada teste
make cover       # cobertura por função
```

Não precisa de banco nem de Docker: os testes usam dublês em memória.

### Rodando um teste específico

```bash
make test-one T=TestCreateHashesPassword
```

O `T` aceita uma regex, e o `PKG` restringe o pacote — útil para não varrer o
projeto inteiro:

```bash
make test-one T=TestCreateHashesPassword PKG=./internal/service
make test-one T='TestDelete|TestGetByID' PKG=./internal/service
make test-one T=TestCreateStatusMapping/e-mail_duplicado PKG=./internal/handler
```

No último exemplo o `/` desce até um subteste da tabela. Os espaços do nome do
subteste viram `_`, que é como o Go os registra.

### Onde ficam os testes

Em Go os arquivos `_test.go` moram **dentro do pacote que testam**, não numa
pasta separada — é isso que dá acesso aos símbolos não exportados e faz o
`go test ./...` e a cobertura funcionarem:

```
internal/service/user_service.go
internal/service/user_service_test.go     <- regras de negócio
internal/handler/user_handler.go
internal/handler/user_handler_test.go     <- tradução HTTP <-> serviço
```

Cada camada é testada isoladamente, com um dublê da camada de baixo definido no
próprio arquivo de teste:

- **service** — um `fakeUserRepository` em memória. Cobre hash da senha, `role`
  padrão, e-mail duplicado, atualização parcial e propagação de falhas de
  infraestrutura.
- **handler** — um `fakeUserService` programável. Cobre o mapeamento de erro
  para status HTTP e garante que um 500 não vaze a mensagem do erro interno.

O `repository` não tem teste unitário: ele é praticamente só chamadas ao GORM,
e testá-lo de verdade exige um Postgres — esse papel fica com a collection do
Postman, que exercita a pilha inteira.

## Testando com o Postman

Importe `postman/ecommerce-api.postman_collection.json` no Postman
(*Import* → *Files*). A collection já traz as variáveis necessárias, então
funciona logo após a importação, com a API no ar.

Ela cobre os 6 endpoints em 15 requisições, divididas em quatro grupos:

| Grupo              | O que faz                                                       |
|--------------------|-----------------------------------------------------------------|
| Health             | Healthcheck                                                     |
| Usuários           | Caminho feliz do CRUD, na ordem                                 |
| Cenários de erro   | 409 duplicado, 400 de validação, 404 inexistente, 400 id inválido |
| Limpeza            | Remove o que foi criado e comprova o soft delete                |

Cada requisição traz asserções (31 no total): status esperado, formato do corpo
e regras de negócio — que o `password` nunca aparece na resposta, que o `role`
assume `customer` quando omitido, e que o `PUT` parcial preserva os campos não
enviados.

**Rode na ordem.** O grupo *Usuários* guarda o `userId` e o `userEmail` em
variáveis de collection que os grupos seguintes reaproveitam — o teste de e-mail
duplicado, por exemplo, depende do usuário criado antes. Use o *Collection
Runner* para executar tudo de uma vez. Cada execução gera e-mails únicos com
timestamp, então a collection pode rodar quantas vezes quiser sem limpar o banco.

Pela linha de comando, com a API no ar:

```bash
make test-api
```

## Banco de dados

O `docker-compose.yml` sobe um `postgres:17-alpine` com healthcheck e volume
nomeado (`ecommerce-pgdata`), então os dados sobrevivem a `make db-down`. Use
`make db-reset` quando quiser começar do zero.

As credenciais vêm do `.env` — o **mesmo arquivo** lido pela aplicação e pelo
Compose, de modo que API e banco nunca saem de sincronia. Alterar `DB_USER`,
`DB_PASSWORD`, `DB_NAME` ou `DB_PORT` exige recriar o volume (`make db-reset`),
já que o Postgres só aplica essas variáveis na primeira inicialização.

## Endpoints

Base: `/api/v1`

| Método | Rota           | Descrição                 |
|--------|----------------|---------------------------|
| GET    | `/health`      | Healthcheck               |
| POST   | `/users`       | Cria um usuário           |
| GET    | `/users`       | Lista todos os usuários   |
| GET    | `/users/:id`   | Busca usuário por ID      |
| PUT    | `/users/:id`   | Atualiza um usuário       |
| DELETE | `/users/:id`   | Remove um usuário (soft)  |

### Exemplo — criar usuário

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Caio Silva",
    "email": "caio@example.com",
    "password": "senha123",
    "role": "customer"
  }'
```

## Próximos passos

Para adicionar novas entidades (ex.: `Product`, `Order`), replique o padrão:

1. Crie a entidade em `internal/model`.
2. Registre-a no `database.Migrate`.
3. Crie `repository`, `service` e `handler` correspondentes.
4. Faça o wiring em `main.go` e registre as rotas em `internal/router`.
