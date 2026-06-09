#!/bin/bash

# ==============================================================================
# Production Deployment Script for Baikal Buuzy
# Target Domain: myinternet-tryf.fun
# Target Server: 45.137.81.0
# Target Email:  valento030@gmail.com
# ==============================================================================

# Exit immediately if a command exits with a non-zero status
set -e

DOMAIN="myinternet-tryf.fun"
EMAIL="valento030@gmail.com"
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== Starting deployment preparation for $DOMAIN ==="

# 1. Root user check
if [ "$(id -u)" -ne 0 ]; then
    echo "ERROR: This script must be run as root. Please run it with sudo:"
    echo "sudo $0"
    exit 1
fi

# 2. Package manager verification (supports apt-based systems like Ubuntu/Debian)
if ! command -v apt-get &> /dev/null; then
    echo "ERROR: This script currently only supports apt-based Linux distributions (e.g., Ubuntu)."
    exit 1
fi

# 3. Check and install Docker
echo "Checking Docker..."
if ! command -v docker &> /dev/null; then
    echo "-> Docker not found. Installing Docker..."
    apt-get update
    apt-get install -y curl
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    rm get-docker.sh
    systemctl enable docker
    systemctl start docker
    echo "-> Docker installed successfully."
else
    echo "-> Docker is already installed: $(docker --version)"
fi

# 4. Check and install Docker Compose
echo "Checking Docker Compose..."
if ! docker compose version &> /dev/null; then
    echo "-> Docker Compose not found. Installing Docker Compose plugin..."
    apt-get update
    apt-get install -y docker-compose-plugin
    echo "-> Docker Compose installed successfully."
else
    echo "-> Docker Compose is already installed: $(docker compose version)"
fi

# 5. Check and install Nginx
echo "Checking Nginx..."
if ! command -v nginx &> /dev/null; then
    echo "-> Nginx not found. Installing Nginx..."
    apt-get update
    apt-get install -y nginx
    systemctl enable nginx
    systemctl start nginx
    echo "-> Nginx installed successfully."
else
    echo "-> Nginx is already installed: $(nginx -v 2>&1)"
fi

# 6. Check and install Certbot
echo "Checking Certbot..."
if ! command -v certbot &> /dev/null; then
    echo "-> Certbot not found. Installing Certbot and Nginx plugin..."
    apt-get update
    apt-get install -y certbot python3-certbot-nginx
    echo "-> Certbot installed successfully."
else
    echo "-> Certbot is already installed: $(certbot --version 2>&1)"
fi

# 7. Verify .env file
echo "Checking .env configuration..."
if [ ! -f "$PROJECT_DIR/.env" ]; then
    echo "-> .env file not found in $PROJECT_DIR."
    if [ -f "$PROJECT_DIR/.env.example" ]; then
        echo "-> Creating .env from .env.example with secure random credentials..."
        cp "$PROJECT_DIR/.env.example" "$PROJECT_DIR/.env"
        
        # Generate secure random credentials
        DB_PASS=$(openssl rand -base64 18 | tr -dc 'a-zA-Z0-9' | head -c 18)
        JWT_SEC=$(openssl rand -hex 32)
        
        # Replace placeholders in the copied .env
        sed -i "s/POSTGRES_PASSWORD=ЗАМЕНИТЕ_НА_СЛОЖНЫЙ_ПАРОЛЬ/POSTGRES_PASSWORD=$DB_PASS/g" "$PROJECT_DIR/.env"
        sed -i "s/DATABASE_URL=postgres:\/\/baikal_user:ЗАМЕНИТЕ_НА_СЛОЖНЫЙ_ПАРОЛЬ/DATABASE_URL=postgres:\/\/baikal_user:$DB_PASS/g" "$PROJECT_DIR/.env"
        sed -i "s/JWT_SECRET=ЗАМЕНИТЕ_НА_ДЛИННЫЙ_СЛУЧАЙНЫЙ_СЕКРЕТ_МИНИМУМ_64_СИМВОЛА/JWT_SECRET=$JWT_SEC/g" "$PROJECT_DIR/.env"
        echo "-> .env file successfully created with secure passwords."
    else
        echo "ERROR: Neither .env nor .env.example exists in $PROJECT_DIR. Cannot proceed."
        exit 1
    fi
else
    echo "-> .env file already exists."
fi

# 8. Setup Nginx site configuration
NGINX_CONF="/etc/nginx/sites-available/baikal-buuzy"
echo "Configuring Nginx reverse proxy..."

cat << EOF > "$NGINX_CONF"
server {
    listen 80;
    server_name $DOMAIN www.$DOMAIN;

    client_max_body_size 20M;

    location / {
        proxy_pass http://127.0.0.1:8000;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header X-Forwarded-Host \$host;
        proxy_set_header X-Forwarded-Port \$server_port;
    }
}
EOF

# Enable configuration and disable default if necessary
echo "Enabling Nginx configuration..."
ln -sf "$NGINX_CONF" "/etc/nginx/sites-enabled/baikal-buuzy"

if [ -f "/etc/nginx/sites-enabled/default" ]; then
    echo "-> Disabling default Nginx site configuration..."
    rm -f "/etc/nginx/sites-enabled/default"
fi

# Validate Nginx config
nginx -t

# Reload Nginx
systemctl reload nginx
echo "-> Nginx reloaded successfully."

# 9. Build and start Docker containers
echo "Starting application containers via Docker Compose..."
cd "$PROJECT_DIR"
docker compose down --remove-orphans || true
docker compose up -d --build
echo "-> Docker containers started successfully."

# 10. SSL configuration via Certbot
echo "Setting up Let's Encrypt SSL certificates..."
# Certbot handles modifying Nginx configurations automatically
if [ -d "/etc/letsencrypt/live/$DOMAIN" ]; then
    echo "-> SSL Certificates already exist for $DOMAIN. Reinstalling/reconfiguring..."
    certbot --nginx -d "$DOMAIN" -d "www.$DOMAIN" --non-interactive --agree-tos --email "$EMAIL" --redirect --reinstall
else
    echo "-> Obtaining new SSL certificate from Let's Encrypt..."
    # Run certbot to request certificate
    set +e # Don't exit script if certbot fails (e.g. DNS propagation issues)
    certbot --nginx -d "$DOMAIN" -d "www.$DOMAIN" --non-interactive --agree-tos --email "$EMAIL" --redirect
    CERTBOT_STATUS=$?
    set -e
    if [ $CERTBOT_STATUS -ne 0 ]; then
        echo "WARNING: Certbot failed to obtain certificates. Please check DNS propagation and try running:"
        echo "sudo certbot --nginx -d $DOMAIN -d www.$DOMAIN"
    else
        echo "-> SSL Certificate obtained and Nginx configured successfully!"
    fi
fi

# 11. Final verification
echo "=== Final deployment check ==="
docker compose ps
systemctl status nginx --no-pager | head -n 15

echo "=============================================================================="
echo "SUCCESS: Deployment configuration completed!"
echo "Your website should now be accessible at: https://$DOMAIN"
echo "=============================================================================="
