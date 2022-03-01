FROM golang:alpine AS builder

WORKDIR /app

# copy just the go mod file and download mods so we can hopefully get some docker layer caching benefits
COPY --chown=1000:1000 ./go.* /app/
RUN ["go", "mod", "download"]

COPY --chown=1000:1000 ./ /app/
RUN ["go", "build", "./cmd/api"]
# RUN ["find", "."]

# we need edge because we built packages for edge
FROM alpine:edge

LABEL maintainer="iggy@atlascloud.xyz"
LABEL org.opencontainers.image.source=https://github.com/atlascloud/packages
EXPOSE 8888

RUN ["apk", "add", "abuild"]

COPY ./api /app/

USER 1000:1000

CMD ["/app/api"]
