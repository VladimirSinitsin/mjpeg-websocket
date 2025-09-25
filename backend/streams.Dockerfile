FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .

ENV CGO_ENABLED=0

RUN go build -ldflags "-s -w -X main.Version=1.0.1" -mod=vendor -o ./server ./cmd/streams/main.go ./cmd/streams/app.go

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
RUN addgroup -g 2020 appgroup && adduser --uid 1020 -H -D -G appgroup appuser
COPY --from=builder /app/server    /app/server
RUN chown appuser:appgroup /app/server && chmod 700 /app/server

USER appuser
CMD ["/app/server"]
