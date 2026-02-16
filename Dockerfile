# ============================================================
# Stage 1: Build the truerepublicd binary
# ============================================================
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache make gcc musl-dev linux-headers git

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo dev)" \
    -o /usr/local/bin/truerepublicd \
    ./

# ============================================================
# Stage 2: Minimal runtime image
# ============================================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates libstdc++ libc6-compat

COPY --from=builder /usr/local/bin/truerepublicd /usr/local/bin/truerepublicd

EXPOSE 26656 26657 1317 9090

VOLUME ["/root/.truerepublic"]

ENTRYPOINT ["truerepublicd"]
CMD ["start"]
