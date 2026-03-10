FROM golang:1.24-alpine AS builder

WORKDIR /build
COPY . .
RUN cd services/api && go build -o /bin/decisionbox-api .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/decisionbox-api /usr/local/bin/

EXPOSE 8080
USER 1000:1000
ENTRYPOINT ["decisionbox-api"]
