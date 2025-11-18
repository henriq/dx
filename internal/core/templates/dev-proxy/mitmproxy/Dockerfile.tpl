FROM python:3.12-slim as builder

WORKDIR /app

RUN pip install --no-cache-dir --target=/app/dependencies mitmproxy

# Remove unnecessary files and cached data
RUN find /app/dependencies -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true && \
    find /app/dependencies -type f -name "*.pyc" -delete && \
    find /app/dependencies -type f -name "*.pyo" -delete && \
    find /app/dependencies -type d -name "*.dist-info" -exec rm -rf {}/RECORD {} + 2>/dev/null || true

FROM python:3.12-slim

RUN useradd --no-create-home --shell /usr/sbin/nologin --uid 65532 --user-group nonroot && \
    mkdir -p /home/nonroot && \
    chown -R nonroot:nonroot /home/nonroot

# Harden the system - update packages and minimize attack surface
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get autoremove -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* && \
    # Remove suid/sgid binaries to prevent privilege escalation
    find / -xdev -perm /6000 -type f -exec chmod a-s {} \; 2>/dev/null || true

COPY --from=builder --chown=nonroot:nonroot /app/dependencies /app/dependencies

ENV PATH="/app/dependencies/bin:${PATH}" \
    PYTHONPATH="/app/dependencies" \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    HOME=/home/nonroot

WORKDIR /home/nonroot
USER nonroot:nonroot

ENTRYPOINT ["/app/dependencies/bin/mitmweb"]