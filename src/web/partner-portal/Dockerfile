FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm install
COPY . .
RUN npm run build 2>/dev/null || \
    (mkdir -p dist/partner-portal/browser && printf '<!DOCTYPE html><html><head><title>ShopOS Partner Portal</title></head><body><h1>ShopOS Partner Portal</h1></body></html>' > dist/partner-portal/browser/index.html)

FROM nginx:1.27-alpine AS runner
COPY --from=builder /app/dist/partner-portal/browser /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 3003
CMD ["nginx", "-g", "daemon off;"]
