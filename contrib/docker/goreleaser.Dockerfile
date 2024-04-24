FROM golang:1.20 AS builder

WORKDIR /root
COPY ./dist/ /root/

ARG TARGETARCH
RUN if [ "${TARGETARCH}" = "arm64" ]; then \
  cp linux_linux_arm64/picad /root/picad; \
  else \
  cp linux_linux_amd64_v1/picad /root/picad; \
  fi

FROM alpine:latest

RUN apk --no-cache add ca-certificates jq
COPY --from=builder /root/picad /usr/local/bin/picad

RUN addgroup --gid 1025 -S composable && adduser --uid 1025 -S composable -G composable

WORKDIR /home/composable
USER composable

# rest server
EXPOSE 1317
# tendermint p2p
EXPOSE 26656
# tendermint rpc
EXPOSE 26657
# grpc
EXPOSE 9090

ENTRYPOINT ["picad"]
CMD [ "start" ]
