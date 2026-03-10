FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /workspace

COPY libs/go-common/ libs/go-common/
COPY services/ai-discovery-worker/ services/ai-discovery-worker/

WORKDIR /workspace/services/ai-discovery-worker

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /ai-discovery-worker .

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

COPY --from=builder /ai-discovery-worker /ai-discovery-worker
COPY --from=builder /workspace/services/ai-discovery-worker/prompts /prompts

RUN chown -R appuser:appgroup /ai-discovery-worker /prompts

USER appuser

ENTRYPOINT ["/ai-discovery-worker"]
