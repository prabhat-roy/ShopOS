FROM ruby:3.3-alpine AS builder

RUN apk add --no-cache build-base

WORKDIR /app

COPY Gemfile Gemfile.lock* ./
RUN bundle install --without development test

COPY . .

FROM ruby:3.3-alpine

RUN apk add --no-cache libstdc++

RUN adduser -D -u 1001 nonroot
USER 1001

WORKDIR /app
COPY --from=builder --chown=nonroot:nonroot /app .
COPY --from=builder --chown=nonroot:nonroot /usr/local/bundle /usr/local/bundle

EXPOSE 8217

CMD ["bundle", "exec", "puma", "-C", "config/puma.rb", "config.ru"]
