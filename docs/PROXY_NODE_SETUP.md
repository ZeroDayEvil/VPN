# Proxy Node Setup Guide

## Overview

The proxy node is the server that handles the actual VPN traffic. Users connect to this server through the client application.

## Quick Setup Options

### Option 1: 3proxy (Recommended for MVP)

3proxy is a lightweight, cross-platform proxy server.

#### Installation (Ubuntu/Debian)

```bash
# Install 3proxy
sudo apt-get update
sudo apt-get install -y 3proxy

# Create configuration
sudo nano /etc/3proxy/3proxy.cfg
```

#### Configuration

```
# 3proxy.cfg
daemon
pidfile /var/run/3proxy.pid
nserver 8.8.8.8
nscache 65536
timeouts 1 5 30 60 180 1800 15 60

# Authentication
users username:CL:password

# Access control
auth strong

# HTTP proxy on port 3128
proxy -p3128 -a
# SOCKS5 on port 1080
socks -p1080 -a

# Allow all
allow *
```

#### Start Service

```bash
sudo systemctl enable 3proxy
sudo systemctl start 3proxy
sudo systemctl status 3proxy
```

### Option 2: Squid Proxy

Squid is a powerful caching proxy server.

#### Installation (Ubuntu/Debian)

```bash
sudo apt-get update
sudo apt-get install -y squid

# Backup default config
sudo cp /etc/squid/squid.conf /etc/squid/squid.conf.backup

# Edit configuration
sudo nano /etc/squid/squid.conf
```

#### Configuration

```
# Port configuration
http_port 3128

# Authentication
auth_param basic program /usr/lib/squid/basic_ncsa_auth /etc/squid/passwords
auth_param basic children 5
auth_param basic realm Proxy Authentication Required
auth_param basic credentialsttl 2 hours

acl authenticated proxy_auth REQUIRED
http_access allow authenticated

# Deny all other access
http_access deny all

# Disable cache for VPN use
cache deny all

# Error page
error_directory /usr/share/squid/errors/en

# Logging
access_log /var/log/squid/access.log squid
cache_log /var/log/squid/cache.log
```

#### Create user

```bash
# Install apache2-utils for htpasswd
sudo apt-get install -y apache2-utils

# Create password file
sudo htpasswd -c /etc/squid/passwords username

# Restart Squid
sudo systemctl restart squid
sudo systemctl status squid
```

### Option 3: Custom Go Proxy Gateway

For more control, create a custom proxy gateway.

#### Create proxy.go

```go
package main

import (
    "io"
    "log"
    "net"
    "net/http"
)

func handleTunneling(w http.ResponseWriter, r *http.Request) {
    destConn, err := net.Dial("tcp", r.Host)
    if err != nil {
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    hijacker, ok := w.(http.Hijacker)
    if !ok {
        http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
        return
    }
    
    clientConn, _, err := hijacker.Hijack()
    if err != nil {
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
        return
    }
    
    go transfer(destConn, clientConn)
    go transfer(clientConn, destConn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
    defer destination.Close()
    defer source.Close()
    io.Copy(destination, source)
}

func handleHTTP(w http.ResponseWriter, req *http.Request) {
    resp, err := http.DefaultTransport.RoundTrip(req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
        return
    }
    defer resp.Body.Close()
    
    copyHeader(w.Header(), resp.Header)
    w.WriteHeader(resp.StatusCode)
    io.Copy(w, resp.Body)
}

func copyHeader(dst, src http.Header) {
    for k, vv := range src {
        for _, v := range vv {
            dst.Add(k, v)
        }
    }
}

func main() {
    server := &http.Server{
        Addr: ":8080",
        Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if r.Method == http.MethodConnect {
                handleTunneling(w, r)
            } else {
                handleHTTP(w, r)
            }
        }),
    }
    
    log.Println("Proxy server starting on :8080")
    log.Fatal(server.ListenAndServe())
}
```

#### Build and run

```bash
go build proxy.go
./proxy
```

## Security Hardening

### 1. Use HTTPS/TLS

Add TLS termination using nginx or use built-in TLS in proxy server.

#### Nginx as TLS frontend

```nginx
server {
    listen 443 ssl http2;
    server_name proxy.example.com;

    ssl_certificate /etc/letsencrypt/live/proxy.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/proxy.example.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:3128;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

### 2. Firewall Configuration

```bash
# Allow only necessary ports
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow 22/tcp   # SSH
sudo ufw allow 443/tcp  # HTTPS
sudo ufw enable
```

### 3. Rate Limiting

Add rate limiting to prevent abuse:

```bash
# Using iptables
sudo iptables -A INPUT -p tcp --dport 443 -m state --state NEW -m recent --set
sudo iptables -A INPUT -p tcp --dport 443 -m state --state NEW -m recent --update --seconds 60 --hitcount 20 -j DROP
```

## Integration with Backend API

The backend API should:

1. Manage proxy node endpoints
2. Generate authentication tokens
3. Assign clients to least-loaded nodes
4. Monitor node health

### Database Schema

Already included in `backend/main.py`:

```sql
CREATE TABLE proxy_nodes (
    id INTEGER PRIMARY KEY,
    endpoint TEXT NOT NULL,
    protocol TEXT NOT NULL,
    region TEXT,
    active BOOLEAN DEFAULT 1
);
```

### Add Proxy Nodes via API

```bash
# Add new proxy node (requires admin authentication)
curl -X POST http://localhost:8000/admin/proxy-nodes \
  -H "Content-Type: application/json" \
  -d '{
    "endpoint": "proxy1.example.com:443",
    "protocol": "https",
    "region": "us-east"
  }'
```

## Monitoring

### Check proxy logs

```bash
# 3proxy
sudo tail -f /var/log/3proxy.log

# Squid
sudo tail -f /var/log/squid/access.log

# Custom Go proxy
# Add logging to your application
```

### Monitor connections

```bash
# Check active connections
sudo netstat -an | grep :3128

# Check bandwidth usage
sudo iftop -i eth0
```

## Scaling

### Multiple Proxy Nodes

1. Set up multiple servers with proxy
2. Add all endpoints to backend database
3. Backend API distributes clients across nodes
4. Use load balancer (optional)

### Geographic Distribution

1. Deploy nodes in different regions
2. Tag nodes with region in database
3. Assign clients to closest region
4. Implement failover to other regions

## Testing

### Test proxy locally

```bash
# Test HTTP
curl -x http://username:password@localhost:3128 http://example.com

# Test HTTPS
curl -x http://username:password@localhost:3128 https://example.com

# Check IP
curl -x http://username:password@localhost:3128 https://api.ipify.org
```

### Test from client

```bash
# Use the VPN client to connect
# Then check your IP
curl https://api.ipify.org
```

## Troubleshooting

### Proxy not starting

1. Check if port is already in use: `sudo netstat -tulpn | grep 3128`
2. Check logs for errors
3. Verify configuration syntax

### Authentication failing

1. Verify username/password file exists
2. Check file permissions
3. Test with known good credentials

### Performance issues

1. Check server resources (CPU, RAM, bandwidth)
2. Increase proxy cache size if applicable
3. Add more proxy nodes
4. Implement rate limiting

## Cost Considerations

### Cloud Providers

- **AWS EC2**: t2.micro ($8-10/month)
- **DigitalOcean**: Basic Droplet ($4-6/month)
- **Vultr**: Cloud Compute ($3.50-6/month)
- **Linode**: Nanode ($5/month)

### Bandwidth

- Monitor bandwidth usage carefully
- Most providers charge for excess bandwidth
- Implement per-user quotas if needed

## Backup Configuration

```bash
# Backup proxy config
sudo cp /etc/3proxy/3proxy.cfg /etc/3proxy/3proxy.cfg.backup
sudo cp /etc/squid/squid.conf /etc/squid/squid.conf.backup

# Backup authentication files
sudo cp /etc/squid/passwords /etc/squid/passwords.backup
```
