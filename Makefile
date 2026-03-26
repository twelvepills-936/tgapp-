ifeq ($(OS),Windows_NT)
CUR_DIR=$(shell echo %CD%)
else
CUR_DIR=$(shell pwd)
endif

# генерация pg.go файлов и swagger
proto.gen:
	docker run --rm \
		-v ${CUR_DIR}:/workspace \
		-w /workspace \
		bufbuild/buf:1.57.0 generate


# proto.deps.update создает buf.lock
proto.deps.update:
	docker run --rm \
		-v ${CUR_DIR}:/workspace \
		-w /workspace \
		bufbuild/buf:1.57.0 dep update

lint:
	docker run --rm -v `pwd`:/app -w /app golangci/golangci-lint:v2.4.0 golangci-lint run --timeout 5m0s -v