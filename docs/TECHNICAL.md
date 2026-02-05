# One-Click VPN Client - Technical Documentation

## Project Overview

This is a complete implementation of the "One-Click Client" for Windows that provides proxy-based VPN access without requiring administrator privileges.

## Architecture

The project consists of three main components:

### 1. Client Application (Go)
- **Location**: `/client`
- **Output**: Single executable `client.exe` (for Windows)
- **Features**:
  - Auto-updates configuration from GitHub
  - Device registration and token management
  - Local HTTP and SOCKS5 proxy servers
  - Windows system proxy configuration
  - API communication (heartbeat, commands)
  - Logging with rotation
  - Graceful shutdown with settings restoration

### 2. Backend API (Python FastAPI)
- **Location**: `/backend`
- **Technology**: FastAPI with SQLite
- **Features**:
  - Device enrollment and management
  - Heartbeat monitoring
  - Command distribution
  - Proxy node management
  - Admin endpoints for testing

### 3. Configuration Files (GitHub)
- **Location**: `/config`
- **Files**:
  - `manifest.json` - Version and SHA256 tracking
  - `config.json` - Application configuration

## Directory Structure

```
VPN/
├── client/                 # Go client application
│   ├── main.go            # Main application entry
│   ├── internal/          # Internal packages
│   │   ├── config/        # Configuration management
│   │   ├── logger/        # Logging system
│   │   ├── proxy/         # Proxy server implementation
│   │   ├── api/           # API client
│   │   ├── sysproxy/      # System proxy management
│   │   └── ui/            # User interface
│   ├── go.mod
│   └── go.sum
├── backend/               # Python FastAPI backend
│   ├── main.py           # API server
│   └── requirements.txt  # Python dependencies
├── config/               # GitHub-hosted config
│   ├── manifest.json    # Version manifest
│   └── config.json      # Application config
├── docs/                # Documentation
└── proxy-node/          # Proxy node setup
```

## Building

### Client (Windows)

To build for Windows from Linux/Mac:

```bash
cd client
GOOS=windows GOARCH=amd64 go build -o Client.exe
```

To build for current platform:

```bash
cd client
go build -o client
```

### Backend

```bash
cd backend
pip3 install -r requirements.txt
```

## Running

### Backend API

```bash
cd backend
python3 main.py
```

The API will start on `http://localhost:8000`

API Documentation: `http://localhost:8000/docs`

### Client

```bash
cd client
./client
```

Or on Windows:
```cmd
Client.exe
```

For console mode without UI:
```bash
./client -no-ui
```

## Configuration

### Update API URL

Edit `config/config.json`:

```json
{
  "api": {
    "base_url": "https://your-api-domain.com"
  }
}
```

### Update GitHub Config URL

Edit `config/config.json`:

```json
{
  "update": {
    "manifest_url": "https://raw.githubusercontent.com/YOUR-ORG/YOUR-REPO/main/config/manifest.json"
  }
}
```

## Testing

### Test Backend API

1. Start the backend:
```bash
cd backend
python3 main.py
```

2. Test enrollment:
```bash
curl -X POST http://localhost:8000/v1/enroll \
  -H "Content-Type: application/json" \
  -d '{"device_id":"test-device","client_version":"1.0.0"}'
```

3. View devices:
```bash
curl http://localhost:8000/admin/devices
```

### Test Client

1. Update `config/config.json` with your API URL
2. Run the client:
```bash
cd client
./client -no-ui
```

3. Check logs in `%APPDATA%/VPNCompany/OneClickClient/logs/`

## Deployment

### Backend Deployment

1. Set up a server with Python 3.8+
2. Install dependencies:
```bash
pip3 install -r requirements.txt
```

3. Set environment variables:
```bash
export DB_PATH=/var/lib/vpn/vpn_client.db
```

4. Run with production server:
```bash
uvicorn main:app --host 0.0.0.0 --port 8000
```

Or use systemd service.

### Client Distribution

1. Build for Windows:
```bash
cd client
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o Client.exe
```

2. Distribute `Client.exe` to users
3. Users run the exe without installation

### Config Distribution

1. Commit `config/manifest.json` and `config/config.json` to GitHub
2. Update `base_url` in config.json to your API server
3. Client will auto-download from GitHub on first run

## Security Notes

### MVP Security Considerations

1. **HTTPS Required**: All API communication must use HTTPS in production
2. **Token Storage**: Device tokens stored in `%APPDATA%` (user-accessible)
3. **Config Verification**: SHA256 check prevents tampering
4. **No Secrets in Logs**: Sensitive data is masked in logs

### Production Security Recommendations

1. Implement config signing (Ed25519)
2. Use certificate pinning
3. Implement token rotation
4. Add rate limiting to API
5. Use encrypted storage for tokens
6. Implement proper authentication for admin endpoints

## API Endpoints

### Client Endpoints

- `POST /v1/enroll` - Register new device
- `POST /v1/heartbeat` - Send status update
- `GET /v1/commands` - Get pending commands
- `POST /v1/command_result` - Report command execution

### Admin Endpoints (Testing Only)

- `GET /admin/devices` - List all devices
- `POST /admin/commands` - Create command for device

## Commands

Supported commands:
- `CONNECT` - Connect to proxy
- `DISCONNECT` - Disconnect from proxy
- `FORCE_CONFIG_REFRESH` - Refresh configuration from GitHub

## Troubleshooting

### Client won't start

1. Check logs in `%APPDATA%/VPNCompany/OneClickClient/logs/`
2. Verify config.json is valid JSON
3. Check API connectivity

### Proxy not working

1. Verify system proxy settings in Windows
2. Check if local ports (18080, 18081) are available
3. Test with curl: `curl -x http://127.0.0.1:18080 http://example.com`

### Backend database errors

1. Check DB_PATH permissions
2. Verify SQLite is available
3. Delete database and restart to reinitialize

## Limitations (MVP)

1. **Not a system VPN**: Only apps that respect system proxy will work
2. **No admin privileges**: Cannot install drivers or system services
3. **Windows only**: Client is designed for Windows (can be adapted for Linux/Mac)
4. **No kill-switch**: No protection if connection drops
5. **No split tunneling**: All or nothing proxy configuration

## Future Enhancements

1. Full VPN mode with WireGuard (requires admin)
2. Always-on and kill-switch features
3. Cross-platform support (macOS, Linux)
4. PAC file support for advanced routing
5. Config signing and verification
6. Certificate pinning
7. Better UI with system tray
8. Auto-update mechanism for client binary
