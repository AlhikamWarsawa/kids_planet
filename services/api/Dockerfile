# build
FROM golang:1.25.5-alpine AS builder
WORKDIR /src
RUN apk add --no-cache ca-certificates git curl

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# dev
FROM golang:1.25.5-alpine AS dev
WORKDIR /app
RUN apk add --no-cache git curl
RUN go install github.com/air-verse/air@latest
COPY . .
EXPOSE 8080
ENV PORT=8080
CMD ["air", "-c", ".air.toml"]

# prod
FROM alpine:3.20 AS prod
WORKDIR /app
RUN apk add --no-cache ca-certificates && update-ca-certificates
COPY --from=builder /out/api /app/api
EXPOSE 8080
ENV PORT=8080
CMD ["/app/api"]