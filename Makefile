
DOCKER_COMPOSE := docker-compose
DB_SERVICE := db

USE_EXTERNAL_DB   ?= no
DB_UP_NEEDED      := $(if $(filter yes,$(USE_EXTERNAL_DB)),,db-up)
DB_ADDR           ?= $(if $(filter yes,$(USE_EXTERNAL_DB)),,localhost:5432)
DB_CHECK_TIMEOUT  ?= 30
DB_CHECK_INTERVAL ?= 2

SCRIPTS := ./scripts
WAIT_DB_READY     := $(SCRIPTS)/wait-db-ready.sh
MIGRATE           := $(SCRIPTS)/migrate.sh

all: generate build

build: ## Build
	go build -o ./bin/fillnames ./cmd/fillnames

generate: ## Generate
	go generate ./...

help: ## Display this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

db-up: ## Start only database
	@if [ -z "$$($(DOCKER_COMPOSE) ps -q $(DB_SERVICE))" ]; then \
		$(DOCKER_COMPOSE) up -d $(DB_SERVICE) && \
		$(WAIT_DB_READY); \
	fi

db-down: ## Stop database
	$(DOCKER_COMPOSE) down $(DB_SERVICE)

db-down-volumes: ## Stop database and remove database volumes
	$(DOCKER_COMPOSE) down -v $(DB_SERVICE)

migrate-up: $(DB_UP_NEEDED) ## Apply all migrations
	DB_ADDR=$(DB_ADDR) $(MIGRATE) up

migrate-down: $(DB_UP_NEEDED) ## Rollback last migration
	DB_ADDR=$(DB_ADDR) $(MIGRATE) down
