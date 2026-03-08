FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service ./cmd/auth-service

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
COPY --from=builder /auth-service /auth-service
EXPOSE 50051
ENTRYPOINT ["/auth-service"]
