# syntax=docker/dockerfile:1

ARG GO_VERSION="1.22"
ARG RUNNER_IMAGE="gcr.io/distroless/static-debian11"

# --------------------------------------------------------
# Builder
# --------------------------------------------------------

FROM golang:${GO_VERSION}-alpine3.18 as builder

ARG GIT_VERSION
ARG GIT_COMMIT

RUN apk add --no-cache \
    ca-certificates \
    build-base \
    linux-headers

# Download go dependencies
WORKDIR /pica
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    go mod download

# Cosmwasm - Download correct libwasmvm version
RUN set -eux; \    
    export ARCH=$(uname -m); \
    WASM_VERSION=$(go list -m all | grep github.com/CosmWasm/wasmvm | awk '{print $2}'); \
    if [ ! -z "${WASM_VERSION}" ]; then \
      wget -O /lib/libwasmvm_muslc.a https://github.com/CosmWasm/wasmvm/releases/download/${WASM_VERSION}/libwasmvm_muslc.${ARCH}.a; \      
    fi; \
    go mod download;
    
# Copy the remaining files
COPY . .

# Build picad binary
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    GOWORK=off go build \
        -mod=readonly \
        -tags "netgo,ledger,muslc" \
        -ldflags \
            "-X github.com/cosmos/cosmos-sdk/version.Name="pica" \
            -X github.com/cosmos/cosmos-sdk/version.AppName="picad" \
            -X github.com/cosmos/cosmos-sdk/version.Version=${GIT_VERSION} \
            -X github.com/cosmos/cosmos-sdk/version.Commit=${GIT_COMMIT} \
            -X github.com/cosmos/cosmos-sdk/version.BuildTags=netgo,ledger,muslc \
            -w -s -linkmode=external -extldflags '-Wl,-z,muldefs -static'" \
        -trimpath \
        -o /pica/build/picad \
        /pica/cmd/picad


# --------------------------------------------------------
# toolkit
# --------------------------------------------------------

FROM busybox:1.35.0-uclibc as busybox
RUN addgroup --gid 1025 -S pica && adduser --uid 1025 -S pica -G pica


# --------------------------------------------------------
# Runner
# --------------------------------------------------------
FROM ${RUNNER_IMAGE}

COPY --from=busybox:1.35.0-uclibc /bin/sh /bin/sh

COPY --from=builder /pica/build/picad /bin/picad

# Install composable user
COPY --from=busybox /etc/passwd /etc/passwd
COPY --from=busybox --chown=1025:1025 /home/pica /home/pica

WORKDIR /home/pica
USER pica

# rest server
EXPOSE 1317
# tendermint p2p
EXPOSE 26656
# tendermint rpc
EXPOSE 26657
# grpc
EXPOSE 9090

ENTRYPOINT ["picad"]