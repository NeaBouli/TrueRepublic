# ============================================================
# Stage 1: Build the truerepublicd binary
# ============================================================
ARG GO_VERSION=1.26.5
FROM golang:${GO_VERSION}-bookworm AS builder

ARG TARGETARCH
ARG VERSION=dev

RUN apt-get update \
    && apt-get install --yes --no-install-recommends build-essential \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
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
FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install --yes --no-install-recommends ca-certificates libgcc-s1 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/local/bin/truerepublicd /usr/local/bin/truerepublicd
COPY --from=builder /usr/local/lib/libwasmvm.*.so /usr/lib/
RUN ldconfig \
    && truerepublicd --help >/dev/null

EXPOSE 26656 26657 1317 9090

VOLUME ["/root/.truerepublic"]

ENTRYPOINT ["truerepublicd"]
CMD ["start"]
