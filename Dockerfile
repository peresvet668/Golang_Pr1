FROM golang:1.27.3-alpine3.24 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/wallet-api ./cmd/api

FROM alpine:3.24

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/wallet-api ./wallet-api

USER app
EXPOSE 8080
ENTRYPOINT ["/app/wallet-api"]
