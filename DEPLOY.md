# 🚀 Гайд по деплою: Baikal Buuzy → myinternet-tryf.fun

> **Стек**: Go (Chi) Backend + Svelte/Vite Frontend + PostgreSQL 15  
> **Домен**: `myinternet-tryf.fun`  
> **Метод**: Docker Compose + Nginx reverse proxy + Let's Encrypt SSL

---

## 📋 Архитектура деплоя

```
Internet
    │
    ▼
[Nginx :80/:443]  ←── SSL termination (Let's Encrypt)
    │
    ├── /api/*      → Go Backend  :8080
    ├── /uploads/*  → Go Backend  :8080 (статика загрузок)
    └── /*          → Svelte SPA  (статические файлы из /dist)
```

---

## 🖥️ Шаг 1: Подготовка VPS

### Требования к серверу
- **OS**: Ubuntu 22.04 LTS (рекомендуется)
- **RAM**: минимум 1 GB (рекомендуется 2 GB)
- **Disk**: минимум 20 GB SSD
- **Провайдеры**: Timeweb, Beget, REG.RU, Selectel, DigitalOcean, Hetzner

### 1.1 Подключение к серверу

```bash
ssh root@<IP_ВАШЕГО_VPS>
```

### 1.2 Обновление системы и установка зависимостей

```bash
apt update && apt upgrade -y

# Основные инструменты
apt install -y git curl wget unzip ufw fail2ban

# Docker
curl -fsSL https://get.docker.com | sh
systemctl enable docker
systemctl start docker

# Docker Compose Plugin
apt install -y docker-compose-plugin

# Nginx
apt install -y nginx

# Certbot (Let's Encrypt)
apt install -y certbot python3-certbot-nginx

# Проверка версий
docker --version
docker compose version
nginx -v
```

### 1.3 Настройка Firewall

```bash
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh          # 22
ufw allow http         # 80
ufw allow https        # 443
ufw enable
ufw status
```

### 1.4 Создание системного пользователя (для безопасности)

```bash
adduser deployer
usermod -aG docker deployer
usermod -aG sudo deployer

# Настройка SSH для нового пользователя
mkdir -p /home/deployer/.ssh
cp ~/.ssh/authorized_keys /home/deployer/.ssh/
chown -R deployer:deployer /home/deployer/.ssh
chmod 700 /home/deployer/.ssh
chmod 600 /home/deployer/.ssh/authorized_keys
```

---

## 📁 Шаг 2: Структура проекта на сервере

```bash
# Переключаемся на deployer
su - deployer

# Создаем рабочую директорию
mkdir -p /home/deployer/baikal-buuzy
cd /home/deployer/baikal-buuzy
```

---

## 🌐 Шаг 3: Настройка DNS

В панели управления вашим доменом `myinternet-tryf.fun` добавьте A-записи:

| Тип | Имя | Значение | TTL |
|-----|-----|----------|-----|
| A | `@` | `<IP_ВАШЕГО_VPS>` | 3600 |
| A | `www` | `<IP_ВАШЕГО_VPS>` | 3600 |

> [!IMPORTANT]
> DNS может обновляться до 24–48 часов. Проверить распространение: [dnschecker.org](https://dnschecker.org)

---

## 🔧 Шаг 4: Подготовка проекта к деплою

### 4.1 Сборка Frontend (локально на Windows)

```powershell
# В PowerShell на вашем ПК
cd C:\Users\tryf\Desktop\diplom\frontend

# Установка зависимостей
npm install

# Сборка для production
npm run build
```

После сборки появится папка `frontend/dist/` — это и есть ваш готовый сайт.

### 4.2 Создание Dockerfile для Backend

Создайте файл `backend/Dockerfile`:

```dockerfile
# ---- Build stage ----
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Копируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/app

# ---- Runtime stage ----
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Irkutsk

WORKDIR /app

# Копируем бинарник и миграции
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations

# Создаем папку для загрузок
RUN mkdir -p /app/uploads

EXPOSE 8080

CMD ["./server"]
```

### 4.3 Создание production .env файла

> [!CAUTION]
> Никогда не коммитьте `.env` в Git! Добавьте его в `.gitignore`.

Создайте файл `/home/deployer/baikal-buuzy/.env` прямо на сервере:

```bash
nano /home/deployer/baikal-buuzy/.env
```

Содержимое (замените значения!):

```env
# Database
POSTGRES_USER=baikal_user
POSTGRES_PASSWORD=СЮДА_СЛОЖНЫЙ_ПАРОЛЬ_ОТ_БД
POSTGRES_DB=baikal_buuzy

# Backend
PORT=8080
DATABASE_URL=postgres://baikal_user:СЮДА_СЛОЖНЫЙ_ПАРОЛЬ_ОТ_БД@db:5432/baikal_buuzy?sslmode=disable
JWT_SECRET=СЮДА_ДЛИННЫЙ_СЛУЧАЙНЫЙ_СЕКРЕТ_JWT
```

Генерация надежных секретов:
```bash
# Генерация пароля для БД
openssl rand -base64 32

# Генерация JWT секрета
openssl rand -hex 64
```

### 4.4 Обновление docker-compose.yml

Замените содержимое `docker-compose.yml` в корне проекта:

```yaml
services:
  db:
    image: postgres:15-alpine
    container_name: baikal_buuzy_db
    restart: always
    env_file: .env
    environment:
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: ${POSTGRES_DB}
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB}"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - app-network

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: baikal_buuzy_backend
    restart: always
    env_file: .env
    environment:
      PORT: ${PORT}
      DATABASE_URL: ${DATABASE_URL}
      JWT_SECRET: ${JWT_SECRET}
    volumes:
      - uploads_data:/app/uploads
    depends_on:
      db:
        condition: service_healthy
    ports:
      - "8080:8080"
    networks:
      - app-network

volumes:
  pgdata:
  uploads_data:

networks:
  app-network:
    driver: bridge
```

---

## 📤 Шаг 5: Загрузка файлов на сервер

### Вариант A: Через Git (рекомендуется)

```bash
# На сервере
cd /home/deployer/baikal-buuzy
git clone https://github.com/ВАШ_АККАУНТ/ВАШ_РЕПО.git .
```

### Вариант Б: Через SCP (из PowerShell на Windows)

```powershell
# Из PowerShell на вашем ПК
# Загрузка всего проекта (кроме node_modules и .git)
scp -r C:\Users\tryf\Desktop\diplom\backend deployer@<IP>:/home/deployer/baikal-buuzy/
scp -r C:\Users\tryf\Desktop\diplom\frontend\dist deployer@<IP>:/home/deployer/baikal-buuzy/frontend/
scp C:\Users\tryf\Desktop\diplom\docker-compose.yml deployer@<IP>:/home/deployer/baikal-buuzy/
```

> [!TIP]
> Для удобства можно использовать **WinSCP** (GUI) или **rsync** для синхронизации файлов.

---

## ⚙️ Шаг 6: Настройка Nginx

### 6.1 Создание конфига для домена

```bash
nano /etc/nginx/sites-available/baikal-buuzy
```

Содержимое:

```nginx
server {
    listen 80;
    server_name myinternet-tryf.fun www.myinternet-tryf.fun;

    # Максимальный размер загружаемых файлов (для картинок)
    client_max_body_size 20M;

    # Расположение собранного Svelte SPA
    root /home/deployer/baikal-buuzy/frontend/dist;
    index index.html;

    # Статические ресурсы с кешированием
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
        try_files $uri =404;
    }

    # Проксирование API на Go бэкенд
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 60s;
        proxy_connect_timeout 10s;
    }

    # Проксирование загруженных файлов (фото блюд)
    location /uploads/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_read_timeout 60s;
    }

    # SPA routing — все не найденные пути отдают index.html
    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

### 6.2 Активация конфига

```bash
# Создаем симлинк
ln -s /etc/nginx/sites-available/baikal-buuzy /etc/nginx/sites-enabled/

# Проверка конфига
nginx -t

# Перезапуск
systemctl reload nginx
```

---

## 🔒 Шаг 7: SSL-сертификат (HTTPS)

> [!IMPORTANT]
> DNS должен быть уже настроен и распространился, иначе certbot не пройдет проверку.

```bash
# Получение сертификата Let's Encrypt
certbot --nginx -d myinternet-tryf.fun -d www.myinternet-tryf.fun \
  --non-interactive --agree-tos --email ВАШ_EMAIL@gmail.com

# Проверка автообновления
certbot renew --dry-run
```

После этого Nginx автоматически обновится и будет слушать на 443 с HTTPS. Certbot настроит автопродление через cron/systemd.

---

## 🐳 Шаг 8: Запуск приложения

```bash
cd /home/deployer/baikal-buuzy

# Первый запуск (с построением образов)
docker compose up -d --build

# Проверка статуса
docker compose ps

# Просмотр логов
docker compose logs -f backend
docker compose logs -f db
```

### Ожидаемый вывод `docker compose ps`:

```
NAME                    STATUS
baikal_buuzy_db         Up (healthy)
baikal_buuzy_backend    Up
```

---

## ✅ Шаг 9: Проверка работоспособности

```bash
# Проверка API (должен вернуть список продуктов)
curl https://myinternet-tryf.fun/api/products

# Проверка загрузки фронтенда
curl -I https://myinternet-tryf.fun

# Проверка статуса Nginx
systemctl status nginx

# Проверка статуса Docker контейнеров
docker compose ps
```

---

## 🔄 Шаг 10: Обновление приложения (Workflow)

Каждый раз когда нужно выкатить новую версию:

```bash
cd /home/deployer/baikal-buuzy

# 1. Получаем новый код (если через git)
git pull origin main

# 2. Загружаем новый frontend/dist (если через scp)
# scp -r ... 

# 3. Пересобираем и перезапускаем
docker compose up -d --build

# 4. Reload Nginx если менялся конфиг
nginx -t && systemctl reload nginx
```

---

## 📊 Шаг 11: Мониторинг и логи

### Просмотр логов

```bash
# Логи бэкенда в реальном времени
docker compose logs -f backend

# Логи базы данных
docker compose logs -f db

# Логи Nginx
tail -f /var/log/nginx/access.log
tail -f /var/log/nginx/error.log
```

### Использование ресурсов

```bash
# Потребление CPU/RAM контейнерами
docker stats

# Место на диске
df -h
docker system df
```

---

## 🛡️ Шаг 12: Безопасность (обязательно)

### 12.1 Смена дефолтных паролей в коде

В файле `backend/cmd/app/main.go` указан хардкод пароль главного администратора:
```go
phone: "+7988548955"
password: "123456"
```

> [!CAUTION]
> **Обязательно смените пароль** через API после первого запуска, или измените код перед деплоем, чтобы не оставлять стандартные credentials на боевом сервере!

### 12.2 Fail2ban (защита от брутфорса)

```bash
systemctl enable fail2ban
systemctl start fail2ban
```

### 12.3 Автоматические обновления безопасности

```bash
apt install -y unattended-upgrades
dpkg-reconfigure --priority=low unattended-upgrades
```

---

## 🐞 Troubleshooting

| Проблема | Диагностика | Решение |
|----------|-------------|---------|
| Сайт не открывается | `curl -I http://myinternet-tryf.fun` | Проверить DNS, firewall, nginx status |
| API возвращает 502 | `docker compose ps` | Backend не запущен: `docker compose logs backend` |
| SSL не работает | `certbot certificates` | DNS не распространился, подождать и повторить certbot |
| БД недоступна | `docker compose logs db` | Проверить `.env`, пароли, healthcheck |
| Нет прав на папку | `ls -la /home/deployer` | `chown -R deployer:deployer /home/deployer/baikal-buuzy` |
| Большой размер образов | `docker system df` | `docker system prune -f` |

---

## 📁 Итоговая структура на сервере

```
/home/deployer/baikal-buuzy/
├── .env                    ← секреты (не в git!)
├── docker-compose.yml
├── backend/
│   ├── Dockerfile          ← новый файл
│   ├── cmd/app/main.go
│   ├── internal/
│   ├── migrations/
│   └── go.mod
└── frontend/
    └── dist/               ← собранный Svelte SPA
        ├── index.html
        └── assets/
```

---

## 📝 Чеклист перед деплоем

- [ ] VPS куплен и доступен по SSH
- [ ] Docker + Docker Compose установлены
- [ ] Nginx установлен
- [ ] DNS A-записи настроены на IP сервера
- [ ] `backend/Dockerfile` создан
- [ ] `docker-compose.yml` обновлен
- [ ] `.env` создан на сервере с сильными паролями
- [ ] `frontend/dist/` собран (`npm run build`)
- [ ] Файлы загружены на сервер
- [ ] Nginx конфиг создан и активирован
- [ ] SSL сертификат получен через certbot
- [ ] `docker compose up -d --build` выполнен
- [ ] API проверен через curl
- [ ] Сайт открывается по HTTPS
- [ ] Дефолтный пароль admin сменен
