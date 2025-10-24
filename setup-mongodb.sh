#!/bin/bash

# MongoDB Setup Script for Ecommerce Backend
# This script sets up MongoDB and creates the Ecommerce_data database

set -e

echo "🗄️ Setting up MongoDB for Ecommerce Backend..."

# Check if MongoDB is installed
if ! command -v mongod &> /dev/null; then
    echo "📦 Installing MongoDB..."
    
    # Update package list
    sudo apt update
    
    # Install MongoDB
    wget -qO - https://www.mongodb.org/static/pgp/server-6.0.asc | sudo apt-key add -
    echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/ubuntu focal/mongodb-org/6.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-6.0.list
    sudo apt update
    sudo apt install -y mongodb-org
    
    # Start and enable MongoDB
    sudo systemctl start mongod
    sudo systemctl enable mongod
else
    echo "✅ MongoDB is already installed"
fi

# Start MongoDB service
echo "🔄 Starting MongoDB service..."
sudo systemctl start mongod
sudo systemctl enable mongod

# Wait for MongoDB to start
sleep 5

# Create the Ecommerce_data database and initial collections
echo "📊 Creating Ecommerce_data database and collections..."

mongosh --eval "
use Ecommerce_data;

// Create users collection with indexes
db.createCollection('users');
db.users.createIndex({ email: 1 }, { unique: true });
db.users.createIndex({ role: 1 });

// Create products collection (for future use)
db.createCollection('products');
db.products.createIndex({ name: 1 });
db.products.createIndex({ category: 1 });
db.products.createIndex({ price: 1 });

// Create orders collection (for future use)
db.createCollection('orders');
db.orders.createIndex({ user_id: 1 });
db.orders.createIndex({ status: 1 });
db.orders.createIndex({ created_at: -1 });

// Create categories collection (for future use)
db.createCollection('categories');
db.categories.createIndex({ name: 1 });

// Insert sample admin user
db.users.insertOne({
    email: 'admin@ecommerce.com',
    password: '\$2a\$10\$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', // password
    first_name: 'Admin',
    last_name: 'User',
    role: 'admin',
    is_active: true,
    created_at: new Date(),
    updated_at: new Date()
});

// Insert sample categories
db.categories.insertMany([
    { name: 'Electronics', description: 'Electronic devices and gadgets', created_at: new Date() },
    { name: 'Clothing', description: 'Fashion and apparel', created_at: new Date() },
    { name: 'Books', description: 'Books and literature', created_at: new Date() },
    { name: 'Home & Garden', description: 'Home improvement and garden supplies', created_at: new Date() }
]);

print('✅ Ecommerce_data database created successfully!');
print('📊 Collections created: users, products, orders, categories');
print('👤 Admin user created: admin@ecommerce.com / password');
print('📋 Sample categories added');
"

# Check MongoDB status
echo "🔍 Checking MongoDB status..."
sudo systemctl status mongod --no-pager

# Test database connection
echo "🧪 Testing database connection..."
mongosh Ecommerce_data --eval "db.stats()"

echo ""
echo "✅ MongoDB setup complete!"
echo ""
echo "Database: Ecommerce_data"
echo "Collections: users, products, orders, categories"
echo "Admin user: admin@ecommerce.com / password"
echo ""
echo "To connect to MongoDB:"
echo "mongosh Ecommerce_data"
echo ""
echo "To check MongoDB status:"
echo "sudo systemctl status mongod"
