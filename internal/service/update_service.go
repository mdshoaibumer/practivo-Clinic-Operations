package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"clinmitra/internal/config"
)

// UpdateInfo holds information about an available update.
type UpdateInfo struct {
	Available      bool   `json:"available"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	DownloadURL    string `json:"downloadURL"`
	ReleaseNotes   string `json:"releaseNotes"`
	PublishedAt    string `json:"publishedAt"`
}

// GitHubRelease represents the GitHub API response for a release.
type GitHubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Body        string        `json:"body"`
	PublishedAt string        `json:"published_at"`
	Assets      []GitHubAsset `json:"assets"`
}

// GitHubAsset represents a downloadable asset in a release.
type GitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// UpdateService handles checking for and applying updates.
type UpdateService struct {
	cfg        *config.Config
	owner      string
	repo       string
	httpClient *http.Client
}

// NewUpdateService creates a new UpdateService.
func NewUpdateService(cfg *config.Config) *UpdateService {
	return &UpdateService{
		cfg:   cfg,
		owner: "mdshoaibumer",
		repo:  "clinmitra-clinic-operations",
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// CheckForUpdate checks GitHub Releases for a newer version.
func (s *UpdateService) CheckForUpdate() (*UpdateInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", s.owner, s.repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ClinmitraDental-Updater")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		slog.Warn("update check failed - no internet?", "error", err)
		return &UpdateInfo{
			Available:      false,
			CurrentVersion: s.cfg.Version,
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// No releases yet
		return &UpdateInfo{
			Available:      false,
			CurrentVersion: s.cfg.Version,
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return &UpdateInfo{
			Available:      false,
			CurrentVersion: s.cfg.Version,
		}, nil
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(s.cfg.Version, "v")

	info := &UpdateInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		ReleaseNotes:   release.Body,
		PublishedAt:    release.PublishedAt,
	}

	if isNewerVersion(currentVersion, latestVersion) {
		info.Available = true
		// Find the Windows installer asset
		for _, asset := range release.Assets {
			if strings.HasSuffix(asset.Name, ".exe") && strings.Contains(strings.ToLower(asset.Name), "setup") {
				info.DownloadURL = asset.BrowserDownloadURL
				break
			}
			// Fallback: any .exe asset
			if strings.HasSuffix(asset.Name, ".exe") {
				info.DownloadURL = asset.BrowserDownloadURL
			}
		}
	}

	return info, nil
}

// DownloadAndInstallUpdate downloads the installer and launches it.
func (s *UpdateService) DownloadAndInstallUpdate(downloadURL string) error {
	if downloadURL == "" {
		return fmt.Errorf("no download URL provided")
	}

	slog.Info("downloading update", "url", downloadURL)

	// Download to temp directory
	resp, err := s.httpClient.Do(mustNewRequest("GET", downloadURL))
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Save installer to temp file
	tmpDir := os.TempDir()
	installerPath := filepath.Join(tmpDir, "ClinmitraDental-Setup.exe")

	out, err := os.Create(installerPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to save installer: %w", err)
	}
	out.Close()

	slog.Info("update downloaded, launching installer", "path", installerPath)

	// Launch the installer with /SILENT flag (NSIS silent install = shows progress, no prompts)
	cmd := exec.Command(installerPath, "/SILENT")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch installer: %w", err)
	}

	// The app will be closed by the installer (NSIS CloseFirst plugin or manual close)
	return nil
}

// isNewerVersion compares semver strings (e.g., "1.0.0" vs "1.1.0").
func isNewerVersion(current, latest string) bool {
	currentParts := parseVersion(current)
	latestParts := parseVersion(latest)

	for i := 0; i < 3; i++ {
		if latestParts[i] > currentParts[i] {
			return true
		}
		if latestParts[i] < currentParts[i] {
			return false
		}
	}
	return false
}

// parseVersion splits "1.2.3" into [1, 2, 3].
func parseVersion(v string) [3]int {
	var parts [3]int
	fmt.Sscanf(v, "%d.%d.%d", &parts[0], &parts[1], &parts[2])
	return parts
}

func mustNewRequest(method, url string) *http.Request {
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Set("User-Agent", "ClinmitraDental-Updater")
	return req
}
