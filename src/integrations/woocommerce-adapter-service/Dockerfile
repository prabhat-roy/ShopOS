FROM php:8.3-fpm-alpine AS builder

RUN apk add --no-cache \
    curl \
    unzip \
    libzip-dev \
    && docker-php-ext-install zip pdo pdo_mysql

COPY --from=composer:2.7 /usr/bin/composer /usr/bin/composer

WORKDIR /app

COPY composer.json composer.lock* ./
RUN composer install --no-dev --optimize-autoloader --no-scripts

COPY . .
RUN composer dump-autoload --optimize

FROM php:8.3-fpm-alpine

RUN apk add --no-cache nginx libzip && docker-php-ext-install zip pdo pdo_mysql

RUN adduser -D -u 1001 nonroot

WORKDIR /app

COPY --from=builder --chown=nonroot:nonroot /app .

COPY <<'EOF' /etc/nginx/http.d/default.conf
server {
    listen 8216;
    root /app/public;
    index index.php;

    location /healthz {
        try_files $uri /index.php?$query_string;
    }

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        fastcgi_pass 127.0.0.1:9000;
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }
}
EOF

EXPOSE 8216

CMD ["sh", "-c", "php-fpm -D && nginx -g 'daemon off;'"]
