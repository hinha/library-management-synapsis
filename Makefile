# Directories
PROTO_DIR = api/proto
OUT_DIR = gen
SWAGGER_OUT_DIR = swagger

# Find all .proto files
PROTOS = $(shell find $(PROTO_DIR) -name "*.proto")

# Default target
all: generate gotag swagger

# Install required plugins
install-plugins:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install github.com/danclive/protoc-gen-go-tag@latest

# Step 1: Generate core code using buf (excluding gotag)
generate:
	buf generate
	./gotag.sh

# Lint proto files using buf
lint:
	buf lint

# Check for breaking changes
breaking:
	buf breaking --against '.git#branch=main'

# Clean all generated files
clean:
	rm -rf $(OUT_DIR)
	rm -rf $(SWAGGER_OUT_DIR)

.PHONY: all install-plugins generate gotag swagger lint breaking clean