package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ZeroDayEvil/VPN/client/internal/config"
	"github.com/ZeroDayEvil/VPN/client/internal/logger"
	"github.com/ZeroDayEvil/VPN/client/internal/proxy"
	"github.com/ZeroDayEvil/VPN/client/internal/api"
	"github.com/ZeroDayEvil/VPN/client/internal/sysproxy"
	"github.com/ZeroDayEvil/VPN/client/internal/ui"
)

const (
	AppVersion = "1.0.0"
	CompanyName = "VPNCompany"
	AppName = "OneClickClient"
)

func main() {
	// Parse command line flags
	noUI := flag.Bool("no-ui", false, "Run without UI (console mode)")
	flag.Parse()

	// Initialize data directory
	dataDir := getDataDir()
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create data directory: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logDir := filepath.Join(dataDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
		os.Exit(1)
	}
	
	log, err := logger.NewLogger(filepath.Join(logDir, "client.log"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Info("Starting One-Click Client v%s", AppVersion)

	// Load or create config
	cfg, err := config.LoadConfig(dataDir)
	if err != nil {
		log.Error("Failed to load config: %v", err)
		os.Exit(1)
	}

	// Initialize device manager
	deviceMgr, err := api.NewDeviceManager(dataDir, cfg)
	if err != nil {
		log.Error("Failed to initialize device manager: %v", err)
		os.Exit(1)
	}

	// Initialize API client
	apiClient := api.NewClient(cfg, deviceMgr, log)

	// Enroll device if needed
	if !deviceMgr.IsEnrolled() {
		log.Info("Device not enrolled, performing enrollment...")
		if err := apiClient.Enroll(); err != nil {
			log.Error("Enrollment failed: %v", err)
			// Continue anyway, will retry on next run
		} else {
			log.Info("Enrollment successful")
		}
	}

	// Initialize proxy manager
	proxyMgr := proxy.NewManager(cfg, log)

	// Initialize system proxy manager
	sysProxyMgr := sysproxy.NewManager(dataDir, log)

	// Create application context
	app := &Application{
		config:      cfg,
		log:         log,
		deviceMgr:   deviceMgr,
		apiClient:   apiClient,
		proxyMgr:    proxyMgr,
		sysProxyMgr: sysProxyMgr,
		dataDir:     dataDir,
	}

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info("Shutdown signal received")
		app.Shutdown()
		os.Exit(0)
	}()

	// Start background services
	go app.startHeartbeat()
	go app.startCommandPolling()

	// Auto-connect on startup
	if err := app.Connect(); err != nil {
		log.Error("Auto-connect failed: %v", err)
	}

	// Start UI
	if *noUI {
		log.Info("Running in console mode")
		select {} // Block forever
	} else {
		if err := ui.Run(app); err != nil {
			log.Error("UI error: %v", err)
			os.Exit(1)
		}
	}
}

type Application struct {
	config      *config.Config
	log         *logger.Logger
	deviceMgr   *api.DeviceManager
	apiClient   *api.Client
	proxyMgr    *proxy.Manager
	sysProxyMgr *sysproxy.Manager
	dataDir     string
	connected   bool
}

func (a *Application) Connect() error {
	a.log.Info("Connecting...")

	// Get proxy profile from API
	profile, err := a.apiClient.GetProxyProfile()
	if err != nil {
		a.log.Error("Failed to get proxy profile: %v", err)
		return fmt.Errorf("failed to get proxy profile: %w", err)
	}

	// Start local proxy
	if err := a.proxyMgr.Start(profile); err != nil {
		a.log.Error("Failed to start proxy: %v", err)
		return fmt.Errorf("failed to start proxy: %w", err)
	}

	// Enable system proxy
	if a.config.Client.EnableSystemProxy {
		proxyAddr := fmt.Sprintf("127.0.0.1:%d", a.config.Client.LocalHTTPPort)
		if err := a.sysProxyMgr.Enable(proxyAddr, a.config.Client.ProxyBypass); err != nil {
			a.log.Error("Failed to enable system proxy: %v", err)
			a.proxyMgr.Stop()
			return fmt.Errorf("failed to enable system proxy: %w", err)
		}
	}

	a.connected = true
	a.log.Info("Connected successfully")
	return nil
}

func (a *Application) Disconnect() error {
	a.log.Info("Disconnecting...")

	// Stop proxy
	if err := a.proxyMgr.Stop(); err != nil {
		a.log.Error("Failed to stop proxy: %v", err)
	}

	// Restore system proxy settings
	if err := a.sysProxyMgr.Restore(); err != nil {
		a.log.Error("Failed to restore system proxy: %v", err)
	}

	a.connected = false
	a.log.Info("Disconnected")
	return nil
}

func (a *Application) Shutdown() {
	a.log.Info("Shutting down...")
	a.Disconnect()
	a.log.Close()
}

func (a *Application) IsConnected() bool {
	return a.connected
}

func (a *Application) GetStatus() string {
	if a.connected {
		return "Connected"
	}
	return "Disconnected"
}

func (a *Application) GetProxySettings() string {
	return fmt.Sprintf("HTTP Proxy: 127.0.0.1:%d\nSOCKS5 Proxy: 127.0.0.1:%d",
		a.config.Client.LocalHTTPPort,
		a.config.Client.LocalSOCKSPort)
}

func (a *Application) startHeartbeat() {
	ticker := time.NewTicker(time.Duration(a.config.Client.HeartbeatIntervalSec) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := a.apiClient.SendHeartbeat(a.connected, a.proxyMgr.GetCurrentNode()); err != nil {
			a.log.Error("Heartbeat failed: %v", err)
		}
	}
}

func (a *Application) startCommandPolling() {
	ticker := time.NewTicker(time.Duration(a.config.Client.CommandsIntervalSec) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		commands, err := a.apiClient.GetCommands()
		if err != nil {
			a.log.Error("Failed to get commands: %v", err)
			continue
		}

		for _, cmd := range commands {
			a.log.Info("Executing command: %s", cmd.Type)
			if err := a.executeCommand(cmd); err != nil {
				a.log.Error("Command execution failed: %v", err)
				a.apiClient.SendCommandResult(cmd.ID, false, err.Error())
			} else {
				a.apiClient.SendCommandResult(cmd.ID, true, "")
			}
		}
	}
}

func (a *Application) executeCommand(cmd *api.Command) error {
	switch cmd.Type {
	case "CONNECT":
		return a.Connect()
	case "DISCONNECT":
		return a.Disconnect()
	case "FORCE_CONFIG_REFRESH":
		return a.config.Refresh()
	default:
		return fmt.Errorf("unknown command: %s", cmd.Type)
	}
}

func getDataDir() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		// Fallback for non-Windows or if APPDATA not set
		home, _ := os.UserHomeDir()
		appData = filepath.Join(home, "AppData", "Roaming")
	}
	return filepath.Join(appData, CompanyName, AppName)
}
