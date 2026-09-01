FROM node:24-alpine AS dependencies

WORKDIR /workspace
RUN corepack enable

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/web/package.json apps/web/package.json
RUN pnpm install --frozen-lockfile

FROM dependencies AS build

COPY apps/web apps/web
RUN pnpm --filter web build

FROM node:24-alpine AS runtime

ENV NODE_ENV=production
ENV HOSTNAME=0.0.0.0
ENV PORT=3000
WORKDIR /app

RUN apk upgrade --no-cache \
    && rm -rf /usr/local/lib/node_modules/npm /usr/local/lib/node_modules/corepack \
    && rm -f /usr/local/bin/npm /usr/local/bin/npx /usr/local/bin/corepack /usr/local/bin/pnpm /usr/local/bin/pnpx \
    && addgroup -S cartlabs \
    && adduser -S -G cartlabs cartlabs

COPY --from=build --chown=cartlabs:cartlabs /workspace/apps/web/.next/standalone ./
COPY --from=build --chown=cartlabs:cartlabs /workspace/apps/web/.next/static ./apps/web/.next/static
COPY --from=build --chown=cartlabs:cartlabs /workspace/apps/web/public ./apps/web/public

USER cartlabs
EXPOSE 3000
CMD ["node", "apps/web/server.js"]
