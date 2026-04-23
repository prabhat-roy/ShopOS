FROM haskell:9.6-slim AS builder

WORKDIR /app

COPY stack.yaml financial-rules-engine.cabal ./
RUN stack setup
RUN stack build --only-dependencies

COPY src src
RUN stack build --copy-bins --local-bin-path /app/bin

FROM debian:12-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    libgmp10 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN useradd -r -u 1001 -s /bin/false nonroot
USER 1001

WORKDIR /app
COPY --from=builder --chown=nonroot:nonroot /app/bin/financial-rules-engine .

EXPOSE 50360

CMD ["./financial-rules-engine"]
