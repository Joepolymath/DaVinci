
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

run-weaviate:
	@docker run -p 8080:8080 -p 50051:50051 cr.weaviate.io/semitechnologies/weaviate:1.35.7