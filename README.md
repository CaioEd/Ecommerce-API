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
| `make test`    | Executa os testes                                          |
| `make clean`   | Remove os artefatos de build                               |

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
