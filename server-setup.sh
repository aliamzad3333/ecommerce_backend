#!/bin/bash

# Ecommerce Backend Server Setup Script
# Run this script on your server to set up the environment

set -e

echo "🚀 Setting up Ecommerce Backend on server..."

# Update system
echo "📦 Updating system packages..."
sudo apt update && sudo apt upgrade -y

# Install required packages
echo "🔧 Installing required packages..."
sudo apt install -y curl wget git nginx mongodb

# Create application directory
echo "📁 Creating application directory..."
sudo mkdir -p /var/www/ecommerce-backend
sudo chown -R www-data:www-data /var/www/ecommerce-backend

# Create systemd service file
echo "⚙️ Setting up systemd service..."
sudo tee /etc/systemd/system/ecommerce-backend.service > /dev/null <<EOF
[Unit]
Description=Ecommerce Backend Service
After=network.target mongodb.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/var/www/ecommerce-backend
ExecStart=/var/www/ecommerce-backend/ecommerce-backend
Restart=always
RestartSec=5
Environment=PORT=8080
Environment=MONGODB_URI=mongodb://localhost:27017
Environment=DATABASE_NAME=ecommerce
Environment=JWT_SECRET=production-secret-key-change-this
Environment=ENV=production

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
sudo systemctl daemon-reload
sudo systemctl enable ecommerce-backend

# Configure Nginx
echo "🌐 Configuring Nginx..."
sudo tee /etc/nginx/sites-available/ecommerce-api > /dev/null <<EOF
server {
    listen 80;
    server_name 130.94.40.85;

    # API routes
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        
        # CORS headers
        add_header Access-Control-Allow-Origin *;
        add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
        add_header Access-Control-Allow-Headers "Content-Type, Authorization";
        
        # Handle preflight requests
        if (\$request_method = 'OPTIONS') {
            add_header Access-Control-Allow-Origin *;
            add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
            add_header Access-Control-Allow-Headers "Content-Type, Authorization";
            add_header Access-Control-Max-Age 1728000;
            add_header Content-Type 'text/plain; charset=utf-8';
            add_header Content-Length 0;
            return 204;
        }
    }

    # Health check endpoint
    location /health {
        proxy_pass http://localhost:8080/health;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    # Frontend (if you have one)
    location / {
        root /var/www/ecommerce-frontend;
        try_files \$uri \$uri/ /index.html;
    }
}
EOF

# Enable the site
sudo ln -sf /etc/nginx/sites-available/ecommerce-api /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default

# Test Nginx configuration
sudo nginx -t

# Start services
echo "🔄 Starting services..."
sudo systemctl start mongodb
sudo systemctl enable mongodb
sudo systemctl restart nginx
sudo systemctl enable nginx

# Create environment file
echo "📝 Creating environment file..."
sudo tee /var/www/ecommerce-backend/.env > /dev/null <<EOF
MONGODB_URI=mongodb://localhost:27017
DATABASE_NAME=ecommerce
JWT_SECRET=production-secret-key-change-this
PORT=8080
ENV=production
EOF

# Set permissions
sudo chown -R www-data:www-data /var/www/ecommerce-backend
sudo chmod +x /var/www/ecommerce-backend

echo "✅ Server setup complete!"
echo ""
echo "Next steps:"
echo "1. Set up GitHub Secrets in your repository:"
echo "   - HOST: 130.94.40.85"
echo "   - USERNAME: your-ssh-username"
echo "   - SSH_KEY: your-private-ssh-key"
echo "   - PORT: 22"
echo ""
echo "2. Push to main branch to trigger deployment"
echo ""
echo "3. Test your API:"
echo "   - Health: http://130.94.40.85/health"
echo "   - API: http://130.94.40.85/api/auth/register"
echo ""
echo "To check service status:"
echo "sudo systemctl status ecommerce-backend"
echo "sudo systemctl status nginx"
echo "sudo systemctl status mongodb"
