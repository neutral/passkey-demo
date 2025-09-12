.PHONY: dev server web build clean run test test-all

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

# Start server and web concurrently; Ctrl-C stops both
run:
	@echo "Starting server and web (Ctrl-C to stop both)...";
	@set -e; \
	( cd server && RP_ID=$${RP_ID:-localhost} ORIGIN=$${ORIGIN:-http://localhost:5173} PORT=$${PORT:-8080} DB_PATH=$${DB_PATH:-server/demo.db} go run ./cmd/api ) & \
	SERVER_PID=$$!; \
	( cd web && npm run dev ) & \
	WEB_PID=$$!; \
	trap 'kill $$SERVER_PID $$WEB_PID 2>/dev/null || true' INT TERM; \
	wait

# Go tests: short by default; filter with PKG and RUN (regex)
test:
	cd server && go test -short $${PKG:-./...} $${RUN:+-run $${RUN}} -v

# Full Go tests (no -short); filter with PKG and RUN (regex)
test-all:
	cd server && go test $${PKG:-./...} $${RUN:+-run $${RUN}} -v
