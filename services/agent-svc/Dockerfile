# Use a pre-built binary to avoid Alpine network issues during build
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata || true

WORKDIR /app

# Copy pre-built binary
COPY agent-svc /app/agent-svc
COPY agent-svc.yaml /app/agent-svc.yaml

RUN chmod +x /app/agent-svc

EXPOSE 8082

CMD ["/app/agent-svc", "-f", "agent-svc.yaml"]
