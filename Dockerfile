# syntax=docker/dockerfile:1

# Which binary to build. Both services live in one module and share the platform
# packages, so one Dockerfile with an argument beats two that drift apart.
#
# A path rather than a name: the services own their subtree (onboarding/cmd,
# demo/cmd) while the migrator stays at the module root, so there is no single
# prefix that would cover all three.
ARG SERVICE_PATH=./onboarding/cmd

# ---------------------------------------------------------------- build
FROM golang:1.26-alpine AS build

WORKDIR /src

# Dependencies are copied on their own so the module download layer is reused
# whenever only source files change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE_PATH

# How many packages the compiler builds at once. Empty by default, which Go
# reads as "one per core" — the fast path on a machine with memory to spare.
# Set it to 1 or 2 on a small VPS: each parallel compile holds its own working
# set, and it is that concurrency, not the build itself, that runs the box out
# of memory.
#
#   docker compose build --build-arg GOMAXPROCS=1 onboarding
ARG GOMAXPROCS
ENV GOMAXPROCS=${GOMAXPROCS}

# CGO off gives a static binary, which is what makes the scratch-like runtime
# stage below possible.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /out/service ${SERVICE_PATH}

# ---------------------------------------------------------------- runtime
FROM alpine:3.22

# wget is used by the compose healthcheck; ca-certificates keeps outbound TLS
# working if the service ever has to call anything.
RUN apk add --no-cache ca-certificates wget \
    && adduser -D -u 10001 cluer

COPY --from=build /out/service /usr/local/bin/service

# Runs unprivileged: nothing here needs root, and a container that does not need
# it should not have it.
USER cluer

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/service"]
