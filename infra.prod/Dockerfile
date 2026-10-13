# Build context: repo root (docker build -f infra.prod/Dockerfile .)
# On Railway: RAILWAY_DOCKERFILE_PATH=infra.prod/Dockerfile
# Migrations (e.g. as the Pre-Deploy Command on Railway): /app/migrate up

FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -H -u 10001 app
WORKDIR /app
COPY --from=build /out/api /app/api
COPY --from=build /out/migrate /app/migrate
COPY migrations ./migrations

USER app
EXPOSE 4989

# Railway injects PORT; if it is not set, LISTEN_ADDR or :4989 is used.
CMD ["sh", "-c", "if [ -n \"$PORT\" ]; then export LISTEN_ADDR=\":$PORT\"; fi; exec /app/api"]
