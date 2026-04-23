FROM swift:5.10-jammy AS builder

WORKDIR /app

COPY Package.swift Package.resolved* ./
RUN swift package resolve

COPY Sources Sources
RUN swift build -c release --static-swift-stdlib

FROM ubuntu:22.04

RUN apt-get update && apt-get install -y --no-install-recommends \
    libcurl4 \
    libxml2 \
    && rm -rf /var/lib/apt/lists/*

RUN useradd -r -u 1001 -s /bin/false nonroot
USER 1001

WORKDIR /app
COPY --from=builder --chown=nonroot:nonroot /app/.build/release/App .

EXPOSE 8214

CMD ["./App"]
