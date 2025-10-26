#!/bin/bash

# SSH Key Setup Script
# This script will temporarily enable password auth, add your SSH key, then disable it again

echo "=== SSH KEY SETUP SCRIPT ==="
echo ""

# Your public key
PUBLIC_KEY="ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDaV50jH70CTMO1BhpmIno2Awlyl7UK/u06SHKOdm92HcpgBYgrL/Mi2TsFpsDHC4uhBCMxuPAB7ynZGAl/LiqInsiqEsVaITCGXRm9kVZTvnyqYOkLVpKCjVfDXhAgq9KqzLmPdbBf39NEp7u3vrVdUzGMuuls4Xt3kRniZvUznouGInGZhLTsR3v7b91q9tp7GBQy+OKeRxaGMMQf8/wK28bB2ZPFyTbi3w3QOi/1i7akdyX3oBHHz2ooqn7yjL1vcYcQKrxca8s1//rHA1EpC5nqaz5gJ+jOKOuz+GTEEH10N5BmV9EjgwzHZ9gbv/cTqi0oqPAzksCSiHMhYPisVhcjlvkPYXW543Tcct+FfAVkjftMdIX1EAi0p/NTNlYPYhGHzgfHulhVbolsfvrwJVE8m/7gBgv9rWHM3XBUefXgbz6Utn1WHMA7HJ9TFaPUMmwCWsVnf4Cy87vqfHrslHNHkM9tryYkIPLtyI25Expcm+h2NgrmxXBoGmNqq8sCLWzXazAI9nrFPu0o1JzGM0F62rFOx28FZXZbCMdxf1iHuiWneLJWSZTdSS+09mZ3OhQAL68jVR2tF6Z8NWA7r1srOarqo/WhsK17+diMihBPK6y22IrNDlk6S7TXa7KOahLi7bZGZPULaUYoQ5bvQGYbMNZmsvdbzmjdPg3HtQ== github-actions"

SERVER_IP="130.94.40.85"
SERVER_USER="root"

echo "Step 1: Temporarily enabling password authentication..."
ssh $SERVER_USER@$SERVER_IP "
    echo 'Enabling password auth temporarily...'
    sed -i 's/PasswordAuthentication no/PasswordAuthentication yes/' /etc/ssh/sshd_config
    systemctl restart sshd
    echo 'Password auth enabled temporarily'
"

echo ""
echo "Step 2: Adding your SSH key to the server..."
ssh $SERVER_USER@$SERVER_IP "
    echo 'Creating .ssh directory if it does not exist...'
    mkdir -p ~/.ssh
    chmod 700 ~/.ssh
    
    echo 'Adding your public key to authorized_keys...'
    echo '$PUBLIC_KEY' >> ~/.ssh/authorized_keys
    chmod 600 ~/.ssh/authorized_keys
    
    echo 'SSH key added successfully!'
"

echo ""
echo "Step 3: Testing SSH key authentication..."
ssh -i ~/.ssh/github_actions_key $SERVER_USER@$SERVER_IP "
    echo 'SSH key authentication works!'
    echo 'Current user: \$(whoami)'
    echo 'Current directory: \$(pwd)'
"

echo ""
echo "Step 4: Disabling password authentication again..."
ssh -i ~/.ssh/github_actions_key $SERVER_USER@$SERVER_IP "
    echo 'Disabling password auth again...'
    sed -i 's/PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
    systemctl restart sshd
    echo 'Password auth disabled - server is secure again!'
"

echo ""
echo "=== SSH KEY SETUP COMPLETE ==="
echo "✅ Your SSH key is now configured on the server"
echo "✅ Password authentication is disabled again"
echo "✅ You can now connect using: ssh -i ~/.ssh/github_actions_key root@130.94.40.85"
echo ""
echo "To make it easier, you can add this to your ~/.ssh/config:"
echo "Host myserver"
echo "    HostName 130.94.40.85"
echo "    User root"
echo "    IdentityFile ~/.ssh/github_actions_key"
echo ""
echo "Then you can just use: ssh myserver"
