# ============================================================
# Stage 1: Build the maintained client-web frontend
# ============================================================
FROM node:22.22.2-alpine@sha256:8ea2348b068a9544dae7317b4f3aafcdc032df1647bb7d768a05a5cad1a7683f AS builder

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

COPY index.html vite.config.ts tsconfig.json tsconfig.app.json tsconfig.node.json \
    postcss.config.js tailwind.config.js ./
COPY scripts ./scripts
COPY src ./src
ARG VITE_TRUEREPUBLIC_RPC=/rpc
ARG VITE_TRUEREPUBLIC_REST=/api
ENV VITE_TRUEREPUBLIC_RPC=${VITE_TRUEREPUBLIC_RPC}
ENV VITE_TRUEREPUBLIC_REST=${VITE_TRUEREPUBLIC_REST}
RUN npm run build

# ============================================================
# Stage 2: Non-root nginx runtime
# ============================================================
FROM nginx:alpine@sha256:4a73073bd557c65b759505da037898b61f1be6cbcc3c2c3aeac22d2a470c1752

RUN rm /etc/nginx/conf.d/default.conf
COPY nginx.conf /etc/nginx/nginx.conf
COPY --from=builder /app/dist /usr/share/nginx/html

USER nginx

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=5 \
  CMD ["wget", "-q", "-O", "/dev/null", "http://127.0.0.1:8080/"]

CMD ["nginx", "-g", "daemon off;"]
