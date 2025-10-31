FROM golang:1.25-alpine AS builder

# Installs system dependencies:
# - build-base: (gcc, make) required for compilation
# - protobuf-dev: The 'protoc' compiler
RUN apk add --no-cache build-base protobuf-dev

# Sets the working directory
WORKDIR /app

# Installs gRPC plugins for Go
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

# Adds Go's bin directory to PATH
ENV PATH="${PATH}:/go/bin"

# Copies .proto files first
COPY proto/ ./proto/

# Copies Go module files and downloads dependencies
COPY server/go.mod server/go.sum ./server/
RUN cd server && go mod download

# Copies the rest of the Go server source code
COPY server/ ./server/

# Generates gRPC code from .proto
# (Running from inside the server directory)
RUN cd server && mkdir pb && protoc -I=../proto \
    --go_out=./pb --go_opt=paths=source_relative \
    --go-grpc_out=./pb --go-grpc_opt=paths=source_relative \
    ../proto/chat.proto

# Builds the Go server statically (without CGO)
# This is crucial for running it in the final 'alpine' image
RUN cd server && CGO_ENABLED=0 GOOS=linux go build -o /server_binary ./main.go

# STAGE 2: The Final (Minimal) Image
FROM alpine:latest

# Copies ONLY the compiled binary from the 'builder' stage
COPY --from=builder /server_binary /server

# Exposes the port used by the gRPC server
EXPOSE 50051

# Command to run the server
CMD ["/server"]