# ============================================================
# Stage 1: Build the truerepublicd binary
# ============================================================
FROM golang:1.26.6-bookworm@sha256:116d58cbd88c1297624acc6e967a060012422bacf9930927e23fb719189c6f36 AS builder

ARG TARGETARCH
ARG VERSION=dev

RUN apt-get update \
    && apt-get install --yes --no-install-recommends build-essential \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOFLAGS=-mod=readonly go build \
    -trimpath \
    -buildvcs=false \
    -ldflags="-s -w -buildid= -X main.version=${VERSION} -X main.upgradePlan=v0.4.1 -linkmode=external -extldflags=-Wl,--build-id=none" \
    -o /usr/local/bin/truerepublicd \
    ./
RUN set -eux; \
    case "${TARGETARCH:-$(dpkg --print-architecture)}" in \
        amd64) wasmvm_arch=x86_64 ;; \
        arm64) wasmvm_arch=aarch64 ;; \
        *) echo "unsupported target architecture: ${TARGETARCH:-unknown}" >&2; exit 1 ;; \
    esac; \
    wasmvm_dir="$(go list -m -f '{{.Dir}}' github.com/CosmWasm/wasmvm/v2)"; \
    install -m 0755 "${wasmvm_dir}/internal/api/libwasmvm.${wasmvm_arch}.so" /usr/local/lib/

# ============================================================
# Stage 2: Minimal runtime image
# ============================================================
FROM debian:bookworm-slim@sha256:abd67ffcfa541b485a3dff59865ab629aa048a6c613e639d36e7456b0b229241

RUN apt-get update \
    && apt-get install --yes --no-install-recommends ca-certificates libgcc-s1 wget \
    && rm -rf /var/lib/apt/lists/* /var/cache/apt/* /var/log/apt \
    && rm -f /var/log/dpkg.log /var/log/alternatives.log /var/cache/ldconfig/aux-cache \
    && groupadd --system truerepublic \
    && useradd --system --gid truerepublic --home-dir /home/truerepublic --create-home truerepublic \
    && sed -i -E '/^truerepublic:/ s/^([^:]*:[^:]*):[0-9]+:/\1::/' /etc/shadow \
    && rm -f /etc/passwd- /etc/group- /etc/shadow- /etc/gshadow- /etc/subuid- /etc/subgid- \
        /var/log/faillog /var/log/lastlog /var/log/wtmp /var/log/btmp \
    && mkdir -p /home/truerepublic/.truerepublic \
    && chown -R truerepublic:truerepublic /home/truerepublic

COPY --from=builder /usr/local/bin/truerepublicd /usr/local/bin/truerepublicd
COPY --from=builder /usr/local/lib/libwasmvm.*.so /usr/lib/
COPY --chmod=755 scripts/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN ldconfig \
    && rm -f /var/cache/ldconfig/aux-cache \
    && truerepublicd --help >/dev/null \
    && truerepublicd --version >/dev/null

USER truerepublic
ENV HOME=/home/truerepublic

EXPOSE 26656

VOLUME ["/home/truerepublic/.truerepublic"]

HEALTHCHECK --interval=15s --timeout=3s --start-period=20s --retries=5 \
  CMD ["truerepublicd", "healthcheck", "live", "--timeout", "2s"]

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["start", "--log_format=json", "--rpc.laddr=tcp://127.0.0.1:26657", "--grpc.enable=false", "--api.enable=false", "--minimum-gas-prices=1000upnyx"]
