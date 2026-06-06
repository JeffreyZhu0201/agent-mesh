# Use pre-built binary to avoid Alpine network issues during build
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata || true

WORKDIR /app

# Copy pre-built binary + config
COPY plugin-svc /app/plugin-svc
COPY plugin-svc.yaml /app/plugin-svc.yaml

RUN chmod +x /app/plugin-svc

EXPOSE 8083

CMD ["/app/plugin-svc", "-f", "plugin-svc.yaml"]
