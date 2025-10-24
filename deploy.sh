#!/bin/bash

# Ecommerce Backend Deployment Script
# Run this script on your server to set up the application

set -e

echo "🚀 Setting up Ecommerce Backend..."

# Create application directory
sudo mkdir -p /var/www/ecommerce-backend
sudo chown -R www-data:www-data /var/www/ecommerce-backend

# Copy systemd service file
sudo cp ecommerce-backend.service /etc/systemd/system/
sudo systemctl daemon-reload

# Enable the service
sudo systemctl enable ecommerce-backend

# Create environment file if it doesn't exist
if [ ! -f /var/www/ecommerce-backend/.env ]; then
    sudo cp env.example /var/www/ecommerce-backend/.env
    echo "📝 Please update /var/www/ecommerce-backend/.env with your configuration"
fi

# Set proper permissions
sudo chown -R www-data:www-data /var/www/ecommerce-backend
sudo chmod +x /var/www/ecommerce-backend

echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "1. Update /var/www/ecommerce-backend/.env with your configuration"
echo "2. Make sure MongoDB is running"
echo "3. The service will start automatically when the binary is deployed"
echo ""
echo "To check service status: sudo systemctl status ecommerce-backend"
echo "To view logs: sudo journalctl -u ecommerce-backend -f"
