FROM golang:1.23-alpine AS builder

WORKDIR /build

COPY shared/ ./shared/
COPY services/worker/ ./services/worker/

RUN cd services/worker && go mod download
RUN cd services/worker && \
    CGO_ENABLED=0 GOOS=linux go build -o /build/worker server/cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /build/worker .

CMD ["./worker"]