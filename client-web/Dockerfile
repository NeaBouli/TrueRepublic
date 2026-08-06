# ============================================================
# Stage 1: Build the maintained client-web frontend
# ============================================================
FROM node:22-alpine AS builder

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

COPY index.html vite.config.ts tsconfig.json tsconfig.app.json tsconfig.node.json \
    postcss.config.js tailwind.config.js ./
COPY src ./src
ARG VITE_TRUEREPUBLIC_RPC=/rpc
ARG VITE_TRUEREPUBLIC_REST=/api
ENV VITE_TRUEREPUBLIC_RPC=${VITE_TRUEREPUBLIC_RPC}
ENV VITE_TRUEREPUBLIC_REST=${VITE_TRUEREPUBLIC_REST}
RUN npm run build

# ============================================================
# Stage 2: Non-root nginx runtime
# ============================================================
FROM nginx:alpine

RUN rm /etc/nginx/conf.d/default.conf
COPY nginx.conf /etc/nginx/nginx.conf
COPY --from=builder /app/dist /usr/share/nginx/html

USER nginx

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=5 \
  CMD ["wget", "-q", "-O", "/dev/null", "http://127.0.0.1:8080/"]

CMD ["nginx", "-g", "daemon off;"]
