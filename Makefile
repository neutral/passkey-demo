
.PHONY: dev server node-server web build clean run test test-ui test-all

dev:
	@echo "Open two terminals:"
	@echo "  1) make server"
	@echo "  2) make web"

server:
	npm run dev --prefix node-server

node-server: server

web:
	npm run dev --prefix web

build:
	npm run build --prefix web

clean:
	rm -f node-server/*.db
	rm -rf web/dist

# Start server and web concurrently; Ctrl-C stops both
run:
	@echo "Starting Node server and web (Ctrl-C to stop both)...";
	@set -e; \
	( cd node-server && RP_ID=$${RP_ID:-localhost} ORIGIN=$${ORIGIN:-http://localhost:5173} PORT=$${PORT:-8080} DB_PATH=$${DB_PATH:-demo.db} npm run dev ) & \
	SERVER_PID=$$!; \
	( cd web && npm run dev ) & \
	WEB_PID=$$!; \
	trap 'kill $$SERVER_PID $$WEB_PID 2>/dev/null || true' INT TERM; \
	wait

test:
	npm run test --prefix node-server

test-ui:
	npm run test:ui --prefix web

test-all:
	$(MAKE) test
	$(MAKE) test-ui
