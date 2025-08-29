.PHONY: dev server web build clean

dev:
	@echo "Open two terminals:"
	@echo "  1) make server"
	@echo "  2) make web"

server:
	cd server && go run ./cmd/api

web:
	cd web && npm run dev

build:
	cd server && go build ./cmd/api
	cd web && npm run build

clean:
	rm -f server/*.db
	rm -rf web/dist
