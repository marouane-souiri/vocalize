ENV_FILE := .env
EXPORT_ENV := export $(shell grep -v '^#' $(ENV_FILE) | xargs)
MAIN_FILE := cmd/app/main.go

run:
	@$(EXPORT_ENV) && go run $(MAIN_FILE)

build:
	 go build $(MAIN_FILE)

test:
	go test ./...

fmt:
	go fmt ./...

.PHONY: run build test fmt
