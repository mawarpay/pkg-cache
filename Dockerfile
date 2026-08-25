# syntax=docker/dockerfile:1
# Test runner image (not a production service binary).
FROM golang:1.26-bookworm

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .

CMD ["go", "test", "./...", "-count=1"]
