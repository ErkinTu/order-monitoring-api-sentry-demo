FROM golang:1.23-alpine AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/api

FROM alpine:3.20

RUN adduser -D appuser
WORKDIR /app

COPY --from=builder /app/server ./server

EXPOSE 8080

USER appuser
CMD ["./server"]
