ifeq ($(POSTGRES_SETUP_TEST),)
	POSTGRES_SETUP_TEST := user=user password=qwerty dbname=pickup_point_db host=localhost port=5432 sslmode=disable
endif

INTERNAL_PKG_PATH=$(CURDIR)/internal
MIGRATION_FOLDER=$(INTERNAL_PKG_PATH)/db/migrations

.PHONY: migration-create
migration-create:
	goose -dir "$(MIGRATION_FOLDER)" create "$(name)" sql

.PHONY: test-migration-up
test-migration-up:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_TEST)" up

.PHONY: test-migration-down
test-migration-down:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_TEST)" down

.PHONY: integration-test
integration-test:
	go test ./... -v --tags=integration

.PHONY: unit-tests
unit-tests:
	go test ./...

build:
	docker-compose build

up-all:
	docker-compose up -d zookeeper kafka1 kafka2 kafka3

down:
	docker-compose down

run:
	go run ./cmd/route-kafka

generate:
	rm -rf internal/pb
	mkdir -p internal/pb

	protoc \
		--proto_path=api/ \
		--go_out=internal/pb \
		--go-grpc_out=internal/pb \
		api/*.proto


up:
	docker-compose app --build
	docker-compose up -d postgres app grafana prometheus node-exporter jaeger
