FROM golang:1.21-alpine

# Install build dependencies for SQLite
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Enable CGO for SQLite and build
RUN CGO_ENABLED=1 go build -o notification-service ./cmd/server

EXPOSE 8080

CMD ["./notification-service"]