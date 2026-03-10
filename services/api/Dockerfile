FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /workspace

# Copy module files first for layer caching
COPY libs/go-common/go.mod libs/go-common/go.sum libs/go-common/
COPY domain-packs/gaming/go/go.mod domain-packs/gaming/go/
COPY services/api/go.mod services/api/go.sum services/api/

WORKDIR /workspace/services/api
RUN go mod download

# Copy source
WORKDIR /workspace
COPY libs/ libs/
COPY domain-packs/ domain-packs/
COPY services/api/ services/api/

WORKDIR /workspace/services/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /decisionbox-api .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -g 1000 -S app && adduser -u 1000 -S app -G app

COPY --from=builder /decisionbox-api /usr/local/bin/decisionbox-api

EXPOSE 8080
USER app
ENTRYPOINT ["decisionbox-api"]
