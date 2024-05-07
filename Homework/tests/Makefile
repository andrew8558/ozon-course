ifeq ($(POSTGRES_SETUP_TEST),)
	POSTGRES_SETUP_TEST := user=test password=test dbname=test host=localhost port=5434 sslmode=disable
endif

INTERNAL_PKG_PATH=$(CURDIR)
MIGRATION_FOLDER=$(INTERNAL_PKG_PATH)/migrations

.PHONY: migration-create
migration-create:
	goose -dir "$(MIGRATION_FOLDER)" create "$(name)" sql

.PHONY: test-migration-up
test-migration-up:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_TEST)" up

.PHONY: test-migration-down
test-migration-down:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_TEST)" down

build:
	docker-compose build

up-all:
	docker-compose up -d zookeeper kafka1

down:
	docker-compose down

run:
	go run ./cmd/route-kafka