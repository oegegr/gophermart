.PHONY: generate-main generate-accrual generate-all

.PHONY: generate install-tools

# Установка инструментов
install-tools:
	go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# Переменные для путей
BIN_DIR := $(shell go env GOPATH)/bin
OAPI_CODEGEN := $(BIN_DIR)/oapi-codegen
SPEC_DIR := spec
GEN_DIR := internal/models
API_PKG := api
ACCRUAL_PKG := accrual


export PATH := $(BIN_DIR):$(PATH)

.PHONY: verify
verify: 
	go fmt ./...
	golangci-lint run ./...


.PHONY: test-all 
test-all: 
	go test ./...

.PHONY: build-gophermart
build-gophermart: 
	    rm -rf bin 
		mkdir -p bin
		chmod +x -R bin
		go build -o bin ./...

.PHONY: run-with-db
run-with-db: build-gophermart run-postgresql run-accrual 
		RUN_ADDRESS=127.0.0.1:8080 \
		ACCRUAL_SYSTEM_ADDRESS=http://127.0.0.1:8081/ \
		DATABASE_URI=postgres://admin:admin@172.28.1.1:5432/gophermart?sslmode=disable \
		bin/gophermart

.PHONY: run-with-db-win
run-with-db-win: build-gophermart run-postgresql  
		SERVER_ADDRESS=127.0.0.1:8080 \
		DATABASE_URI=postgres://admin:admin@127.0.0.1:5432/gophermart?sslmode=disable \
		bin/gophermart

.PHONY: run-accrual
run-accrual: 
	pkill accrual || true
	nohup cmd/accrual/accrual_linux_amd64 -a 127.0.0.1:8081 -d postgres://admin:admin@127.0.0.1:5432/accrual?sslmode=disable &

.PHONY: run-postgresql
run-postgresql: 
	docker rm -f $$(docker ps -q  -f=name=postgres) || true
# 	docker volume rm postgres-data || true
	docker run -d --name postgres \
	  -e POSTGRES_USER=admin \
	  -e POSTGRES_PASSWORD=admin \
	  -e POSTGRES_DB=gophermart \
	  -p 172.28.1.1:5432:5432 \
	  -v postgres-data:/var/lib/postgresql/data \
	  postgres:latest 
	sleep 5

# Генерация кода для основного API
generate-api:
	oapi-codegen \
		-generate types \
		-package $(API_PKG) \
		-o $(GEN_DIR)/$(API_PKG)/types.gen.go \
		$(SPEC_DIR)/api.yml

# Генерация кода для API начислений
generate-accrual:
	oapi-codegen \
		-generate types \
		-package $(ACCRUAL_PKG) \
		-o $(GEN_DIR)/$(ACCRUAL_PKG)/types.gen.go \
		$(SPEC_DIR)/accrual.yml
	
	oapi-codegen \
		-generate client \
		-package $(ACCRUAL_PKG) \
		-o $(GEN_DIR)/$(ACCRUAL_PKG)/client.gen.go \
		$(SPEC_DIR)/accrual.yml

# Генерация всего кода
generate: generate-api generate-accrual

# Проверка установки oapi-codegen
check-tools:
	@which oapi-codegen || (echo "oapi-codegen not found. Run 'make install-tools' first." && exit 1)

# Очистка сгенерированных файлов
clean:
	rm -rf $(GEN_DIR)/*