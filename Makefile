.PHONY: generate-main generate-accrual generate-all

.PHONY: generate install-tools

# Установка инструментов
install-tools:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# Переменные для путей
BIN_DIR := $(shell go env GOPATH)/bin
OAPI_CODEGEN := $(BIN_DIR)/oapi-codegen
SPEC_DIR := spec
GEN_DIR := internal/models
API_PKG := api
ACCRUAL_PKG := accrual


export PATH := $(BIN_DIR):$(PATH)

# Генерация кода для основного API
generate-api:
	oapi-codegen \
		-generate types \
		-package $(API_PKG) \
		-o $(GEN_DIR)/$(API_PKG)/types.gen.go \
		$(SPEC_DIR)/api.yml
	
# 	oapi-codegen \
# 		-generate server \
# 		-package $(API_PKG) \
# 		-o $(GEN_DIR)/$(API_PKG)/server.gen.go \
# 		$(SPEC_DIR)/api.yml
	
# 	oapi-codegen \
# 		-generate client \
# 		-package $(API_PKG) \
# 		-o $(GEN_DIR)/$(API_PKG)/client.gen.go \
# 		$(SPEC_DIR)/api.yml
	
# 	oapi-codegen \
# 		-generate spec \
# 		-package $(API_PKG) \
# 		-o $(GEN_DIR)/$(API_PKG)/spec.gen.go \
# 		$(SPEC_DIR)/api.yml

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