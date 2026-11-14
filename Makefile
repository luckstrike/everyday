GO ?= go
.PHONY: install dev run backend frontend check build google-auth
install:
	cd frontend && npm ci
dev:
	GO="$(GO)" ./scripts/dev.sh
run: build
	exec ./build/everyday -web-dir ./frontend/dist
backend:
	cd backend && $(GO) run ./cmd/server
frontend:
	cd frontend && npm run dev
check:
	cd backend && $(GO) test -race ./...
	cd backend && $(GO) vet ./...
	cd frontend && npm run check && npm test
build:
	cd frontend && npm run build
	mkdir -p build
	cd backend && $(GO) build -o ../build/everyday ./cmd/server
# Sign in to Google Calendar once. Example: make google-auth ARGS="-client-file ~/Downloads/client.json"
google-auth:
	cd backend && $(GO) run ./cmd/googleauth $(ARGS)
