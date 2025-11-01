# Deployment Fix - Why Backend Turns Off After Deployment

## Issues Found

### 1. **Critical Bug in Deployment Script (Line 74)**
   - **Problem**: The script had `sudo mv /www/wwwroot/ecommerce-backend/ecommerce-backend /www/wwwroot/ecommerce-backend/ecommerce-backend`
   - **Issue**: This tries to move file to itself, which doesn't work and prevents proper binary placement
   - **Fix**: Removed the redundant move command since `scp-action` already uploads to the correct location

### 2. **Service Restart Issues**
   - **Problem**: Using `stop` then `start` can fail if service is already stopped
   - **Fix**: Changed to `restart` which handles both cases better

### 3. **Missing Error Handling**
   - **Problem**: Deployment script doesn't check if service actually started
   - **Fix**: Added service status check with journalctl logs on failure

### 4. **Upload Directory Permissions**
   - **Problem**: Upload directories might not exist or have wrong permissions
   - **Fix**: Added explicit directory creation and permission setting

## Fixes Applied

✅ Fixed deployment script in `.github/workflows/deploy.yml`:
   - Removed buggy file move command
   - Changed `start` to `restart` for better reliability
   - Added service status verification
   - Added health check retries
   - Added upload directory setup
   - Added error logging on failure

✅ Created `ecommerce-backend.service` systemd file:
   - Proper service configuration
   - Environment variables
   - Auto-restart on failure
   - Security settings

## Server Setup Required

### 1. Install Systemd Service (One-time setup on server)

SSH into your server and run:

```bash
# Copy service file to systemd directory
sudo cp /www/wwwroot/ecommerce-backend/ecommerce-backend.service /etc/systemd/system/

# Edit the service file with your production values
sudo nano /etc/systemd/system/ecommerce-backend.service

# Update these environment variables:
# - MONGODB_URI (your actual MongoDB connection string)
# - JWT_SECRET (your production secret)
# - Any other required env vars

# Reload systemd and enable service
sudo systemctl daemon-reload
sudo systemctl enable ecommerce-backend
sudo systemctl start ecommerce-backend

# Check status
sudo systemctl status ecommerce-backend
```

### 2. Verify Service Status

```bash
# Check if service is running
sudo systemctl status ecommerce-backend

# Check logs if service fails
sudo journalctl -u ecommerce-backend -n 50 --no-pager

# Check if binary exists
ls -la /www/wwwroot/ecommerce-backend/ecommerce-backend
```

### 3. Test Health Endpoint

```bash
curl http://localhost:8080/health
```

## Common Issues

### Service Won't Start
1. Check logs: `sudo journalctl -u ecommerce-backend -n 50`
2. Verify binary exists and is executable: `ls -la /www/wwwroot/ecommerce-backend/ecommerce-backend`
3. Check MongoDB is running: `sudo systemctl status mongod`
4. Verify environment variables in service file

### 502 Bad Gateway
- Usually means nginx can't reach backend on port 8080
- Check if backend is listening: `sudo netstat -tlnp | grep 8080`
- Verify nginx proxy_pass points to `http://localhost:8080`

### Binary Missing After Deployment
- The fixed deployment script now verifies binary exists before starting service
- Check GitHub Actions logs to see if upload succeeded

## Next Deployment

After merging this fix, the next deployment should:
1. ✅ Properly place the binary
2. ✅ Create upload directories with correct permissions
3. ✅ Restart service reliably
4. ✅ Verify service started successfully
5. ✅ Retry health checks before reporting success
6. ✅ Show logs if service fails to start
