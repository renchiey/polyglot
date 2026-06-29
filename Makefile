DB_URL ?= postgres://app:app@localhost:5432/app?sslmode=disable

.PHONY: help db-up db-down migrate-up migrate-down migrate-new sqlc server server-build mobile mobile-web web web-install web-build tts-setup tidy

PIPER_VOICE_NAME ?= zh_CN-huayan-medium

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-14s\033[0m %s\n",$$1,$$2}'

db-up: ## Start Postgres in Docker
	docker compose up -d db

db-down: ## Stop and remove containers
	docker compose down

migrate-up: ## Apply all up migrations
	migrate -path server/internal/db/migrations -database "$(DB_URL)" up

migrate-down: ## Roll back the last migration
	migrate -path server/internal/db/migrations -database "$(DB_URL)" down 1

migrate-new: ## Create a migration: make migrate-new name=add_widgets
	migrate create -ext sql -dir server/internal/db/migrations -seq $(name)

sqlc: ## Generate typed DB code from SQL
	cd server && sqlc generate

server: ## Run the API
	cd server && go run ./cmd/api

server-build: ## Build the API binary into server/bin
	cd server && go build -o bin/api ./cmd/api

tts-setup: ## Install Piper and download the Mandarin voice (optional; enables server TTS)
	pip install piper-tts
	python3 -m piper.download_voices $(PIPER_VOICE_NAME) --download-dir server/voices
	@echo "Voice installed in server/voices — restart the server to enable TTS."

tidy: ## go mod tidy
	cd server && go mod tidy

mobile: ## Start Expo (choose platform interactively)
	cd mobile && npm run start

mobile-web: ## Start Expo on web
	cd mobile && npm run web

web-install: ## Install web client dependencies
	cd web && npm install

web: ## Start the React web client on :8081 (claymorphism UI)
	cd web && npm run dev

web-build: ## Type-check and build the web client
	cd web && npm run build
