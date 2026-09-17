FROM node:24-alpine AS base

# Build arguments for configurable environment variables
ARG PREFIX=/api/v1
ARG BACKEND=http://backend:3001

RUN apk update && apk upgrade && \
    apk add --no-cache libc6-compat dumb-init && \
    rm -rf /var/cache/apk/*

WORKDIR /app

# Stage 1: Install dependencies
FROM base AS deps

COPY package.json package-lock.json* ./

# Build dependencies are required for Next.js TypeScript compilation. Only the
# generated standalone output is copied into the runtime image.
RUN npm ci --no-audit --no-fund

# Stage 2: Build the application
FROM base AS builder

COPY --from=deps /app/node_modules ./node_modules
COPY . .

# Ensure public directory exists (create empty one if it doesn't)
RUN mkdir -p public

ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production
ENV PREFIX=${PREFIX}
ENV BACKEND=${BACKEND}

RUN npm run build

# Stage 3: Create a minimal production image
FROM base AS runner

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ENV PREFIX=${PREFIX}
ENV BACKEND=${BACKEND}

RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

WORKDIR /app

# Create .next directory with proper ownership
RUN mkdir -p .next && \
    chown -R nextjs:nodejs .next

# Create public directory and copy assets if they exist
RUN mkdir -p public

# Copy public assets
COPY --from=builder /app/public ./public/

COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD node -e "require('http').get('http://localhost:3000/api/health', (res) => { process.exit(res.statusCode === 200 ? 0 : 1) })" || exit 1

ENTRYPOINT ["dumb-init", "--"]

CMD ["node", "server.js"]
