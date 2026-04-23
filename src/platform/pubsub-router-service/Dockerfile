FROM elixir:1.17-alpine AS builder

WORKDIR /app

RUN apk add --no-cache build-base git

COPY mix.exs mix.lock* ./
RUN mix local.hex --force && mix local.rebar --force
RUN mix deps.get --only prod

COPY config config
COPY lib lib

ENV MIX_ENV=prod
RUN mix deps.compile
RUN mix release

FROM alpine:3.20

RUN apk add --no-cache libstdc++ openssl ncurses-libs

RUN adduser -D -u 1001 nonroot
USER 1001

WORKDIR /app

COPY --from=builder --chown=nonroot:nonroot /app/_build/prod/rel/pubsub_router ./

EXPOSE 50352

CMD ["bin/pubsub_router", "start"]
