ARG GO_VERSION=1.27

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /ghostdrop \
    ./cmd/ghostdrop

FROM alpine:3.24

RUN apk add --no-cache ca-certificates \
    && addgroup -S ghostdrop \
    && adduser -S -G ghostdrop ghostdrop \
    && mkdir -p /data \
    && chown ghostdrop:ghostdrop /data

WORKDIR /app

COPY --from=builder /ghostdrop /app/ghostdrop

ENV GHOSTDROP_WEB_HOST=0.0.0.0
ENV GHOSTDROP_DISABLE_BROWSER=true

EXPOSE 8080

VOLUME ["/data"]

USER ghostdrop

ENTRYPOINT ["/app/ghostdrop", "/data/config.yaml"]
