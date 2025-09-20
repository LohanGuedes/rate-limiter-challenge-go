FROM golang:1.24.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o notification-service ./notification/cmd/main.go

RUN ls -la ./notification-service

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy from alias
COPY --from=builder /app/notification-service ./notification-service

RUN chmod +x ./notification-service

RUN ls -la ./notification-service

EXPOSE 8080

CMD ["./notification-service"]
