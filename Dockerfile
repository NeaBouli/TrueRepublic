# ============================================================
# Stage 1: Build the truerepublicd binary
# ============================================================
FROM golang:1.26.5-bookworm AS builder

RUN apt-get update \
    && apt-get install --yes --no-install-recommends build-essential git \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo dev)" \
    -o /usr/local/bin/truerepublicd \
    ./
RUN cp "$(go list -m -f '{{.Dir}}' github.com/CosmWasm/wasmvm/v2)/internal/api/libwasmvm.$(uname -m).so" /usr/local/lib/

# ============================================================
# Stage 2: Minimal runtime image
# ============================================================
FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install --yes --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/local/bin/truerepublicd /usr/local/bin/truerepublicd
COPY --from=builder /usr/local/lib/libwasmvm.*.so /usr/lib/
RUN ldconfig

EXPOSE 26656 26657 1317 9090

VOLUME ["/root/.truerepublic"]

ENTRYPOINT ["truerepublicd"]
CMD ["start"]
