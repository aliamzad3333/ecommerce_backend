#!/bin/bash
# Script to update nginx configuration to proxy /uploads/ requests
# This ensures images are always accessible after deployment

NGINX_CONFIG="/www/server/panel/vhost/nginx/ecommerce.conf"

# Backup existing config
if [ -f "$NGINX_CONFIG" ]; then
    sudo cp "$NGINX_CONFIG" "${NGINX_CONFIG}.backup.$(date +%Y%m%d_%H%M%S)" || true
fi

# Check if /uploads/ location already exists
if sudo grep -q "location.*^~.*/uploads/" "$NGINX_CONFIG" 2>/dev/null; then
    echo "✅ Nginx /uploads/ location already configured"
    exit 0
fi

# Create temp file with the uploads location block
TEMP_FILE=$(mktemp)
cat > "$TEMP_FILE" << 'EOF'
    # Uploads - proxy to Go backend (^~ gives this priority over regex)
    # This ensures images are always accessible after deployment
    location ^~ /uploads/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # CORS headers
        add_header Access-Control-Allow-Origin *;
        add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
        add_header Access-Control-Allow-Headers "Content-Type, Authorization";
        
        # Cache images for 1 year
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
    
EOF

# Insert before /api/ location
sudo sed -i "/location \/api\//r $TEMP_FILE" "$NGINX_CONFIG"
rm -f "$TEMP_FILE"

# Test and reload nginx
if sudo nginx -t 2>/dev/null; then
    sudo systemctl reload nginx
    echo "✅ Nginx configuration updated and reloaded"
    exit 0
else
    echo "⚠️ WARNING: Nginx config test failed, restoring backup..."
    BACKUP_FILE=$(ls -t "${NGINX_CONFIG}.backup."* 2>/dev/null | head -1)
    if [ -n "$BACKUP_FILE" ]; then
        sudo cp "$BACKUP_FILE" "$NGINX_CONFIG" || true
        sudo nginx -t && sudo systemctl reload nginx
    fi
    exit 1
fi

