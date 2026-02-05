package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Config struct {
	ConfigVersion    string       `json:"config_version"`
	MinClientVersion string       `json:"min_client_version"`
	API              APIConfig    `json:"api"`
	Client           ClientConfig `json:"client"`
	Update           UpdateConfig `json:"update"`
	
	dataDir string
}

type APIConfig struct {
	BaseURL           string `json:"base_url"`
	EnrollPath        string `json:"enroll_path"`
	HeartbeatPath     string `json:"heartbeat_path"`
	CommandsPath      string `json:"commands_path"`
	CommandResultPath string `json:"command_result_path"`
}

type ClientConfig struct {
	HeartbeatIntervalSec int    `json:"heartbeat_interval_sec"`
	CommandsIntervalSec  int    `json:"commands_interval_sec"`
	LogLevel             string `json:"log_level"`
	LocalHTTPPort        int    `json:"local_http_port"`
	LocalSOCKSPort       int    `json:"local_socks_port"`
	EnableSystemProxy    bool   `json:"enable_system_proxy"`
	UsePAC               bool   `json:"use_pac"`
	ProxyBypass          string `json:"proxy_bypass"`
}

type UpdateConfig struct {
	ManifestURL string `json:"manifest_url"`
}

type Manifest struct {
	Version          string `json:"version"`
	ConfigFile       string `json:"config_file"`
	SHA256           string `json:"sha256"`
	MinClientVersion string `json:"min_client_version"`
}

func LoadConfig(dataDir string) (*Config, error) {
	configPath := filepath.Join(dataDir, "config.json")
	
	// Try to load local config
	cfg, err := loadLocalConfig(configPath)
	if err != nil {
		// If local config doesn't exist, try to download from GitHub
		cfg, err = downloadConfig(dataDir)
		if err != nil {
			return nil, fmt.Errorf("failed to load or download config: %w", err)
		}
	}
	
	cfg.dataDir = dataDir
	
	// Try to update config from GitHub
	if err := cfg.checkForUpdates(); err != nil {
		// Log error but continue with existing config
		fmt.Printf("Warning: failed to check for updates: %v\n", err)
	}
	
	return cfg, nil
}

func loadLocalConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	
	return &cfg, nil
}

func downloadConfig(dataDir string) (*Config, error) {
	// For initial setup, use embedded default config
	defaultConfig := &Config{
		ConfigVersion:    "1.0.0",
		MinClientVersion: "1.0.0",
		API: APIConfig{
			BaseURL:           "https://api.example.com",
			EnrollPath:        "/v1/enroll",
			HeartbeatPath:     "/v1/heartbeat",
			CommandsPath:      "/v1/commands",
			CommandResultPath: "/v1/command_result",
		},
		Client: ClientConfig{
			HeartbeatIntervalSec: 30,
			CommandsIntervalSec:  10,
			LogLevel:             "info",
			LocalHTTPPort:        18080,
			LocalSOCKSPort:       18081,
			EnableSystemProxy:    true,
			UsePAC:               false,
			ProxyBypass:          "localhost;127.0.0.1;<local>",
		},
		Update: UpdateConfig{
			ManifestURL: "https://raw.githubusercontent.com/ZeroDayEvil/VPN/main/config/manifest.json",
		},
	}
	
	// Save default config
	configPath := filepath.Join(dataDir, "config.json")
	data, _ := json.MarshalIndent(defaultConfig, "", "  ")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return nil, err
	}
	
	return defaultConfig, nil
}

func (c *Config) checkForUpdates() error {
	if c.Update.ManifestURL == "" {
		return nil
	}
	
	// Download manifest
	resp, err := http.Get(c.Update.ManifestURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download manifest: status %d", resp.StatusCode)
	}
	
	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return err
	}
	
	// Check if update is needed
	configPath := filepath.Join(c.dataDir, "config.json")
	currentSHA, err := calculateSHA256(configPath)
	if err == nil && currentSHA == manifest.SHA256 {
		// Config is up to date
		return nil
	}
	
	// Download new config
	configURL := c.Update.ManifestURL[:len(c.Update.ManifestURL)-len("manifest.json")] + manifest.ConfigFile
	resp, err = http.Get(configURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download config: status %d", resp.StatusCode)
	}
	
	// Read and verify config
	configData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	
	// Verify SHA256
	h := sha256.New()
	h.Write(configData)
	downloadedSHA := hex.EncodeToString(h.Sum(nil))
	
	if downloadedSHA != manifest.SHA256 {
		return fmt.Errorf("config SHA256 mismatch")
	}
	
	// Save new config
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return err
	}
	
	// Reload config
	newCfg, err := loadLocalConfig(configPath)
	if err != nil {
		return err
	}
	
	// Update current config while preserving dataDir
	oldDataDir := c.dataDir
	*c = *newCfg
	c.dataDir = oldDataDir
	
	return nil
}

func (c *Config) Refresh() error {
	return c.checkForUpdates()
}

func calculateSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	
	return hex.EncodeToString(h.Sum(nil)), nil
}
