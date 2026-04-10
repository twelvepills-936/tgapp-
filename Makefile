ifeq ($(OS),Windows_NT)
CUR_DIR=$(shell echo %CD%)
else
CUR_DIR=$(shell pwd)
endif

# Generate gRPC / Swagger files from proto
proto.gen:
	docker run --rm \
		-v ${CUR_DIR}:/workspace \
		-w /workspace \
		bufbuild/buf:1.57.0 generate

# Update proto dependencies (regenerates buf.lock)
proto.deps.update:
	docker run --rm \
		-v ${CUR_DIR}:/workspace \
		-w /workspace \
		bufbuild/buf:1.57.0 dep update

# Run linter
lint:
	docker run --rm -v "$(CUR_DIR)":/app -w /app golangci/golangci-lint:v2.4.0 golangci-lint run --timeout 5m0s -v

# Run all tests
test:
	go test ./...

# Run with coverage
test.cover:
	go test ./... -cover

# Format all Go files
fmt:
	gofmt -w .

# Start Postgres and apply migrations
db.up:
	docker compose up -d postgres
	docker compose run --rm migrate

# Run migrations manually
db.migrate:
	docker compose run --rm migrate

# Stop and wipe all containers + volumes
db.down:
	docker compose down -v