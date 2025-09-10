FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . . 
RUN go build -o api

# Runtime stage (copy certifications and binary)
FROM alpine:3.20
WORKDIR /app

# Copy CA certificates and binary
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/api .

EXPOSE 8080
CMD ["./api"]