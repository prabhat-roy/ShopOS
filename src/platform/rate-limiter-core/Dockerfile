FROM ghcr.io/nicholasgasior/docker-zig:0.13.0 AS builder

WORKDIR /app

COPY build.zig .
COPY src src

RUN zig build -Doptimize=ReleaseSafe

FROM debian:12-slim

RUN useradd -r -u 1001 -s /bin/false nonroot
USER 1001

WORKDIR /app
COPY --from=builder --chown=nonroot:nonroot /app/zig-out/bin/rate-limiter-core .

EXPOSE 50355

CMD ["./rate-limiter-core"]
