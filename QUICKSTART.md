# Quick Start Guide

## For Testing Locally (Right Now)

### Step 1: Start the Backend API

```bash
cd backend
python3 main.py
```

You should see:
```
INFO:     Started server process
INFO:     Waiting for application startup.
INFO:     Application startup complete.
INFO:     Uvicorn running on http://0.0.0.0:8000
```

Leave this running.

### Step 2: Test the API

In a new terminal:
```bash
# Test API is running
curl http://localhost:8000/

# Should return:
# {"name":"VPN One-Click API","version":"1.0.0","status":"running"}
```

### Step 3: Run the Client

In another terminal:
```bash
cd client
go run main.go -no-ui
```

You should see:
```
Starting One-Click Client v1.0.0
Device not enrolled, performing enrollment...
Enrollment successful
Connecting...
HTTP proxy listening on 127.0.0.1:18080
SOCKS5 proxy listening on 127.0.0.1:18081
Connected successfully
Running in console mode...
Press Ctrl+C to exit
```

### Step 4: Test the Proxy

In another terminal:
```bash
# Test HTTP proxy
curl -x http://127.0.0.1:18080 http://httpbin.org/ip

# Test SOCKS5 proxy
curl --socks5 127.0.0.1:18081 http://httpbin.org/ip
```

### Step 5: Check Device Status

```bash
# View enrolled devices
curl http://localhost:8000/admin/devices
```

### Step 6: Send a Command

```bash
# Send disconnect command
curl -X POST "http://localhost:8000/admin/commands?device_id=YOUR_DEVICE_ID&command_type=DISCONNECT"

# The client should disconnect and show:
# Executing command: DISCONNECT
# Disconnecting...
```

## For Production Deployment

See `docs/DEPLOYMENT.md` for complete production deployment guide.

## Troubleshooting

### Backend won't start
- Check if port 8000 is already in use: `lsof -i :8000`
- Install dependencies: `pip3 install -r requirements.txt`

### Client won't start
- Check backend is running on localhost:8000
- Check logs (created in data directory)
- Verify Go is installed: `go version`

### Proxy not working
- Check if ports 18080 and 18081 are available
- Verify client shows "Connected successfully"
- Check firewall settings

## What's Working

✅ Client enrollment
✅ Device registration
✅ Local HTTP proxy on port 18080
✅ Local SOCKS5 proxy on port 18081
✅ Heartbeat monitoring
✅ Command distribution
✅ Config auto-update from GitHub
✅ Logging with rotation

## Next Steps

1. ✅ Local testing (you're here)
2. 📖 Read `docs/DEPLOYMENT.md` for production setup
3. 🌐 Set up your server and domain
4. 🚀 Deploy to production
5. 📦 Distribute Client.exe to users

---

**Need help?** Check the documentation:
- `README.md` - Project overview
- `docs/TECHNICAL.md` - Technical details
- `docs/DEPLOYMENT.md` - Production deployment
- `REQUIREMENTS_FROM_USER.md` - What you need to provide
