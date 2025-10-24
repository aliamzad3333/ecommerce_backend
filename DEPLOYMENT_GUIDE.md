# 🚀 Complete Deployment Guide for Ecommerce Backend

## 📋 **Prerequisites**
- Server with aaPanel installed
- SSH access to your server
- GitHub repository with the code

## 🗄️ **Step 1: MongoDB Setup**

### **Connect to your server:**
```bash
ssh your-username@130.94.40.85
```

### **Run MongoDB setup:**
```bash
# Download and run MongoDB setup
curl -O https://raw.githubusercontent.com/your-repo/ecommerce-backend/main/setup-mongodb.sh
chmod +x setup-mongodb.sh
sudo ./setup-mongodb.sh
```

This will:
- Install MongoDB (if not already installed)
- Create the `Ecommerce_data` database
- Create collections: `users`, `products`, `orders`, `categories`
- Add sample data and admin user
- Set up proper indexes

## ⚙️ **Step 2: Backend Setup**

### **Run the backend setup:**
```bash
# Download and run backend setup
curl -O https://raw.githubusercontent.com/your-repo/ecommerce-backend/main/quick-deploy.sh
chmod +x quick-deploy.sh
sudo ./quick-deploy.sh
```

This will:
- Create `/www/wwwroot/ecommerce-backend/` directory
- Set up systemd service
- Create environment file with `Ecommerce_data` database
- Set proper permissions

## 🌐 **Step 3: Nginx Configuration in aaPanel**

### **Login to aaPanel:**
1. Go to `http://130.94.40.85:8888`
2. Login with your aaPanel credentials

### **Configure Nginx:**
1. Go to **Website** > **Your Site** > **Settings**
2. Click on **Configuration Files**
3. Add this configuration:

```nginx
# API routes - proxy to Go backend
location /api/ {
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    
    # CORS headers
    add_header Access-Control-Allow-Origin *;
    add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
    add_header Access-Control-Allow-Headers "Content-Type, Authorization";
    
    # Handle preflight requests
    if ($request_method = 'OPTIONS') {
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
    proxy_pass http://127.0.0.1:8080/health;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

4. **Save** and **Reload Nginx**

## 🔧 **Step 4: GitHub Actions Setup**

### **Set up GitHub Secrets:**
1. Go to your GitHub repository
2. **Settings** > **Secrets and variables** > **Actions**
3. Add these secrets:
   - `HOST`: `130.94.40.85`
   - `USERNAME`: Your SSH username
   - `SSH_KEY`: Your private SSH key
   - `PORT`: `22`

### **Deploy via GitHub Actions:**
```bash
# Push to trigger deployment
git add .
git commit -m "Deploy to production with Ecommerce_data database"
git push origin main
```

## 🧪 **Step 5: Testing**

### **Test the deployment:**
```bash
# Health check
curl http://130.94.40.85/health

# Register user
curl -X POST http://130.94.40.85/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }'

# Login
curl -X POST http://130.94.40.85/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

## 📊 **Database Structure**

Your `Ecommerce_data` database will have:

### **Collections:**
- **`users`** - User accounts with authentication
- **`products`** - Product catalog
- **`orders`** - Order management
- **`categories`** - Product categories

### **Sample Data:**
- Admin user: `admin@ecommerce.com` / `password`
- Sample categories: Electronics, Clothing, Books, Home & Garden

## 🔍 **Monitoring**

### **Check services:**
```bash
# Backend service
sudo systemctl status ecommerce-backend

# MongoDB service
sudo systemctl status mongod

# Nginx service
sudo systemctl status nginx
```

### **View logs:**
```bash
# Backend logs
sudo journalctl -u ecommerce-backend -f

# MongoDB logs
sudo tail -f /var/log/mongodb/mongod.log
```

## 🎯 **API Endpoints**

Your API will be available at:
- **Health**: `http://130.94.40.85/health`
- **Register**: `http://130.94.40.85/api/auth/register`
- **Login**: `http://130.94.40.85/api/auth/login`
- **Profile**: `http://130.94.40.85/api/profile`
- **Admin Dashboard**: `http://130.94.40.85/api/admin/dashboard`

## 🚨 **Troubleshooting**

### **If backend won't start:**
```bash
# Check logs
sudo journalctl -u ecommerce-backend -f

# Restart service
sudo systemctl restart ecommerce-backend
```

### **If MongoDB connection fails:**
```bash
# Check MongoDB status
sudo systemctl status mongod

# Test connection
mongosh Ecommerce_data
```

### **If Nginx proxy fails:**
```bash
# Test Nginx configuration
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx
```

## ✅ **Success Indicators**

You'll know the deployment is successful when:
- ✅ Health check returns: `{"status":"ok","timestamp":"..."}`
- ✅ User registration works
- ✅ User login returns JWT token
- ✅ Protected routes work with JWT
- ✅ Admin dashboard accessible with admin role
