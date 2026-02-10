
BACKEND_DIR := backend
BACKEND_BIN := app

.PHONY: build-backend run-backend clean-backend

build-backend:
	@cd $(BACKEND_DIR) && go build -o $(BACKEND_BIN) ./cmd

run-backend:
	@cd $(BACKEND_DIR) && go run ./cmd

run-binary: build-backend
	@./$(BACKEND_DIR)/$(BACKEND_BIN)

clean-backend:
	@rm -f $(BACKEND_DIR)/$(BACKEND_BIN)