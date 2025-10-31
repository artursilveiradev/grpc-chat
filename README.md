# gRPC Chat

This project demonstrates the use of [gRPC with bidirectional streams](https://grpc.io/docs/what-is-grpc/core-concepts/#bidirectional-streaming-rpc). The application is a simple chat where clients connect to a server to exchange messages. The gRPC server is written in [Golang](https://go.dev/) and the gRPC client is written in [Node.js](https://nodejs.org/).

## Running

The easiest way to run the project is with [Docker](https://www.docker.com/). Follow the step-by-step instructions below.

### 1. Build the images

Build the Docker images with the following command:

```
docker compose build
```

### 2. Start the server

Start the server with the following command:

```
docker compose up -d server
```

You can view the logs with:

```
docker logs -f grpc-server
```

### 3. Start the client

Start the client with the following command, replacing `<username>` with any username:

```
docker compose run --rm client <username>
```

To simulate additional clients, simply run the command above in a new terminal window.