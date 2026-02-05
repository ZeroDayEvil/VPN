# VPN One-Click Client

A turnkey VPN proxy client solution for Windows that works **without administrator privileges**.

## 🎯 What is This?

This is a complete, production-ready implementation of a one-click VPN proxy client system consisting of:

1. **Windows Client** - Single-file executable (Client.exe)
2. **Backend API** - FastAPI-based management server
3. **Proxy Infrastructure** - Configurable proxy nodes
4. **GitHub-based Config** - Automatic configuration updates

## ✨ Key Features

- ✅ **No Admin Rights Required** - Works with standard user privileges
- ✅ **One-Click Connect** - Automatic setup and connection
- ✅ **Auto-Updates** - Configuration updates from GitHub
- ✅ **System Proxy Integration** - Works with most applications automatically
- ✅ **HTTP & SOCKS5** - Dual proxy support
- ✅ **Graceful Shutdown** - Restores original network settings
- ✅ **Remote Management** - Control via web API

## 📋 Requirements

### Client
- Windows 10/11 (64-bit)
- No installation or admin rights needed

### Server
- Ubuntu 20.04+ (or any Linux)
- Python 3.8+
- Domain name (recommended)

## 🚀 Quick Start

### For Users

1. Download `Client.exe`
2. Run it - that's it! ✨

The client will:
- Auto-configure itself
- Register your device
- Set up local proxy
- Connect automatically

### For Administrators

See comprehensive guides in `/docs`:

- **[DEPLOYMENT.md](docs/DEPLOYMENT.md)** - Complete deployment guide
- **[TECHNICAL.md](docs/TECHNICAL.md)** - Technical documentation
- **[PROXY_NODE_SETUP.md](docs/PROXY_NODE_SETUP.md)** - Proxy server setup
- **[USER_GUIDE.md](docs/USER_GUIDE.md)** - End-user documentation

## 📁 Project Structure

```
VPN/
├── client/                 # Go client application
│   ├── main.go            # Main entry point
│   └── internal/          # Internal packages
│       ├── config/        # Configuration management
│       ├── logger/        # Logging system
│       ├── proxy/         # Local proxy servers
│       ├── api/           # Backend API client
│       ├── sysproxy/      # Windows proxy settings
│       └── ui/            # User interface
├── backend/               # Python FastAPI backend
│   ├── main.py           # API server
│   └── requirements.txt  # Dependencies
├── config/               # GitHub-hosted configuration
│   ├── manifest.json    # Version tracking
│   └── config.json      # Application config
└── docs/                # Documentation
    ├── DEPLOYMENT.md    # Deployment guide
    ├── TECHNICAL.md     # Technical docs
    ├── PROXY_NODE_SETUP.md  # Proxy setup
    └── USER_GUIDE.md    # User guide
```

## 🔧 Building

### Client (Windows Executable)

```bash
cd client
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o Client.exe
```

### Backend API

```bash
cd backend
pip3 install -r requirements.txt
python3 main.py
```

## 🧪 Testing Locally

### 1. Start Backend

```bash
cd backend
python3 main.py
# API runs on http://localhost:8000
```

### 2. Update Config

Edit `config/config.json`:
```json
{
  "api": {
    "base_url": "http://localhost:8000"
  }
}
```

### 3. Run Client

```bash
cd client
go run main.go -no-ui
```

## 📚 API Endpoints

### Client Endpoints
- `POST /v1/enroll` - Device registration
- `POST /v1/heartbeat` - Status updates
- `GET /v1/commands` - Get pending commands
- `POST /v1/command_result` - Report command execution

### Admin Endpoints
- `GET /admin/devices` - List all devices
- `POST /admin/commands` - Send commands to devices

Full API docs: `http://localhost:8000/docs`

## 🔒 Security

### MVP Security Features
- HTTPS-only communication (in production)
- SHA256 config verification
- Token-based device authentication
- No sensitive data in logs

### Production Recommendations
- Implement config signing
- Use certificate pinning
- Enable token rotation
- Add rate limiting
- Use encrypted token storage

## ⚠️ Important Limitations

This is a **PROXY MODE** VPN, not a full system VPN:

✅ **Works With:**
- Web browsers (Chrome, Firefox, Edge, etc.)
- Most Windows applications
- Command-line tools
- Windows Store apps

❌ **May Not Work With:**
- Some games
- Applications with custom network code
- VPN-blocking software

**Note:** Applications that don't respect system proxy settings will need manual configuration.

## 🎯 Use Cases

- **Corporate Access** - Remote access without VPN client installation
- **Privacy Enhancement** - Basic privacy for web browsing
- **Geo-restriction Bypass** - Access region-locked content
- **Development Testing** - Test applications through proxy
- **No-Admin Environments** - Use on locked-down computers

## 🚦 Deployment Checklist

- [ ] Deploy Backend API on server with HTTPS
- [ ] Set up Proxy Node(s)
- [ ] Configure SSL certificates
- [ ] Update config.json with production URLs
- [ ] Commit configs to GitHub
- [ ] Build Windows client
- [ ] Test end-to-end connection
- [ ] Distribute client to users

## 📖 Documentation

Comprehensive documentation available in `/docs`:

1. **[DEPLOYMENT.md](docs/DEPLOYMENT.md)** - Step-by-step deployment guide
   - Backend API setup
   - Proxy node configuration
   - SSL/HTTPS setup
   - GitHub configuration
   - Testing procedures

2. **[TECHNICAL.md](docs/TECHNICAL.md)** - Technical details
   - Architecture overview
   - Building and running
   - Configuration options
   - Troubleshooting
   - Security considerations

3. **[PROXY_NODE_SETUP.md](docs/PROXY_NODE_SETUP.md)** - Proxy setup
   - 3proxy setup
   - Squid setup
   - Custom proxy implementation
   - Security hardening
   - Monitoring and scaling

4. **[USER_GUIDE.md](docs/USER_GUIDE.md)** - End-user guide
   - Installation instructions
   - How to use
   - Troubleshooting
   - Known limitations

## 🤝 Contributing

This is a production implementation of the MVP specification. For changes:

1. Review the technical specification in the problem statement
2. Ensure compatibility with existing components
3. Test thoroughly on Windows
4. Update documentation

## 📄 License

This project implements a technical specification for a VPN proxy client system. Use according to your needs.

## 🆘 Support

### For Users
Check the [User Guide](docs/USER_GUIDE.md) for common issues and solutions.

### For Administrators
See [Technical Documentation](docs/TECHNICAL.md) for detailed troubleshooting.

### Logs Location
Client logs: `%APPDATA%\VPNCompany\OneClickClient\logs\`

## 🎉 What's Next?

Current version is MVP 1.0. Future enhancements could include:

- Full VPN mode with WireGuard (requires admin)
- Cross-platform support (macOS, Linux)
- Kill-switch functionality
- Split tunneling
- Auto-update for client binary
- Advanced UI with system tray
- PAC file support
- Config signing and verification

## 📝 Notes

- This implements the technical specification dated 2026-02-05
- It's designed for rapid deployment ("today-tomorrow")
- Focus is on simplicity and "no admin rights" requirement
- All components are production-ready for MVP use

---

**Ready to deploy?** Start with the [Deployment Guide](docs/DEPLOYMENT.md)!

**Need technical details?** Check the [Technical Documentation](docs/TECHNICAL.md)!