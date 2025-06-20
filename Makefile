# Makefile untuk build .proto ke Go, gRPC, gRPC-Gateway, dan OpenAPI using buf.build

PROTO_DIR = api/proto
OUT_DIR = gen
SWAGGER_OUT_DIR = swagger

# Legacy protoc commands (kept for reference)
# PROTOC = protoc
# PROTOC_GEN_GO = --go_out=$(OUT_DIR) --go_opt=paths=source_relative
# PROTOC_GEN_GRPC = --go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative
# PROTOC_GEN_GATEWAY = --grpc-gateway_out=$(OUT_DIR) --grpc-gateway_opt=paths=source_relative
# PROTOC_GEN_OPENAPI = --openapiv2_out=$(OUT_DIR) --openapiv2_opt=logtostderr=true

PROTOS = $(shell find $(PROTO_DIR) -name "*.proto")

all: generate swagger

# Install required plugins
install-plugins:
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# Generate code using buf
generate:
	buf generate

# Lint proto files using buf
lint:
	buf lint

# Check for breaking changes
breaking:
	buf breaking --against '.git#branch=main'

# Clean generated files
clean:
	rm -rf $(OUT_DIR)
	rm -rf $(SWAGGER_OUT_DIR)

.PHONY: all install-plugins generate swagger lint breaking clean
