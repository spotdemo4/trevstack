## BUF BUILD
FROM bufbuild/buf:1.50.0 AS buf
WORKDIR /buf

# Create client and server services
COPY buf.yaml buf.gen.yaml base.openapi.yaml ./
COPY proto ./proto
RUN buf generate


## CLIENT BUILD
FROM node:22-alpine AS client
WORKDIR /client

# Install client dependencies
COPY client/package.json client/package-lock.json ./
RUN npm ci

# Get client source
COPY client .

# Get buf service
COPY --from=buf /buf/client/src/lib/services ./src/lib/services

# Build client
RUN npm run build


## SERVER BUILD
FROM golang:1.23 AS server
WORKDIR /server

# Install server dependencies
COPY server/go.mod server/go.sum ./
RUN go mod download && go mod verify

# Get server source
COPY server .

# Get client build
COPY --from=client /client/build ./client

# Get buf service
COPY --from=buf /buf/server/internal/services ./internal/services

# Build server
RUN go build -v -o /server/main .

CMD ["/server/main"]