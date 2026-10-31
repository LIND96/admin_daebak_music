FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /migrate ./cmd/migrate

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /server /server
COPY --from=builder /migrate /migrate
COPY migrations ./migrations
EXPOSE 8080
CMD ["/bin/sh", "-c", "/migrate up && /server"]