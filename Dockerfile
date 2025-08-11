# stage build
FROM golang:1.24.3-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o marketflow ./cmd/marketstream

# stage runtime
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /src/marketflow /app/marketflow
EXPOSE 8080
ENTRYPOINT ["/app/marketflow"]
