# ---- Build stage: собираем Svelte SPA ----
FROM node:22-alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей для кеширования слоев
COPY package.json package-lock.json ./

# Устанавливаем зависимости
RUN npm install

# Копируем весь исходный код фронтенда
COPY . .

# Собираем production-бандл
RUN npm run build

# ---- Runtime stage: Nginx отдает статику + reverse proxy ----
FROM nginx:1.27-alpine

# Удаляем дефолтный конфиг
RUN rm /etc/nginx/conf.d/default.conf

# Копируем наш конфиг Nginx
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Копируем собранный SPA из builder-стадии
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
