FROM golang:1.23-alpine AS builder

WORKDIR /build

COPY shared/ ./shared/
COPY services/app/ ./services/app/

RUN cd services/app && go mod download
RUN cd services/app && \
    CGO_ENABLED=0 GOOS=linux go build -o /build/app cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /build/app .
COPY GeoLite2-City.mmdb .

EXPOSE 8080
CMD ["./app"]