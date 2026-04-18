.PHONY: dev backend frontend redis ngrok

# Makefile at: root/Makefile
ROOT := $(CURDIR)
BACKEND_DIR := backend
FRONTEND_DIR := frontend
BACKEND_MAIN := main.go

## Run only the Go backend
backend:
	cd $(BACKEND_DIR) && go run .

## Run only the React Native / Expo app
frontend:
	cd $(FRONTEND_DIR) && npm run start

## Run only the Redis server
redis:
	redis-server

## Run only the ngrok tunnel
ngrok:
	ngrok http 8090

## Run backend + frontend together + ngrok tunnel + Redis server
dev:
	@osascript \
		-e "tell application \"Terminal\" to activate" \
		-e "tell application \"Terminal\" to do script \"cd '$(ROOT)' && redis-server\"" \
		-e "tell application \"Terminal\" to do script \"for i in {1..30}; do redis-cli ping >/dev/null 2>&1 && break; sleep 1; done; cd '$(ROOT)/backend' && go run .\"" \
		-e "tell application \"Terminal\" to do script \"cd '$(ROOT)/frontend' && npm run start\"" \
		-e "tell application \"Terminal\" to do script \"cd '$(ROOT)' && ngrok http 8090\""

