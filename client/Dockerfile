# STAGE 1: Dependency Builder
FROM node:24-alpine AS deps

WORKDIR /app

# Copy only package files and install dependencies
COPY client/package.json client/package-lock.json ./client/
RUN cd client && npm install

# STAGE 2: Final Image
FROM node:24-alpine

WORKDIR /app

# Copy installed dependencies from 'deps' stage
COPY --from=deps /app/client/node_modules ./client/node_modules/

# Copy client code
COPY client/ ./client/

# IMPORTANT: Copy .proto files, as Node.js reads them at runtime
COPY proto/ ./proto/

# Entry point. 'CMD' will be passed by 'docker-compose run'
ENTRYPOINT ["node", "./client/client.js"]