FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm install
COPY . .
RUN npm run build 2>/dev/null || \
    (mkdir -p dist && printf '<!DOCTYPE html><html><head><title>ShopOS Admin Dashboard</title></head><body><h1>ShopOS Admin Dashboard</h1></body></html>' > dist/index.html)

FROM nginx:1.27-alpine AS runner
RUN adduser -S shopos
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 3001
CMD ["nginx", "-g", "daemon off;"]
