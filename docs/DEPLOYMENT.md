# Deployment Guide

## Prerequisites

Before deploying, you need:

1. A server for the Backend API (VPS with Ubuntu 20.04+ recommended)
2. A server for the Proxy Node (can be same or different server)
3. A domain name (for API and optionally for proxy)
4. SSL certificates (Let's Encrypt recommended)
5. GitHub account for hosting config files

## Step 1: Deploy Backend API

### 1.1 Prepare Server

```bash
# SSH into your server
ssh user@your-api-server.com

# Update system
sudo apt-get update
sudo apt-get upgrade -y

# Install Python 3 and pip
sudo apt-get install -y python3 python3-pip python3-venv

# Install nginx for reverse proxy
sudo apt-get install -y nginx
```

### 1.2 Set Up Application

```bash
# Create application directory
sudo mkdir -p /opt/vpn-api
cd /opt/vpn-api

# Clone or upload your backend code
# For this guide, we'll create the files manually

# Upload main.py and requirements.txt
sudo nano /opt/vpn-api/main.py
# (paste backend/main.py content)

sudo nano /opt/vpn-api/requirements.txt
# (paste backend/requirements.txt content)

# Create virtual environment
sudo python3 -m venv venv
sudo chown -R $USER:$USER /opt/vpn-api

# Activate and install dependencies
source venv/bin/activate
pip install -r requirements.txt
```

### 1.3 Create Systemd Service

```bash
sudo nano /etc/systemd/system/vpn-api.service
```

Content:
```ini
[Unit]
Description=VPN API Service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/vpn-api
Environment="PATH=/opt/vpn-api/venv/bin"
Environment="DB_PATH=/var/lib/vpn/vpn_client.db"
ExecStart=/opt/vpn-api/venv/bin/uvicorn main:app --host 127.0.0.1 --port 8000
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```

```bash
# Create database directory
sudo mkdir -p /var/lib/vpn
sudo chown www-data:www-data /var/lib/vpn

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable vpn-api
sudo systemctl start vpn-api
sudo systemctl status vpn-api
```

### 1.4 Configure Nginx

```bash
sudo nano /etc/nginx/sites-available/vpn-api
```

Content:
```nginx
server {
    listen 80;
    server_name api.yourdomain.com;

    location / {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/vpn-api /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 1.5 Install SSL Certificate

```bash
# Install certbot
sudo apt-get install -y certbot python3-certbot-nginx

# Get certificate
sudo certbot --nginx -d api.yourdomain.com

# Test auto-renewal
sudo certbot renew --dry-run
```

### 1.6 Verify API

```bash
# Test API
curl https://api.yourdomain.com/

# Should return:
# {"name":"VPN One-Click API","version":"1.0.0","status":"running"}
```

## Step 2: Deploy Proxy Node

### 2.1 Set Up 3proxy

```bash
# SSH into proxy server (can be same as API server)
ssh user@your-proxy-server.com

# Install 3proxy
sudo apt-get update
sudo apt-get install -y 3proxy

# Create configuration
sudo nano /etc/3proxy/3proxy.cfg
```

Content:
```
daemon
pidfile /var/run/3proxy.pid
nserver 8.8.8.8
nscache 65536
timeouts 1 5 30 60 180 1800 15 60

# Log
log /var/log/3proxy.log D
logformat "- +_L%t.%. %N.%p %E %U %C:%c %R:%r %O %I %h %T"

# Authentication (change username and password)
users vpnuser:CL:VPNPassword123

# Access control
auth strong

# HTTP proxy
proxy -p3128 -a

# Allow all
allow *
```

```bash
# Set permissions
sudo chown proxy:proxy /etc/3proxy/3proxy.cfg
sudo chmod 600 /etc/3proxy/3proxy.cfg

# Enable and start
sudo systemctl enable 3proxy
sudo systemctl start 3proxy
sudo systemctl status 3proxy

# Check logs
sudo tail -f /var/log/3proxy.log
```

### 2.2 Configure Firewall

```bash
# Allow proxy port
sudo ufw allow 3128/tcp

# Allow SSH (if not already allowed)
sudo ufw allow 22/tcp

# Enable firewall
sudo ufw enable
```

### 2.3 Test Proxy

```bash
# From another machine
curl -x http://vpnuser:VPNPassword123@your-proxy-server.com:3128 http://api.ipify.org

# Should return proxy server's IP
```

## Step 3: Configure GitHub

### 3.1 Update Config Files

Update `config/config.json`:
```json
{
  "config_version": "1.0.0",
  "min_client_version": "1.0.0",
  "api": {
    "base_url": "https://api.yourdomain.com",
    "enroll_path": "/v1/enroll",
    "heartbeat_path": "/v1/heartbeat",
    "commands_path": "/v1/commands",
    "command_result_path": "/v1/command_result"
  },
  "client": {
    "heartbeat_interval_sec": 30,
    "commands_interval_sec": 10,
    "log_level": "info",
    "local_http_port": 18080,
    "local_socks_port": 18081,
    "enable_system_proxy": true,
    "use_pac": false,
    "proxy_bypass": "localhost;127.0.0.1;<local>"
  },
  "update": {
    "manifest_url": "https://raw.githubusercontent.com/YOUR-ORG/VPN/main/config/manifest.json"
  }
}
```

### 3.2 Update Manifest

```bash
# Calculate SHA256
cd config
sha256sum config.json

# Update manifest.json with new SHA256
```

### 3.3 Commit and Push

```bash
git add config/
git commit -m "Update configuration for production"
git push origin main
```

## Step 4: Build and Distribute Client

### 4.1 Build Windows Client

On a machine with Go installed:

```bash
cd client

# Build for Windows
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o Client.exe

# Optional: Sign the executable (recommended for production)
# signtool sign /f certificate.pfx /p password /t http://timestamp.digicert.com Client.exe
```

### 4.2 Distribute Client

Options:
1. Host on your website for download
2. Distribute via email
3. Use GitHub Releases
4. Use a file sharing service

### 4.3 Update Proxy Profile in Backend

Update the proxy endpoint in `backend/main.py`:

```python
# In enroll() function
proxy_profile = ProxyProfile(
    node_endpoint="your-proxy-server.com:3128",  # Update this
    protocol="http",
    auth={
        "type": "basic",
        "username": "vpnuser",      # Update this
        "password": "VPNPassword123"  # Update this
    },
    expires_at=(datetime.datetime.utcnow() + datetime.timedelta(hours=24)).isoformat() + "Z"
)
```

Restart API:
```bash
sudo systemctl restart vpn-api
```

## Step 5: Testing End-to-End

### 5.1 Test from Windows Client

1. Download Client.exe
2. Run it
3. Check logs in `%APPDATA%\VPNCompany\OneClickClient\logs\`
4. Verify enrollment: Check `http://api.yourdomain.com/admin/devices`
5. Test internet: Visit https://api.ipify.org

### 5.2 Verify Connection

```bash
# On API server, check devices
curl https://api.yourdomain.com/admin/devices

# Should show your device with "connected": true
```

## Step 6: Monitoring

### 6.1 Monitor API

```bash
# Check API logs
sudo journalctl -u vpn-api -f

# Check nginx logs
sudo tail -f /var/log/nginx/access.log
```

### 6.2 Monitor Proxy

```bash
# Check proxy logs
sudo tail -f /var/log/3proxy.log

# Check connections
sudo netstat -an | grep :3128
```

### 6.3 Monitor Resources

```bash
# Install htop
sudo apt-get install -y htop

# Monitor
htop

# Check disk space
df -h
```

## Security Checklist

- [ ] API uses HTTPS only
- [ ] Proxy credentials are strong
- [ ] Firewall is configured
- [ ] SSL certificates are valid
- [ ] Admin endpoints are protected
- [ ] Database backups are configured
- [ ] Logs are rotated
- [ ] System is up to date

## Backup Strategy

### Database Backup

```bash
# Create backup script
sudo nano /opt/vpn-api/backup.sh
```

Content:
```bash
#!/bin/bash
BACKUP_DIR="/var/backups/vpn"
DB_PATH="/var/lib/vpn/vpn_client.db"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR
cp $DB_PATH $BACKUP_DIR/vpn_client_$DATE.db

# Keep only last 7 days
find $BACKUP_DIR -name "vpn_client_*.db" -mtime +7 -delete
```

```bash
# Make executable
sudo chmod +x /opt/vpn-api/backup.sh

# Add to cron (daily at 2 AM)
sudo crontab -e
# Add: 0 2 * * * /opt/vpn-api/backup.sh
```

## Troubleshooting

### API not starting

1. Check logs: `sudo journalctl -u vpn-api -n 50`
2. Check if port 8000 is available: `sudo netstat -tulpn | grep 8000`
3. Check permissions on database directory

### Proxy not working

1. Check 3proxy status: `sudo systemctl status 3proxy`
2. Check logs: `sudo tail -f /var/log/3proxy.log`
3. Verify firewall: `sudo ufw status`
4. Test locally: `curl -x http://localhost:3128 http://example.com`

### Client can't connect

1. Check config.json has correct API URL
2. Verify API is accessible: `curl https://api.yourdomain.com/`
3. Check client logs in %APPDATA%
4. Verify firewall rules

## Maintenance

### Update Backend

```bash
# SSH to server
cd /opt/vpn-api

# Pull latest code
git pull

# Restart service
sudo systemctl restart vpn-api
```

### Update Config

```bash
# Update config files in GitHub
# Clients will auto-download on next run or restart
```

### Update Client

Build new Client.exe and distribute to users.

## Scaling

### Add More Proxy Nodes

1. Set up additional proxy servers
2. Add to database:
```bash
curl -X POST https://api.yourdomain.com/admin/proxy-nodes \
  -H "Content-Type: application/json" \
  -d '{"endpoint":"proxy2.yourdomain.com:3128","protocol":"http","region":"us-west"}'
```
3. Backend will distribute clients across nodes

### Scale API

1. Use a proper database (PostgreSQL)
2. Use Redis for caching
3. Add load balancer
4. Deploy multiple API instances
