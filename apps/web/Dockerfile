# build
FROM node:22-alpine3.23 AS base
WORKDIR /app

COPY package.json package-lock.json* pnpm-lock.yaml* ./
RUN if [ -f pnpm-lock.yaml ]; then corepack enable && pnpm i; \
    elif [ -f package-lock.json ]; then npm ci; \
    else npm i; fi

COPY . .

# dev
FROM base AS dev
EXPOSE 5173
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "5173"]

# prod
FROM alpine:3.20 AS prod
WORKDIR /out

RUN printf '%s\n' \
'<!doctype html>' \
'<html><head><meta charset="utf-8"/><meta name="viewport" content="width=device-width,initial-scale=1"/>' \
'<title>Kids Planet</title></head>' \
'<body style="font-family:system-ui,sans-serif;padding:24px">' \
'<h1>Kids Planet</h1><p>Nginx static OK (200)</p><p>API health: <code>/api/health</code></p>' \
'</body></html>' \
> /out/index.html

CMD ["sh", "-lc", "echo web placeholder generated; ls -lah /out; exit 0"]