FROM golang:alpine AS builder

WORKDIR /app

# go build needs git to be installed
RUN apk add --no-cache git

# copy just the go mod file and download mods so we can hopefully get some docker layer caching benefits
COPY ./go.* /app/
RUN ["go", "mod", "download"]

COPY ./ /app/
ENV CGO_ENABLED=0
RUN ["go", "build", "-ldflags=\"-extldflags=-static\"", "./cmd/api/"]
# RUN ["find", "."]

# we need edge because we built packages for edge
FROM alpine:edge

LABEL maintainer="packages@atlascloud.xyz"
LABEL org.opencontainers.image.source=https://github.com/atlascloud/packages
LABEL org.opencontainers.image.description="API for managing packages service"

EXPOSE 8888

HEALTHCHECK --interval=5m --timeout=10s \
    CMD curl -fs localhost:8888/health/ready

# the service internally calls abuild to create the package index
RUN ["apk", "add", "--no-cache", "abuild"]

COPY --from=builder /app/api /app/

USER 1000:1000

CMD ["/app/api"]
