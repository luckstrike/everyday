GO ?= go
.PHONY: install dev backend frontend check build
install:
	cd frontend && npm ci
dev:
	GO="$(GO)" ./scripts/dev.sh
backend:
	cd backend && $(GO) run ./cmd/server
frontend:
	cd frontend && npm run dev
check:
	cd backend && $(GO) test -race ./...
	cd backend && $(GO) vet ./...
	cd frontend && npm run check && npm test
build:
	cd backend && $(GO) build -o /tmp/everyday-server ./cmd/server
	cd frontend && npm run build
