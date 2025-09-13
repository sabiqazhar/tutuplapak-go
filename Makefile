# Database Migration Makefile
# Requirements: golang-migrate tool
# Install with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Database configuration
DB_HOST ?= localhost
DB_PORT ?= 5433
DB_USER ?= user
DB_PASSWORD ?= password
DB_NAME ?= tutuplapak
DB_SSL_MODE ?= disable

# Migration settings
MIGRATIONS_DIR ?= ./migrations
DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

# Colors for output
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

.PHONY: help migrate-create migrate-up migrate-down migrate-drop migrate-force migrate-version migrate-status

help: ## Show this help message
	@echo "$(GREEN)Available commands:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(RESET) %s\n", $$1, $$2}'

migrate-create: ## Create a new migration file (usage: make migrate-create NAME=create_users_table)
	@if [ -z "$(NAME)" ]; then \
		echo "$(RED)Error: NAME is required$(RESET)"; \
		echo "Usage: make migrate-create NAME=create_users_table"; \
		exit 1; \
	fi
	@echo "$(GREEN)Creating migration: $(NAME)$(RESET)"
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)
	@echo "$(GREEN)Migration files created in $(MIGRATIONS_DIR)$(RESET)"

migrate-up: ## Run all pending migrations
	@echo "$(GREEN)Running migrations up...$(RESET)"
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up
	@echo "$(GREEN)Migrations completed successfully$(RESET)"

migrate-up-one: ## Run only the next pending migration
	@echo "$(GREEN)Running one migration up...$(RESET)"
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up 1
	@echo "$(GREEN)Migration completed successfully$(RESET)"

migrate-down: ## Rollback the last migration
	@echo "$(YELLOW)Rolling back last migration...$(RESET)"
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1
	@echo "$(GREEN)Migration rolled back successfully$(RESET)"

migrate-down-all: ## Rollback all migrations (WARNING: This will drop all tables!)
	@echo "$(RED)WARNING: This will rollback ALL migrations and may drop all tables!$(RESET)"
	@echo "$(YELLOW)Are you sure? This action cannot be undone. Type 'yes' to continue:$(RESET)"
	@read -r confirm && [ "$$confirm" = "yes" ] || (echo "Aborted." && exit 1)
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down
	@echo "$(GREEN)All migrations rolled back$(RESET)"

migrate-drop: ## Drop everything in database (WARNING: Destructive!)
	@echo "$(RED)WARNING: This will drop everything in the database!$(RESET)"
	@echo "$(YELLOW)Are you sure? This action cannot be undone. Type 'yes' to continue:$(RESET)"
	@read -r confirm && [ "$$confirm" = "yes" ] || (echo "Aborted." && exit 1)
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" drop
	@echo "$(GREEN)Database dropped$(RESET)"

migrate-force: ## Force set migration version (usage: make migrate-force VERSION=123)
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)Error: VERSION is required$(RESET)"; \
		echo "Usage: make migrate-force VERSION=123"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Forcing migration version to $(VERSION)...$(RESET)"
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $(VERSION)
	@echo "$(GREEN)Migration version forced to $(VERSION)$(RESET)"

migrate-version: ## Show current migration version
	@echo "$(GREEN)Current migration version:$(RESET)"
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

migrate-status: ## Show migration status
	@echo "$(GREEN)Migration status:$(RESET)"
	@echo "Database URL: $(DATABASE_URL)"
	@echo "Migrations directory: $(MIGRATIONS_DIR)"
	@if [ -d "$(MIGRATIONS_DIR)" ]; then \
		echo "Available migrations:"; \
		ls -la $(MIGRATIONS_DIR)/*.sql 2>/dev/null | wc -l | xargs -I {} echo "  {} migration files found"; \
	else \
		echo "$(YELLOW)Migrations directory does not exist$(RESET)"; \
	fi
	@make migrate-version

migrate-goto: ## Migrate to specific version (usage: make migrate-goto VERSION=123)
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)Error: VERSION is required$(RESET)"; \
		echo "Usage: make migrate-goto VERSION=123"; \
		exit 1; \
	fi
	@echo "$(GREEN)Migrating to version $(VERSION)...$(RESET)"
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" goto $(VERSION)
	@echo "$(GREEN)Migration to version $(VERSION) completed$(RESET)"

# Development helpers
dev-setup: ## Setup development environment (create migrations directory)
	@echo "$(GREEN)Setting up development environment...$(RESET)"
	@mkdir -p $(MIGRATIONS_DIR)
	@echo "$(GREEN)Created migrations directory: $(MIGRATIONS_DIR)$(RESET)"
	@echo "$(YELLOW)Make sure to install golang-migrate:$(RESET)"
	@echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"

dev-reset: ## Reset development database (drop + migrate up)
	@echo "$(YELLOW)Resetting development database...$(RESET)"
	@make migrate-drop
	@make migrate-up
	@echo "$(GREEN)Database reset completed$(RESET)"

# Example migration creation commands
example-migrations: ## Create example migration files
	@echo "$(GREEN)Creating example migrations...$(RESET)"
	@make migrate-create NAME=create_users_table
	@make migrate-create NAME=create_posts_table
	@make migrate-create NAME=add_index_to_users_email
	@echo "$(GREEN)Example migrations created$(RESET)"