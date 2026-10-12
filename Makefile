.SILENT:
.NOTPARALLEL:

.DEFAULT_GOAL := help

## Settings
SHELL := /bin/sh
.SHELLFLAGS := -eu -c

MIGRATIONS_SERVICE := migrations

export RUN_AS_USER := $(shell id -u)

-include .env
export

## Colors
COLOR_RESET   = \033[0m
COLOR_INFO    = \033[32m
COLOR_COMMENT = \033[33m
COLOR_MAGENTA = \033[35m

help:
	printf "${COLOR_COMMENT}Usage:${COLOR_RESET}\n"
	printf " make [target]\n\n"
	printf "${COLOR_COMMENT}Available targets:${COLOR_RESET}\n"
	awk '/^[a-zA-Z\-_0-9\.@]+:/ { \
	helpMessage = match(lastLine, /^## (.*)/); \
	if (helpMessage) { \
		helpCommand = substr($$1, 0, index($$1, ":")); \
		helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
		printf " ${COLOR_INFO}%-30s${COLOR_RESET} %s\n", helpCommand, helpMessage; \
	} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)


## Docker / оточення

## Запустити бази даних
container@start:
	docker compose up -d --wait postgresql mariadb clickhouse
.PHONY: container@start

## Зупинити оточення
container@stop:
	docker compose down --remove-orphans
.PHONY: container@stop

## Перезапустити бази даних
container@restart: container@stop container@start
.PHONY: container@restart

## Зібрати або оновити образи
container@build:
	docker compose pull
	docker compose build --pull
.PHONY: container@build

## Переглянути логи всіх сервісів
container@logs:
	docker compose logs -f
.PHONY: container@logs

## Shell у контейнері мігратора
container@shell:
	docker compose run --rm --entrypoint sh $(MIGRATIONS_SERVICE)
.PHONY: container@shell

## Перевірити стан контейнерів
container@ps:
	docker compose ps
.PHONY: container@ps


## Міграції

## Створити SQL-міграцію: make migrations@create DB=postgres-main name=add_users_table.sql
migrations@create:
	@if [ -z "$(name)" ]; then printf 'Usage: make migrations@create DB=postgres-main name=add_users_table.sql\n' >&2; exit 1; fi
	docker compose run --rm $(MIGRATIONS_SERVICE) -db="$(DB)" create "$(name)"
.PHONY: migrations@create

## Застосувати міграції
migrations@up:
	docker compose run --rm $(MIGRATIONS_SERVICE) -db="$(DB)" up
.PHONY: migrations@up

## Застосувати одну наступну міграцію
migrations@up-one:
	docker compose run --rm $(MIGRATIONS_SERVICE) -db="$(DB)" up-by-one
.PHONY: migrations@up-one

## Застосувати міграції до заданої версії: make migrations@up-to DB=postgres-main version=20261010120000
migrations@up-to:
	@if [ -z "$(version)" ]; then printf 'Usage: make migrations@up-to DB=postgres-main version=20261010120000\n' >&2; exit 1; fi
	docker compose run --rm $(MIGRATIONS_SERVICE) -db="$(DB)" up-to "$(version)"
.PHONY: migrations@up-to

## Відкотити останню міграцію
migrations@down:
	docker compose run --rm $(MIGRATIONS_SERVICE) -db="$(DB)" down
.PHONY: migrations@down

## Відкотити міграції до заданої версії
migrations@down-to:
	@if [ -z "$(version)" ]; then printf 'Usage: make migrations@down-to DB=postgres-main version=20261010120000\n' >&2; exit 1; fi
	docker compose run --rm $(MIGRATIONS_SERVICE) -db="$(DB)" down-to "$(version)"
.PHONY: migrations@down-to

## Повторно виконати останню міграцію
migrations@redo:
	docker compose run --rm $(MIGRATIONS_SERVICE) -db="$(DB)" redo
.PHONY: migrations@redo

## Відкотити всі міграції
migrations@reset:
	docker compose run --rm $(MIGRATIONS_SERVICE) -db="$(DB)" reset
.PHONY: migrations@reset

## Переглянути статус міграцій
migrations@status:
	docker compose run --rm $(MIGRATIONS_SERVICE) -db="$(DB)" status
.PHONY: migrations@status

## Переглянути поточну версію міграцій
migrations@version:
	docker compose run --rm $(MIGRATIONS_SERVICE) -db="$(DB)" version
.PHONY: migrations@version

## Виправити номери міграцій
migrations@fix:
	docker compose run --rm $(MIGRATIONS_SERVICE) -db="$(DB)" fix
.PHONY: migrations@fix


## Допоміжні команди

## Перевірити Go-код
project@check:
	docker compose run --rm $(MIGRATIONS_SERVICE) go vet ./...
.PHONY: project@check

## Запустити тести
project@test:
	docker compose run --rm $(MIGRATIONS_SERVICE) go test ./...
.PHONY: project@test

## Перевірити форматування Go-коду
project@fmt:
	docker compose run --rm $(MIGRATIONS_SERVICE) gofmt -w .
.PHONY: project@fmt