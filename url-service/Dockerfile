FROM golang:1.25-alpine AS builder

ENV GOTOOLCHAIN=local
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/url-service .

FROM alpine:3.21

RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=builder /app/url-service .

EXPOSE 8080

ENTRYPOINT ["./url-service"]