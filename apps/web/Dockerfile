# deps
FROM node:22-alpine3.23 AS deps
WORKDIR /app

COPY package.json package-lock.json* pnpm-lock.yaml* ./
RUN if [ -f pnpm-lock.yaml ]; then corepack enable && pnpm i --frozen-lockfile; \
    elif [ -f package-lock.json ]; then npm ci; \
    else npm i; fi

# dev
FROM node:22-alpine3.23 AS dev
WORKDIR /app
COPY --from=deps /app/node_modules /app/node_modules
COPY . .
EXPOSE 5173
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "5173"]

# build
FROM node:22-alpine3.23 AS build
WORKDIR /app
ENV NODE_ENV=production
COPY --from=deps /app/node_modules /app/node_modules
COPY . .
RUN npm run build

# prod
FROM alpine:3.20 AS prod
WORKDIR /dist
COPY --from=build /app/build/ /dist/

CMD ["sh", "-lc", "rm -rf /out/* && mkdir -p /out && cp -R /dist/* /out/ && echo 'web dist exported to /out' && ls -lah /out && exit 0"]
