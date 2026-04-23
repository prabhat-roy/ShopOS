FROM ghcr.io/cirruslabs/flutter:stable AS builder

WORKDIR /app

COPY pubspec.yaml pubspec.lock* ./
RUN flutter pub get

COPY . .
RUN flutter build web --release

FROM nginx:1.27-alpine

RUN adduser -D -u 1001 nonroot

COPY --from=builder --chown=nonroot:nonroot /app/build/web /usr/share/nginx/html

COPY <<'EOF' /etc/nginx/conf.d/default.conf
server {
    listen 3005;
    root /usr/share/nginx/html;
    index index.html;

    location /healthz {
        return 200 '{"status":"ok"}';
        add_header Content-Type application/json;
    }

    location / {
        try_files $uri $uri/ /index.html;
    }
}
EOF

EXPOSE 3005

CMD ["nginx", "-g", "daemon off;"]
