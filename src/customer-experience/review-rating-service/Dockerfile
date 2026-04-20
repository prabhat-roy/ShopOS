# ── Stage 1: dependencies ──────────────────────────────────────────────────
FROM node:22-alpine AS deps

WORKDIR /app

COPY package.json package-lock.json* ./

RUN npm install --omit=dev && npm cache clean --force

# ── Stage 2: production image ──────────────────────────────────────────────
FROM node:22-alpine AS runner

# Install dumb-init for proper PID 1 signal handling
RUN apk add --no-cache dumb-init

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy production node_modules from deps stage
COPY --from=deps --chown=appuser:appgroup /app/node_modules ./node_modules

# Copy application source
COPY --chown=appuser:appgroup index.js ./
COPY --chown=appuser:appgroup src/ ./src/

USER appuser

EXPOSE 8400

ENV NODE_ENV=production

ENTRYPOINT ["dumb-init", "--"]
CMD ["node", "index.js"]
