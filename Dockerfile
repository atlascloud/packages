FROM --platform=$BUILDPLATFORM golang:alpine AS builder

WORKDIR /app

# go build needs git to be installed
RUN apk add --no-cache git

COPY ./ /app/
ARG TARGETARCH
ARG TARGETOS
RUN ["go", "mod", "download"]
RUN ["go", "mod", "tidy"]
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" ./cmd/api

# we need edge because we built packages for edge
FROM alpine:edge

LABEL maintainer="packages@atlascloud.xyz"
LABEL org.opencontainers.image.source=https://github.com/atlascloud/packages
LABEL org.opencontainers.image.description="API for managing packages service"

EXPOSE 8888

HEALTHCHECK --interval=5m --timeout=10s \
    CMD curl -fs localhost:8888/health/ready

COPY --from=builder /app/api /app/

USER 1000:1000

CMD ["/app/api"]
