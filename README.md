# Ecommerce Backend

A modern Go-based backend API for ecommerce applications with JWT authentication, role-based access control, and production-ready features.

> Last updated: 2025-01-16 - Repository maintenance

## 🏗️ Project Structure

This project follows modern Go conventions with a clean, scalable architecture:

```
ecommerce-backend/
├── internal/                 # Private application code
│   ├── config/              # Configuration management
│   ├── database/            # Database layer
│   ├── errors/              # Custom error types
│   ├── handlers/            # HTTP handlers
│   ├── logger/              # Structured logging
│   ├── middleware/           # HTTP middleware
│   ├── models/              # Data models
│   ├── server/              # Server setup
│   └── utils/               # Utility functions
├── .github/workflows/       # GitHub Actions
├── main.go                  # Application entry point
├── Makefile                 # Build automation
├── Dockerfile               # Container configuration
└── docker-compose.yml       # Multi-container setup
```

## ✨ Features

- **🔐 JWT Authentication** - Secure token-based authentication
- **👥 Role-Based Access Control** - Admin and user roles
- **📊 Structured Logging** - JSON logging with context
- **🛡️ Graceful Shutdown** - Proper server lifecycle management
- **🧪 Comprehensive Testing** - Unit tests with coverage
- **🐳 Docker Support** - Containerized deployment
- **🚀 Auto-Deployment** - GitHub Actions CI/CD
- **⚡ Performance** - Optimized for production

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- MongoDB 4.4+
- Docker (optional)

### Local Development

1. **Clone and setup:**
   ```bash
   git clone <repository-url>
   cd ecommerce-backend
   make setup
   ```

2. **Configure environment:**
   ```bash
   cp env.example .env
   # Edit .env with your settings
   ```

3. **Run the application:**
   ```bash
   make run
   # or for development with hot reload:
   make dev
   ```

### Using Docker

```bash
# Build and run with Docker Compose
docker-compose up --build

# Or build Docker image
make docker-build
make docker-run
```

## 📚 API Documentation

### Authentication Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/auth/register` | Register new user | ❌ |
| POST | `/api/auth/login` | User login | ❌ |
| GET | `/api/profile` | Get user profile | ✅ |
| GET | `/api/admin/dashboard` | Admin dashboard | ✅ (Admin) |

### Health Check

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Application health status |

## 🔧 Configuration

The application uses environment variables for configuration:

```bash
# Server Configuration
PORT=8080
READ_TIMEOUT=10s
WRITE_TIMEOUT=10s
IDLE_TIMEOUT=120s

# Database Configuration
MONGODB_URI=mongodb://admin:SecureMongoDB123!@localhost:27017
DATABASE_NAME=ecommerce
DB_TIMEOUT=10s
DB_MAX_POOL_SIZE=100

# JWT Configuration
JWT_SECRET=your-secret-key
JWT_EXPIRATION=24h

# Environment
ENV=development  # development or production
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./internal/handlers/...
```

## 🏭 Production Deployment

### Server Setup

1. **Install systemd service:**
   ```bash
   sudo ./deploy.sh
   ```

2. **Configure GitHub Secrets:**
   - `HOST`: Server IP address
   - `USERNAME`: SSH username
   - `SSH_KEY`: Private SSH key
   - `PORT`: SSH port (default: 22)

3. **Nginx Configuration:**
   ```nginx
   location /api/ {
       proxy_pass http://localhost:8080;
       proxy_set_header Host $host;
       proxy_set_header X-Real-IP $remote_addr;
       proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
       proxy_set_header X-Forwarded-Proto $scheme;
   }
   ```

### Auto-Deployment

The GitHub Actions workflow provides:
- ✅ Automated testing on PR/push
- ✅ Security scanning
- ✅ Production deployment on main branch
- ✅ Service restart and health checks

## 📖 Usage Examples

### Register User
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Access Protected Route
```bash
curl -X GET http://localhost:8080/api/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Admin Dashboard
```bash
curl -X GET http://localhost:8080/api/admin/dashboard \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"
```

## 👥 Role Management

### User Roles

- **`user`**: Default role for regular users
- **`admin`**: Administrative access to dashboard

### Making a User Admin

Update the user role in MongoDB:

```javascript
db.users.updateOne(
  { email: "admin@example.com" },
  { $set: { role: "admin" } }
)
```

## 🛠️ Development

### Available Commands

```bash
make build          # Build the application
make test           # Run tests
make test-coverage  # Run tests with coverage
make fmt            # Format code
make lint           # Lint code
make clean          # Clean build artifacts
make dev            # Run with hot reload
```

### Code Quality

This project follows Go best practices:
- ✅ Proper error handling
- ✅ Structured logging
- ✅ Comprehensive testing
- ✅ Clean architecture
- ✅ Security best practices
- ✅ Performance optimization

## 📄 License

This project is licensed under the MIT License.
