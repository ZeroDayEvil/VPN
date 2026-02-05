# One-Click VPN Client - User Guide

## What is this?

One-Click VPN Client is a simple application that allows you to connect to a VPN proxy service without needing administrator privileges on your Windows computer.

**Important**: This is a PROXY mode VPN, not a full system VPN. It works with most applications (like web browsers) but some applications may not be covered.

## System Requirements

- Windows 10 or Windows 11 (64-bit)
- No administrator privileges required
- Active internet connection

## Installation

1. Download `Client.exe`
2. Run `Client.exe` - no installation needed!

That's it! The application will:
- Create a folder for settings in `%APPDATA%\VPNCompany\OneClickClient\`
- Download the latest configuration automatically
- Register your device with the service
- Connect automatically

## How to Use

### First Run

When you first run the application:
1. It will register your device
2. Set up a local proxy on your computer
3. Configure Windows to use the proxy
4. Show "Connected" status

### Normal Use

The application runs in the background. You'll see:
- **Status**: Connected or Disconnected
- **Proxy Settings**: The local proxy address

Your internet traffic will be routed through the VPN proxy automatically.

### Closing the Application

**Important**: When you close the application:
1. It will disconnect from the VPN
2. Your original network settings will be restored
3. Your internet will work normally again

## Proxy Settings

The application sets up two local proxy servers:

- **HTTP Proxy**: `127.0.0.1:18080`
- **SOCKS5 Proxy**: `127.0.0.1:18081`

Most applications will use the system proxy automatically. For applications that need manual configuration, use the HTTP proxy address.

## What Works?

✅ **Works Automatically:**
- Web browsers (Chrome, Firefox, Edge)
- Most Windows applications
- Windows Store apps
- Many command-line tools

❌ **May Need Manual Configuration:**
- Some games
- Some VPN-blocking applications
- Applications with custom network code

## Troubleshooting

### Application won't start

1. Make sure you have internet access
2. Try running as different user
3. Check if ports 18080 and 18081 are available

### Internet not working when connected

1. Disconnect the VPN
2. Check if your original internet works
3. Try connecting again
4. Contact support if issue persists

### Some websites don't work

Some applications don't respect system proxy settings. You may need to:
1. Configure the application manually to use `127.0.0.1:18080`
2. Or use the SOCKS5 proxy at `127.0.0.1:18081`

### How to check if it's working?

1. Visit https://www.whatismyip.com/ before connecting
2. Connect to VPN
3. Visit https://www.whatismyip.com/ again
4. Your IP address should be different

## Files and Folders

The application stores data in:
```
%APPDATA%\VPNCompany\OneClickClient\
├── config.json       (Configuration)
├── device.json       (Your device ID and token)
├── net_backup.json   (Backup of network settings)
└── logs\            (Application logs)
```

## Privacy and Security

- Your device gets a unique ID when you first run the app
- All communication with the VPN service is encrypted (HTTPS)
- Your device token is stored locally
- Logs don't contain sensitive information

## Uninstallation

1. Close the application
2. Delete `Client.exe`
3. (Optional) Delete `%APPDATA%\VPNCompany\OneClickClient\` to remove all data

## Support

If you need help:
1. Check the logs in `%APPDATA%\VPNCompany\OneClickClient\logs\`
2. Make sure you're running the latest version
3. Contact your VPN service provider

## Command Line Options

Advanced users can run the application with:

```cmd
Client.exe -no-ui    # Run without user interface (console mode)
```

## Known Limitations

1. **Not a full VPN**: Only works with applications that respect system proxy
2. **No automatic reconnection**: If internet drops, you may need to restart
3. **No kill-switch**: Traffic may leak if connection drops
4. **Windows only**: Does not work on Mac or Linux

## Tips

- Keep the application running for continuous VPN protection
- Check the status regularly to ensure you're connected
- If you have issues, try disconnecting and connecting again
- The application auto-updates its configuration from the internet
