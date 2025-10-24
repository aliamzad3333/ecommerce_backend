#!/bin/bash

# Quick Deployment Script for aaPanel Server
# Run this on your server: 130.94.40.85

echo "🚀 Starting Ecommerce Backend Deployment..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root or with sudo"
    exit 1
fi

# Create application directory
echo "📁 Creating application directory..."
mkdir -p /www/wwwroot/ecommerce-backend
chown -R www:www /www/wwwroot/ecommerce-backend

# Create systemd service file
echo "⚙️ Creating systemd service..."
cat > /etc/systemd/system/ecommerce-backend.service << 'EOF'
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
Environment=MONGODB_URI=mongodb://localhost:27017
Environment=DATABASE_NAME=Ecommerce_data
Environment=JWT_SECRET=production-secret-key-change-this
Environment=ENV=production

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
systemctl daemon-reload
systemctl enable ecommerce-backend

# Create environment file
echo "📝 Creating environment file..."
cat > /www/wwwroot/ecommerce-backend/.env << 'EOF'
MONGODB_URI=mongodb://localhost:27017
DATABASE_NAME=Ecommerce_data
JWT_SECRET=production-secret-key-change-this
PORT=8080
ENV=production
EOF

# Set permissions
chown -R www:www /www/wwwroot/ecommerce-backend
chmod +x /www/wwwroot/ecommerce-backend

echo "✅ Basic setup complete!"
echo ""
echo "Next steps:"
echo "1. Upload your Go binary to /www/wwwroot/ecommerce-backend/"
echo "2. Configure Nginx in aaPanel"
echo "3. Start the service"
echo ""
echo "To check service status: systemctl status ecommerce-backend"
