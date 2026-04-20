# Stage 1: Build / dependency install
FROM node:22-alpine AS deps

WORKDIR /app

COPY package*.json ./
RUN npm install --omit=dev

# Stage 2: Final runtime image
FROM node:22-alpine AS runtime

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy production node_modules from deps stage
COPY --from=deps /app/node_modules ./node_modules

# Copy application source
COPY --chown=appuser:appgroup . .

# Drop to non-root user
USER appuser

EXPOSE 8204

ENV NODE_ENV=production

CMD ["node", "index.js"]
