# syntax=docker/dockerfile:1.5
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS deps
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

FROM deps AS build
COPY . .
RUN apk add --no-cache \
    ca-certificates=20250619-r0 \
    tzdata=2025b-r0 \
    && update-ca-certificates
ARG TARGETOS TARGETARCH VERSION
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -buildvcs=false -ldflags "-s -w -X main.Version=${VERSION:-dev}" -o /bin/bot ./cmd/bot

FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /bin/bot ./bot
COPY --from=build /src/db /db
COPY --from=build /src/translations /translations
ENV DISABLE_ENV_FILE=true
ENTRYPOINT ["/app/bot"]
