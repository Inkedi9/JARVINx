package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var reRestarts = regexp.MustCompile(`(\d+)\s+restarts?`)

// ContainerState représente l'état d'un container Docker
type ContainerState struct {
	ID           string
	Name         string
	Image        string
	Status       string    // Docker State: "running", "exited", "paused", etc.
	Running      bool
	Exited       bool
	Unhealthy    bool      // parsed from raw Docker Status string
	RestartCount int       // parsed from "(N restarts)" in raw Status string
	CreatedAt    time.Time // parsed from Created unix timestamp
}

// DockerAvailable vérifie si Docker est accessible
func DockerAvailable() bool {
	client := dockerHTTPClient()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://docker/version", nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode == http.StatusOK
}

// ListContainers retourne tous les containers (running + stopped)
func ListContainers(ctx context.Context) ([]ContainerState, error) {
	client := dockerHTTPClient()

	req, err := http.NewRequestWithContext(ctx, "GET",
		"http://docker/containers/json?all=true", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("docker request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("docker status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Structure brute de l'API Docker
	var raw []struct {
		ID      string   `json:"Id"`
		Names   []string `json:"Names"`
		Image   string   `json:"Image"`
		State   string   `json:"State"`
		Status  string   `json:"Status"`  // human-readable: "Up 2h (unhealthy)", "(3 restarts)"
		Created int64    `json:"Created"` // unix timestamp
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	containers := make([]ContainerState, 0, len(raw))
	for _, c := range raw {
		name := c.ID[:12] // fallback
		if len(c.Names) > 0 {
			name = c.Names[0]
			// Docker préfixe les noms avec "/"
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
		}

		restartCount := 0
		if m := reRestarts.FindStringSubmatch(c.Status); len(m) > 1 {
			if n, err := strconv.Atoi(m[1]); err == nil {
				restartCount = n
			}
		}

		containers = append(containers, ContainerState{
			ID:           c.ID[:12],
			Name:         name,
			Image:        c.Image,
			Status:       c.State,
			Running:      c.State == "running",
			Exited:       c.State == "exited",
			Unhealthy:    strings.Contains(c.Status, "unhealthy"),
			RestartCount: restartCount,
			CreatedAt:    time.Unix(c.Created, 0),
		})
	}

	return containers, nil
}

// dockerHTTPClient crée un client HTTP qui parle au socket Docker
func dockerHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				if runtime.GOOS == "windows" {
					// Named pipe Windows
					return dialWindowsDockerPipe(ctx)
				}
				// Unix socket Linux/macOS
				return (&net.Dialer{}).DialContext(ctx, "unix", "/var/run/docker.sock")
			},
		},
	}
}

// dialWindowsDockerPipe se connecte au named pipe Docker sur Windows
func dialWindowsDockerPipe(ctx context.Context) (net.Conn, error) {
	// Docker Desktop sur Windows expose aussi un socket TCP
	d := &net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", "localhost:2375")
	if err != nil {
		// Fallback — essaie le socket npipe via TCP proxy de Docker Desktop
		return nil, fmt.Errorf("docker not accessible on Windows (enable TCP in Docker Desktop settings): %w", err)
	}
	return conn, nil
}
