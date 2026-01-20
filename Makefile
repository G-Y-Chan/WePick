# Makefile at: root/Makefile

BACKEND_DIR := backend
FRONTEND_DIR := frontend
BACKEND_MAIN := main.go

.PHONY: dev backend frontend backend-test frontend-test

## Run only the Go backend
backend:
	cd $(BACKEND_DIR) && go run .

## Run only the React Native / Expo app
frontend:
	cd $(FRONTEND_DIR) && npm run start

## Run backend + frontend together (dev mode)
dev:
	cd $(BACKEND_DIR) && go run $(BACKEND_MAIN) &
	cd $(FRONTEND_DIR) && npm run start

