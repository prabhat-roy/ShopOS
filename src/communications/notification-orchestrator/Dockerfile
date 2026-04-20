# ─── Stage 1: dependencies ────────────────────────────────────────────────────
FROM node:22-alpine AS deps

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm install --omit=dev

# ─── Stage 2: final image ─────────────────────────────────────────────────────
FROM node:22-alpine

RUN addgroup -g 10001 appgroup && adduser -u 10001 -G appgroup -D appuser

WORKDIR /app

COPY --from=deps /app/node_modules ./node_modules
COPY . .

RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8080
CMD ["node", "index.js"]