FROM golang:1.23-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /immichmcp ./cmd/immichmcp

FROM alpine:3.20
RUN apk add --no-cache ca-certificates curl

WORKDIR /app
COPY --from=build /immichmcp .

ENV MCP_PORT=5000

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:${MCP_PORT}/health || exit 1

EXPOSE 5000

ENTRYPOINT ["/app/immichmcp"]
