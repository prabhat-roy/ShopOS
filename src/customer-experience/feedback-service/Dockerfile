# ---- Build Stage ----
FROM node:22-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm install --omit=dev

# ---- Production Stage ----
FROM node:22-alpine AS production

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/node_modules ./node_modules
COPY package*.json ./
COPY index.js ./
COPY src/ ./src/

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8410

CMD ["node", "index.js"]
