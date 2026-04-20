# ---- Build stage ----
FROM node:22-alpine AS builder

WORKDIR /app

# Copy dependency manifests first to exploit layer cache
COPY package*.json ./
RUN npm install --omit=dev

# Copy application source
COPY src ./src
COPY index.js ./

# ---- Runtime stage ----
FROM node:22-alpine
RUN addgroup -g 10001 appgroup && adduser -u 10001 -G appgroup -D appuser
WORKDIR /app
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/src ./src
COPY --from=builder /app/index.js ./
RUN chown -R appuser:appgroup /app
USER appuser
EXPOSE 8080
CMD ["node", "index.js"]