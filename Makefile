# Makefile portável entre Linux, macOS e Windows.
#
# A maior parte dos alvos chama apenas `go` e `docker compose`, que já são
# multiplataforma. O que depende do shell fica isolado nas variáveis abaixo.

# Detecta Windows para definir a extensão do binário.
ifeq ($(OS),Windows_NT)
    BINARY := bin/api.exe
    # Distingue cmd.exe puro de Git Bash/MSYS: em cmd, "echo $OSTYPE" não
    # expande e devolve a string literal; num shell POSIX vira "msys".
    ifeq ($(shell echo $$OSTYPE),$$OSTYPE)
        WIN_CMD := 1
    endif
else
    BINARY := bin/api
endif

# Comandos que não existem igualmente nos dois mundos.
ifdef WIN_CMD
    SET_UTF8 := @chcp 65001 >NUL
    RM_BIN   := if exist bin rd /s /q bin & if exist coverage.out del /q coverage.out
    MAKE_ENV := if exist .env (echo .env já existe, nada a fazer.) else (copy .env.example .env >NUL && echo .env criado a partir do .env.example)
    # O cmd preserva os espaços do echo; aspas apareceriam na saída.
    Q :=
else
    SET_UTF8 := @true
    RM_BIN   := rm -rf bin coverage.out
    MAKE_ENV := if [ -f .env ]; then echo ".env já existe, nada a fazer."; else cp .env.example .env && echo ".env criado a partir do .env.example"; fi
    # Sem aspas o sh colapsa os espaços e quebra o alinhamento das colunas.
    Q := "
endif

# Carrega o .env (se existir) para que os alvos conheçam as credenciais do banco.
ifneq (,$(wildcard .env))
include .env
export
endif

DB_USER ?= postgres
DB_NAME ?= ecommerce
DB_PORT ?= 5432

COMPOSE := docker compose
COLLECTION := postman/ecommerce-api.postman_collection.json

# Sobrescreva na linha de comando para filtrar os testes:
#   make test-one T=TestCreateHashesPassword
#   make test-one T=TestCreateHashesPassword PKG=./internal/service
PKG ?= ./...
T   ?= .

.DEFAULT_GOAL := help
.PHONY: help env deps run dev build db-up db-down db-reset db-logs db-shell fmt vet test test-v test-one cover test-api clean

help:
	$(SET_UTF8)
	@echo $(Q)Comandos disponíveis:$(Q)
	@echo $(Q)  make env        Cria o .env a partir do .env.example, sem sobrescrever$(Q)
	@echo $(Q)  make deps       Instala/sincroniza as dependências do Go$(Q)
	@echo $(Q)  make db-up      Sobe o Postgres via Docker e aguarda ficar pronto$(Q)
	@echo $(Q)  make db-down    Para o Postgres mantendo os dados$(Q)
	@echo $(Q)  make db-reset   Destrói o volume e recria o banco do zero$(Q)
	@echo $(Q)  make db-logs    Acompanha os logs do Postgres$(Q)
	@echo $(Q)  make db-shell   Abre um psql dentro do container$(Q)
	@echo $(Q)  make run        Sobe a API$(Q)
	@echo $(Q)  make dev        Fluxo completo: .env + banco + API$(Q)
	@echo $(Q)  make build      Compila o binário em $(BINARY)$(Q)
	@echo $(Q)  make fmt        Formata o código$(Q)
	@echo $(Q)  make vet        Roda a análise estática$(Q)
	@echo $(Q)  make test       Executa todos os testes unitários$(Q)
	@echo $(Q)  make test-v     Idem, com a saída detalhada de cada teste$(Q)
	@echo $(Q)  make test-one   Roda um teste específico: make test-one T=TestNome$(Q)
	@echo $(Q)  make cover      Relatório de cobertura por função$(Q)
	@echo $(Q)  make test-api   Roda a collection do Postman via Newman$(Q)
	@echo $(Q)  make clean      Remove os artefatos de build$(Q)

env:
	$(SET_UTF8)
	@$(MAKE_ENV)

deps:
	go mod tidy
	go mod download

# --wait bloqueia até o healthcheck do Compose passar, sem precisar de um
# laço de espera em shell (que não seria portável para o cmd.exe).
db-up:
	$(COMPOSE) up -d --wait
	$(SET_UTF8)
	@echo Postgres pronto na porta $(DB_PORT)

db-down:
	$(COMPOSE) down

db-reset:
	$(COMPOSE) down -v
	@$(MAKE) --no-print-directory db-up

db-logs:
	$(COMPOSE) logs -f postgres

db-shell:
	$(COMPOSE) exec postgres psql -U $(DB_USER) -d $(DB_NAME)

run:
	go run .

dev: env db-up run

build:
	go build -o $(BINARY) .

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test ./...

test-v:
	go test -v ./...

# T aceita uma regex: TestCreate, TestCreate|TestUpdate, TestX/subteste.
test-one:
	go test -v -run "$(T)" $(PKG)

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Roda a collection do Postman via Newman. Precisa do Node e da API no ar.
test-api:
	npx --yes newman run $(COLLECTION)

clean:
	$(RM_BIN)
