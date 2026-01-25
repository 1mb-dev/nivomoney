# =============================================================================
# Nivo Frontend - Multi-stage Docker Build
# =============================================================================

# =============================================================================
# Stage 1: Build shared package
# =============================================================================
FROM node:20-alpine AS shared-builder

WORKDIR /app/frontend

# Copy shared package
COPY shared/ ./shared/

# Build shared package
WORKDIR /app/frontend/shared
RUN npm install --silent && npm run build

# =============================================================================
# Stage 2: Build user-app
# =============================================================================
FROM node:20-alpine AS user-app-builder

WORKDIR /app/frontend

ARG VITE_API_URL=https://api.nivomoney.com

# Copy shared (with build artifacts)
COPY --from=shared-builder /app/frontend/shared ./shared

# Copy user-app package files
COPY user-app/package*.json ./user-app/

# Update @nivo/shared to use local file path and install
WORKDIR /app/frontend/user-app
RUN sed -i 's|"@nivo/shared": "[^"]*"|"@nivo/shared": "file:../shared"|' package.json && \
    npm install --silent

# Copy source and build
COPY user-app/ ./
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

# =============================================================================
# Stage 3: Build admin-app
# =============================================================================
FROM node:20-alpine AS admin-app-builder

WORKDIR /app/frontend

ARG VITE_API_URL=https://api.nivomoney.com

# Copy shared (with build artifacts)
COPY --from=shared-builder /app/frontend/shared ./shared

# Copy admin-app package files
COPY admin-app/package*.json ./admin-app/

# Update @nivo/shared to use local file path and install
WORKDIR /app/frontend/admin-app
RUN sed -i 's|"@nivo/shared": "[^"]*"|"@nivo/shared": "file:../shared"|' package.json && \
    npm install --silent

# Copy source and build
COPY admin-app/ ./
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

# =============================================================================
# Stage 4: Build user-admin-app
# =============================================================================
FROM node:20-alpine AS user-admin-app-builder

WORKDIR /app/frontend

ARG VITE_API_URL=https://api.nivomoney.com

# Copy shared (with build artifacts)
COPY --from=shared-builder /app/frontend/shared ./shared

# Copy user-admin-app package files
COPY user-admin-app/package*.json ./user-admin-app/

# Update @nivo/shared to use local file path and install
WORKDIR /app/frontend/user-admin-app
RUN sed -i 's|"@nivo/shared": "[^"]*"|"@nivo/shared": "file:../shared"|' package.json && \
    npm install --silent

# Copy source and build
COPY user-admin-app/ ./
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

# =============================================================================
# Stage 5: Production nginx image
# =============================================================================
FROM nginx:1.25-alpine AS production

RUN rm /etc/nginx/conf.d/default.conf

RUN mkdir -p /usr/share/nginx/html/user-app \
    && mkdir -p /usr/share/nginx/html/admin-app \
    && mkdir -p /usr/share/nginx/html/user-admin-app \
    && mkdir -p /var/cache/nginx/client_temp \
    && mkdir -p /var/cache/nginx/proxy_temp \
    && mkdir -p /var/cache/nginx/fastcgi_temp \
    && mkdir -p /var/cache/nginx/uwsgi_temp \
    && mkdir -p /var/cache/nginx/scgi_temp \
    && mkdir -p /var/www/certbot

COPY --from=user-app-builder /app/frontend/user-app/dist /usr/share/nginx/html/user-app
COPY --from=admin-app-builder /app/frontend/admin-app/dist /usr/share/nginx/html/admin-app
COPY --from=user-admin-app-builder /app/frontend/user-admin-app/dist /usr/share/nginx/html/user-admin-app

RUN chown -R nginx:nginx /usr/share/nginx/html \
    && chown -R nginx:nginx /var/cache/nginx \
    && chown -R nginx:nginx /var/log/nginx \
    && chown -R nginx:nginx /var/www/certbot \
    && touch /var/run/nginx.pid \
    && chown nginx:nginx /var/run/nginx.pid

EXPOSE 80 443

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q -O - http://localhost/nginx-health || exit 1

CMD ["nginx", "-g", "daemon off;"]
