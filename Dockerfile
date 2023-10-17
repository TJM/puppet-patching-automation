FROM golang:1.20-alpine AS builder

ENV USER=appuser UID=10001

# We need make, git and ca-certificates
RUN mkdir /app && \
    touch /app/.env && \
    apk add --no-cache make git ca-certificates && \
    update-ca-certificates && \
    adduser \
    --disabled-password \
    --gecos "" \
    --home "/app" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"
ADD . /app
WORKDIR /app
RUN make build

# --- RUNTIME ---
FROM scratch

# Copy required system files
COPY --from=builder /etc/passwd /etc/group /etc/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy application files
COPY --from=builder /app/assets /app/assets
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/config /app/config
COPY --from=builder /app/.env /app/
COPY --from=builder /app/patching-automation.linux /app/patching-automation

USER appuser:appuser
WORKDIR /app
ENTRYPOINT ["/app/patching-automation"]
