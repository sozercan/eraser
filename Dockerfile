# syntax=mcr.microsoft.com/oss/moby/dockerfile:1.3.1

# Build the manager binary
FROM golang:1.17 AS builder

ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT=""

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    GOARM=${TARGETVARIANT}

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
ENV GOCACHE=/root/gocache
RUN \
    --mount=type=cache,target=${GOCACHE} \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY . .

FROM builder AS manager-build
RUN \
    --mount=type=cache,target=${GOCACHE} \
    --mount=type=cache,target=/go/pkg/mod \
    go build -ldflags '-w -extldflags "-static"' -o out/manager main.go

FROM builder AS eraser-build
RUN \
    --mount=type=cache,target=${GOCACHE} \
    --mount=type=cache,target=/go/pkg/mod \
    go build -ldflags '-w -extldflags "-static"' -o out/eraser ./pkg/eraser

FROM gcr.io/distroless/static:latest as eraser
COPY --from=eraser-build /workspace/out/eraser /
ENTRYPOINT ["/eraser"]

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot AS manager
WORKDIR /
COPY --from=manager-build /workspace/out/manager .
USER 65532:65532
ENTRYPOINT ["/manager"]
