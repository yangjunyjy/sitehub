FROM golang:1.24-bookworm AS build
ARG TARGET=service/rbac/rpc
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -trimpath -o /out/server "./${TARGET}" && mkdir -p /out/etc && if [ -d "${TARGET}/etc" ]; then cp -r "${TARGET}/etc/." /out/etc/; fi

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata && rm -rf /var/lib/apt/lists/* && useradd -u 10001 -m app && mkdir -p /app /data/files /data/releases && chown -R app:app /app /data
WORKDIR /app
COPY --from=build /out/server /app/server
COPY --from=build /out/etc /app/etc
USER app
ENTRYPOINT ["/app/server"]
