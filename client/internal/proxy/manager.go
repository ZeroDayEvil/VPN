package proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"

	"github.com/ZeroDayEvil/VPN/client/internal/api"
	"github.com/ZeroDayEvil/VPN/client/internal/config"
	"github.com/ZeroDayEvil/VPN/client/internal/logger"
)

type Manager struct {
	config      *config.Config
	log         *logger.Logger
	httpServer  *http.Server
	socksServer net.Listener
	profile     *api.ProxyProfile
	currentNode string
	mu          sync.Mutex
	running     bool
}

func NewManager(cfg *config.Config, log *logger.Logger) *Manager {
	return &Manager{
		config: cfg,
		log:    log,
	}
}

func (m *Manager) Start(profile *api.ProxyProfile) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("proxy already running")
	}

	m.profile = profile
	m.currentNode = profile.NodeEndpoint

	// Start HTTP proxy
	if err := m.startHTTPProxy(); err != nil {
		return fmt.Errorf("failed to start HTTP proxy: %w", err)
	}

	// Start SOCKS5 proxy (optional)
	if m.config.Client.LocalSOCKSPort > 0 {
		if err := m.startSOCKS5Proxy(); err != nil {
			m.log.Warning("Failed to start SOCKS5 proxy: %v", err)
		}
	}

	m.running = true
	m.log.Info("Proxy started successfully")
	return nil
}

func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	// Stop HTTP proxy
	if m.httpServer != nil {
		m.httpServer.Close()
		m.httpServer = nil
	}

	// Stop SOCKS5 proxy
	if m.socksServer != nil {
		m.socksServer.Close()
		m.socksServer = nil
	}

	m.running = false
	m.log.Info("Proxy stopped")
	return nil
}

func (m *Manager) GetCurrentNode() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.currentNode
}

func (m *Manager) startHTTPProxy() error {
	addr := fmt.Sprintf("127.0.0.1:%d", m.config.Client.LocalHTTPPort)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handleHTTPRequest(w, r)
	})

	m.httpServer = &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		m.log.Info("HTTP proxy listening on %s", addr)
		if err := m.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			m.log.Error("HTTP proxy error: %v", err)
		}
	}()

	return nil
}

func (m *Manager) handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	// For CONNECT method (HTTPS)
	if r.Method == http.MethodConnect {
		m.handleConnect(w, r)
		return
	}

	// For regular HTTP requests
	m.handleHTTP(w, r)
}

func (m *Manager) handleConnect(w http.ResponseWriter, r *http.Request) {
	// Get the target host
	targetAddr := r.Host
	if _, _, err := net.SplitHostPort(targetAddr); err != nil {
		targetAddr = net.JoinHostPort(targetAddr, "443")
	}

	// Connect to target
	targetConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		m.log.Error("Failed to connect to %s: %v", targetAddr, err)
		http.Error(w, "Failed to connect", http.StatusServiceUnavailable)
		return
	}
	defer targetConn.Close()

	// Hijack the connection
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
	defer clientConn.Close()

	// Send 200 Connection Established
	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	// Bidirectional copy
	go io.Copy(targetConn, clientConn)
	io.Copy(clientConn, targetConn)
}

func (m *Manager) handleHTTP(w http.ResponseWriter, r *http.Request) {
	// Create a new request to the target
	targetURL := r.URL
	if !targetURL.IsAbs() {
		targetURL.Scheme = "http"
		targetURL.Host = r.Host
	}

	proxyReq, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		m.log.Error("Failed to create proxy request: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// Remove hop-by-hop headers
	proxyReq.Header.Del("Proxy-Connection")
	proxyReq.Header.Del("Proxy-Authenticate")
	proxyReq.Header.Del("Proxy-Authorization")

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		m.log.Error("Proxy request failed: %v", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy status code
	w.WriteHeader(resp.StatusCode)

	// Copy body
	io.Copy(w, resp.Body)
}

func (m *Manager) startSOCKS5Proxy() error {
	addr := fmt.Sprintf("127.0.0.1:%d", m.config.Client.LocalSOCKSPort)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	m.socksServer = listener

	go func() {
		m.log.Info("SOCKS5 proxy listening on %s", addr)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go m.handleSOCKS5(conn)
		}
	}()

	return nil
}

func (m *Manager) handleSOCKS5(conn net.Conn) {
	defer conn.Close()

	// SOCKS5 handshake
	buf := make([]byte, 256)

	// Read version and methods
	n, err := conn.Read(buf)
	if err != nil || n < 2 {
		return
	}

	// Check version
	if buf[0] != 0x05 {
		return
	}

	// Send no authentication required
	conn.Write([]byte{0x05, 0x00})

	// Read request
	n, err = conn.Read(buf)
	if err != nil || n < 7 {
		return
	}

	// Parse address
	var targetAddr string
	switch buf[3] {
	case 0x01: // IPv4
		if n < 10 {
			return
		}
		targetAddr = fmt.Sprintf("%d.%d.%d.%d:%d",
			buf[4], buf[5], buf[6], buf[7],
			int(buf[8])<<8|int(buf[9]))
	case 0x03: // Domain name
		addrLen := int(buf[4])
		if n < 5+addrLen+2 {
			return
		}
		domain := string(buf[5 : 5+addrLen])
		port := int(buf[5+addrLen])<<8 | int(buf[6+addrLen])
		targetAddr = fmt.Sprintf("%s:%d", domain, port)
	default:
		conn.Write([]byte{0x05, 0x08, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	// Connect to target
	targetConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		conn.Write([]byte{0x05, 0x05, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}
	defer targetConn.Close()

	// Send success
	conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})

	// Bidirectional copy
	go io.Copy(targetConn, conn)
	io.Copy(conn, targetConn)
}
