# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .

# Build the binary
RUN go build -o data_digger ./cmd/digger

# Runtime Stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/data_digger .
COPY --from=builder /app/cmd/digger/sources ./sources

# Entrypoint to run the tool
ENTRYPOINT ["./data_digger"]
CMD ["--help"]
