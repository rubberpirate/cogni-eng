FROM oven/bun:1.4.4 AS compile
WORKDIR /source
COPY package.json bun.lock ./
COPY apps/web/package.json ./apps/web/package.json
RUN bun install --frozen-lockfile
COPY apps/web ./apps/web
COPY clients ./clients
ENV NEXT_TELEMETRY_DISABLED=1
ARG API_INTERNAL_URL=http://gateway:8080
ENV API_INTERNAL_URL=$API_INTERNAL_URL
RUN bun run build

FROM oven/bun:1.4.4-slim AS runtime
WORKDIR /app
ENV NODE_ENV=production NEXT_TELEMETRY_DISABLED=1 HOSTNAME=0.0.0.0 PORT=3000
COPY --from=compile --chown=bun:bun /source/apps/web/.next/standalone ./
COPY --from=compile --chown=bun:bun /source/apps/web/.next/static ./apps/web/.next/static
USER bun
EXPOSE 3000
CMD ["bun", "apps/web/server.js"]
