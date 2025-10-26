#!/bin/bash

# Ecommerce Backend Deployment Script for aaPanel
# This script is designed to work with your existing aaPanel setup

set -e

echo "🚀 Deploying Ecommerce Backend with aaPanel..."

# Create application directory in aaPanel's www directory
echo "📁 Creating application directory..."
sudo mkdir -p /www/wwwroot/ecommerce-backend
sudo chown -R www:www /www/wwwroot/ecommerce-backend

# Create systemd service file
echo "⚙️ Setting up systemd service..."
sudo tee /etc/systemd/system/ecommerce-backend.service > /dev/null <<EOF
[Unit]
Description=Ecommerce Backend Service
After=network.target

[Service]
Type=simple
User=www
WorkingDirectory=/www/wwwroot/ecommerce-backend
ExecStart=/www/wwwroot/ecommerce-backend/ecommerce-backend
Restart=always
RestartSec=5
Environment=PORT=8080
Environment=MONGODB_URI=mongodb://admin:SecureMongoDB123!@localhost:27017
Environment=DATABASE_NAME=ecommerce
Environment=JWT_SECRET=production-secret-key-change-this
Environment=ENV=production

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
sudo systemctl daemon-reload
sudo systemctl enable ecommerce-backend

# Create environment file
echo "📝 Creating environment file..."
sudo tee /www/wwwroot/ecommerce-backend/.env > /dev/null <<EOF
MONGODB_URI=mongodb://admin:SecureMongoDB123!@localhost:27017
DATABASE_NAME=ecommerce
JWT_SECRET=production-secret-key-change-this
PORT=8080
ENV=production
EOF

# Set permissions
sudo chown -R www:www /www/wwwroot/ecommerce-backend
sudo chmod +x /www/wwwroot/ecommerce-backend

echo "✅ aaPanel deployment setup complete!"
echo ""
echo "Next steps:"
echo "1. Configure Nginx in aaPanel"
echo "2. Set up GitHub Secrets"
echo "3. Deploy via GitHub Actions"
echo ""
echo "To check service status:"
echo "sudo systemctl status ecommerce-backend"
