FROM golang:1.24-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/gateway ./gateway

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata && rm -rf /var/lib/apt/lists/* && useradd -u 10001 -m app && mkdir -p /app/etc /data/nacos && chown -R app:app /app /data
WORKDIR /app
COPY --from=build /out/gateway /app/gateway
COPY deploy/gateway/gateway.yaml /app/etc/gateway.yaml
COPY gateway/etc/routes.json /app/etc/routes.json
USER app
ENTRYPOINT ["/app/gateway"]
CMD ["-f", "/app/etc/gateway.yaml"]
