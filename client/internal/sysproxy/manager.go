package sysproxy

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/ZeroDayEvil/VPN/client/internal/logger"
)

// Manager handles Windows system proxy settings
// On Linux/Mac, this would use different methods
type Manager struct {
	dataDir    string
	log        *logger.Logger
	backupFile string
}

type proxyBackup struct {
	ProxyEnable     int    `json:"proxy_enable"`
	ProxyServer     string `json:"proxy_server"`
	ProxyOverride   string `json:"proxy_override"`
	AutoConfigURL   string `json:"auto_config_url"`
}

func NewManager(dataDir string, log *logger.Logger) *Manager {
	return &Manager{
		dataDir:    dataDir,
		log:        log,
		backupFile: filepath.Join(dataDir, "net_backup.json"),
	}
}

func (m *Manager) Enable(proxyAddr, bypass string) error {
	// Backup current settings
	if err := m.backupSettings(); err != nil {
		m.log.Warning("Failed to backup proxy settings: %v", err)
	}

	// On Windows, we would use registry to set proxy
	// For this cross-platform implementation, we'll simulate it
	m.log.Info("Setting system proxy to %s", proxyAddr)
	
	// In a real Windows implementation, we would do:
	// 1. Open HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings
	// 2. Set ProxyEnable = 1
	// 3. Set ProxyServer = proxyAddr
	// 4. Set ProxyOverride = bypass
	// 5. Call WinHttpSetIEProxyConfigForCurrentUser or InternetSetOption
	
	// For MVP cross-platform testing, we'll just log
	m.log.Info("System proxy enabled: %s (bypass: %s)", proxyAddr, bypass)
	
	return nil
}

func (m *Manager) Restore() error {
	// Load backup
	data, err := os.ReadFile(m.backupFile)
	if err != nil {
		// No backup found, disable proxy
		m.log.Info("No backup found, disabling proxy")
		return m.disable()
	}

	var backup proxyBackup
	if err := json.Unmarshal(data, &backup); err != nil {
		m.log.Error("Failed to parse backup: %v", err)
		return m.disable()
	}

	// Restore settings
	m.log.Info("Restoring proxy settings from backup")
	
	// In a real Windows implementation, restore from backup
	// For MVP, just log
	m.log.Info("Proxy settings restored")
	
	// Remove backup file
	os.Remove(m.backupFile)
	
	return nil
}

func (m *Manager) disable() error {
	m.log.Info("Disabling system proxy")
	
	// In a real Windows implementation:
	// 1. Set ProxyEnable = 0
	// 2. Clear ProxyServer
	// 3. Call InternetSetOption
	
	return nil
}

func (m *Manager) backupSettings() error {
	// In a real Windows implementation, read current registry values
	backup := proxyBackup{
		ProxyEnable:   0,
		ProxyServer:   "",
		ProxyOverride: "",
		AutoConfigURL: "",
	}

	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.backupFile, data, 0644)
}

// Windows-specific functions would go here
// These would use syscall or golang.org/x/sys/windows to interact with registry

/*
Example Windows implementation:

import (
	"golang.org/x/sys/windows/registry"
)

func setWindowsProxy(proxyAddr, bypass string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	// Enable proxy
	if err := key.SetDWordValue("ProxyEnable", 1); err != nil {
		return err
	}

	// Set proxy server
	if err := key.SetStringValue("ProxyServer", proxyAddr); err != nil {
		return err
	}

	// Set bypass list
	if err := key.SetStringValue("ProxyOverride", bypass); err != nil {
		return err
	}

	// Notify system of changes
	// This requires calling InternetSetOption with INTERNET_OPTION_SETTINGS_CHANGED
	// or WinHttpSetIEProxyConfigForCurrentUser

	return nil
}
*/
