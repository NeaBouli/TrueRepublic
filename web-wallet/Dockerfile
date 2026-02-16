# ============================================================
# Stage 1: Build React app
# ============================================================
FROM node:20-alpine AS builder

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --production=false

COPY . .
ENV CI=false
RUN npm run build

# ============================================================
# Stage 2: Serve with nginx
# ============================================================
FROM nginx:alpine

RUN rm /etc/nginx/conf.d/default.conf
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/build /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
