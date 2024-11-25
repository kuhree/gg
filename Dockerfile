ARG APP_ENV="production"

FROM golang:1.23.3-alpine AS deps
ENV GOOS=linux
ENV CGO_ENABLED=1
WORKDIR /app

COPY go.mod go.sum ./
RUN apk update && apk upgrade \ 
  && apk add --no-cache git ca-certificates tzdata build-base curl \
  && update-ca-certificates \
  && go mod download

FROM deps AS builder
WORKDIR /app

COPY . .
RUN go build \
  -ldflags '-w -s -extldflags "-static"' \
  -a -o ./entrypoint \
  ./cmd/gg/main.go

FROM alpine:latest AS runner
ENV USER=gg
ENV UID=10001
WORKDIR /app

COPY --from=deps /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=deps /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=deps /etc/passwd /etc/passwd   
COPY --from=deps /etc/group /etc/group
COPY --from=builder /app/entrypoint .

RUN adduser \
  --disabled-password --gecos "" --home "/nonexistent" --shell "/sbin/nologin" \
  --no-create-home --uid "${UID}" "${USER}" \
  && chown -R ${USER}:${USER} /app 

USER gg
CMD ["./entrypoint"]
