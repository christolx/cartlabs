FROM golang:1.26-alpine AS build

ARG APP
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY apps ./apps
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/app ./apps/${APP}

FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
    && addgroup -S cartlabs \
    && adduser -S -G cartlabs cartlabs

WORKDIR /app
COPY --from=build /out/app /usr/local/bin/app
COPY --chown=cartlabs:cartlabs migrations ./migrations
COPY --chown=cartlabs:cartlabs seeds ./seeds

USER cartlabs
ENTRYPOINT ["/usr/local/bin/app"]
