GO ?= go
.PHONY: install dev backend frontend check build
install:
	cd frontend && npm ci
dev:
	GO="$(GO)" ./scripts/dev.sh
backend:
	cd backend && DASHBOARD_DEV=1 $(GO) run ./cmd/server
frontend:
	cd frontend && npm run dev
check:
	cd backend && $(GO) test -race ./...
	cd backend && $(GO) vet ./...
	cd frontend && npm run check && npm test
build:
	cd backend && $(GO) build -o /tmp/apartment-display-server ./cmd/server
	cd frontend && npm run build
