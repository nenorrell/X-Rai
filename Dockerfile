FROM golang:1.24-alpine AS build
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o xrai ./cmd/xrai

FROM alpine:3.21
WORKDIR /app
COPY --from=build /app/xrai /app/xrai
ENTRYPOINT ["/app/xrai"]
