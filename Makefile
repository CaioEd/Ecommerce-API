# Carrega o .env (se existir) para que os alvos conheçam as credenciais do banco.
ifneq (,$(wildcard .env))
include .env
export
endif

DB_USER ?= postgres
DB_NAME ?= ecommerce
DB_PORT ?= 5432

BINARY := bin/api
COMPOSE := docker compose

.DEFAULT_GOAL := help
.PHONY: help env deps run dev build db-up db-down db-reset db-logs db-shell fmt vet test clean

help: ## Lista os comandos disponíveis
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

env: ## Cria o .env a partir do .env.example (não sobrescreve o existente)
	@test -f .env && echo ".env já existe, nada a fazer." || (cp .env.example .env && echo ".env criado a partir do .env.example")

deps: ## Instala/sincroniza as dependências do Go
	go mod tidy
	go mod download

db-up: ## Sobe o Postgres via Docker e aguarda ficar pronto
	$(COMPOSE) up -d
	@printf "aguardando o Postgres aceitar conexões"
	@until $(COMPOSE) exec -T postgres pg_isready -U $(DB_USER) -d $(DB_NAME) >/dev/null 2>&1; do \
		printf "."; sleep 1; \
	done
	@echo " pronto (porta $(DB_PORT))"

db-down: ## Para o Postgres mantendo os dados
	$(COMPOSE) down

db-reset: ## Destrói o volume e recria o banco do zero
	$(COMPOSE) down -v
	@$(MAKE) --no-print-directory db-up

db-logs: ## Acompanha os logs do Postgres
	$(COMPOSE) logs -f postgres

db-shell: ## Abre um psql no container
	$(COMPOSE) exec postgres psql -U $(DB_USER) -d $(DB_NAME)

run: ## Sobe a API (executa o AutoMigrate no boot)
	go run .

dev: env db-up run ## Fluxo completo: .env + banco + API

build: ## Compila o binário em bin/api
	go build -o $(BINARY) .

fmt: ## Formata o código
	go fmt ./...

vet: ## Roda a análise estática
	go vet ./...

test: ## Executa os testes
	go test ./...

clean: ## Remove os artefatos de build
	rm -rf bin
