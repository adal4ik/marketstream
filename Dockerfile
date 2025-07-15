FROM golang:1.24.3-alpine

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main ./cmd/marketstream

EXPOSE 8080

CMD ["./main"]
