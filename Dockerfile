FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLE=0 GOOS=linux go build -o /app/chess-server ./cmd/server/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/chess-server /app/chess-server

EXPOSE 7070
EXPOSE 7777
EXPOSE 8080
EXPOSE 8888
EXPOSE 9090
EXPOSE 9999


CMD ["/app/chess-server"]
