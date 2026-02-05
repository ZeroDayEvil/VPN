package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/ZeroDayEvil/VPN/client/internal/config"
	"github.com/ZeroDayEvil/VPN/client/internal/logger"
)

type DeviceManager struct {
	deviceID    string
	deviceToken string
	dataDir     string
	config      *config.Config
}

type deviceData struct {
	DeviceID    string `json:"device_id"`
	DeviceToken string `json:"device_token"`
}

func NewDeviceManager(dataDir string, cfg *config.Config) (*DeviceManager, error) {
	dm := &DeviceManager{
		dataDir: dataDir,
		config:  cfg,
	}

	deviceFile := filepath.Join(dataDir, "device.json")
	data, err := os.ReadFile(deviceFile)
	if err == nil {
		var dd deviceData
		if err := json.Unmarshal(data, &dd); err == nil {
			dm.deviceID = dd.DeviceID
			dm.deviceToken = dd.DeviceToken
			return dm, nil
		}
	}

	// Generate new device ID
	dm.deviceID = uuid.New().String()
	return dm, nil
}

func (dm *DeviceManager) IsEnrolled() bool {
	return dm.deviceToken != ""
}

func (dm *DeviceManager) GetDeviceID() string {
	return dm.deviceID
}

func (dm *DeviceManager) GetDeviceToken() string {
	return dm.deviceToken
}

func (dm *DeviceManager) SetDeviceToken(token string) error {
	dm.deviceToken = token

	dd := deviceData{
		DeviceID:    dm.deviceID,
		DeviceToken: token,
	}

	data, err := json.MarshalIndent(dd, "", "  ")
	if err != nil {
		return err
	}

	deviceFile := filepath.Join(dm.dataDir, "device.json")
	return os.WriteFile(deviceFile, data, 0644)
}

type Client struct {
	config    *config.Config
	deviceMgr *DeviceManager
	log       *logger.Logger
	client    *http.Client
}

func NewClient(cfg *config.Config, deviceMgr *DeviceManager, log *logger.Logger) *Client {
	return &Client{
		config:    cfg,
		deviceMgr: deviceMgr,
		log:       log,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type EnrollRequest struct {
	DeviceID      string `json:"device_id"`
	ClientVersion string `json:"client_version"`
	OSVersion     string `json:"os_version"`
	Hostname      string `json:"hostname"`
}

type EnrollResponse struct {
	DeviceToken  string        `json:"device_token"`
	ProxyProfile *ProxyProfile `json:"proxy_profile"`
}

type ProxyProfile struct {
	NodeEndpoint string    `json:"node_endpoint"`
	Protocol     string    `json:"protocol"`
	Auth         AuthInfo  `json:"auth"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type AuthInfo struct {
	Type     string `json:"type"`
	Token    string `json:"token,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

func (c *Client) Enroll() error {
	hostname, _ := os.Hostname()
	
	req := EnrollRequest{
		DeviceID:      c.deviceMgr.GetDeviceID(),
		ClientVersion: "1.0.0",
		OSVersion:     "Windows 10",
		Hostname:      hostname,
	}

	var resp EnrollResponse
	if err := c.makeRequest("POST", c.config.API.EnrollPath, req, &resp); err != nil {
		return err
	}

	return c.deviceMgr.SetDeviceToken(resp.DeviceToken)
}

type HeartbeatRequest struct {
	Connected     bool   `json:"connected"`
	CurrentNode   string `json:"current_node"`
	LastError     string `json:"last_error"`
	ClientVersion string `json:"client_version"`
	UptimeSec     int64  `json:"uptime_sec"`
}

func (c *Client) SendHeartbeat(connected bool, currentNode string) error {
	req := HeartbeatRequest{
		Connected:     connected,
		CurrentNode:   currentNode,
		ClientVersion: "1.0.0",
		UptimeSec:     0,
	}

	return c.makeRequest("POST", c.config.API.HeartbeatPath, req, nil)
}

type Command struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

type CommandsResponse struct {
	Commands []Command `json:"commands"`
}

func (c *Client) GetCommands() ([]*Command, error) {
	var resp CommandsResponse
	if err := c.makeRequest("GET", c.config.API.CommandsPath, nil, &resp); err != nil {
		return nil, err
	}

	result := make([]*Command, len(resp.Commands))
	for i := range resp.Commands {
		result[i] = &resp.Commands[i]
	}
	return result, nil
}

type CommandResultRequest struct {
	CommandID string `json:"command_id"`
	Status    string `json:"status"`
	Details   string `json:"details,omitempty"`
}

func (c *Client) SendCommandResult(commandID string, success bool, details string) error {
	status := "ok"
	if !success {
		status = "error"
	}

	req := CommandResultRequest{
		CommandID: commandID,
		Status:    status,
		Details:   details,
	}

	return c.makeRequest("POST", c.config.API.CommandResultPath, req, nil)
}

func (c *Client) GetProxyProfile() (*ProxyProfile, error) {
	// For MVP, return a mock profile
	// In production, this would be returned by enroll or a separate endpoint
	return &ProxyProfile{
		NodeEndpoint: "proxy.example.com:443",
		Protocol:     "http",
		Auth: AuthInfo{
			Type:     "token",
			Token:    "mock-token",
		},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

func (c *Client) makeRequest(method, path string, body interface{}, response interface{}) error {
	url := c.config.API.BaseURL + path

	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.deviceMgr.IsEnrolled() {
		req.Header.Set("Authorization", "Bearer "+c.deviceMgr.GetDeviceToken())
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	if response != nil {
		return json.NewDecoder(resp.Body).Decode(response)
	}

	return nil
}
