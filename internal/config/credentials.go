package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const DefaultClientName = "default"

var (
	errInvalidClientName   = errors.New("invalid client name")
	errCredentialsNotFound = errors.New("credentials not found: run 'frontcli auth setup' first")
)

// OAuthCredentials holds OAuth client credentials for the Front API.
type OAuthCredentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri,omitempty"`
}

// NormalizeClientName validates and normalizes a client name.
func NormalizeClientName(raw string) (string, error) {
	name := strings.ToLower(strings.TrimSpace(raw))
	if name == "" {
		return "", fmt.Errorf("%w: empty", errInvalidClientName)
	}

	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			continue
		}

		return "", fmt.Errorf("%w: %q", errInvalidClientName, raw)
	}

	return name, nil
}

// NormalizeClientNameOrDefault normalizes a client name, defaulting to "default" if empty.
func NormalizeClientNameOrDefault(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return DefaultClientName, nil
	}

	return NormalizeClientName(raw)
}

// ClientCredentialsPath returns the path to credentials for a client.
func ClientCredentialsPath(client string) (string, error) {
	dir, err := ClientsDir()
	if err != nil {
		return "", err
	}

	normalized, err := NormalizeClientNameOrDefault(client)
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, normalized+".json"), nil
}

// ClientCredentialsExists checks if credentials exist for a client.
func ClientCredentialsExists(client string) (bool, error) {
	path, err := ClientCredentialsPath(client)
	if err != nil {
		return false, err
	}

	if _, statErr := os.Stat(path); statErr != nil {
		if os.IsNotExist(statErr) {
			return false, nil
		}

		return false, fmt.Errorf("stat credentials: %w", statErr)
	}

	return true, nil
}

// ReadClientCredentials reads OAuth credentials for a client.
func ReadClientCredentials(client string) (OAuthCredentials, error) {
	path, err := ClientCredentialsPath(client)
	if err != nil {
		return OAuthCredentials{}, err
	}

	b, err := os.ReadFile(path) //nolint:gosec // credentials path
	if err != nil {
		if os.IsNotExist(err) {
			return OAuthCredentials{}, errCredentialsNotFound
		}

		return OAuthCredentials{}, fmt.Errorf("read credentials: %w", err)
	}

	var creds OAuthCredentials
	if err := json.Unmarshal(b, &creds); err != nil {
		return OAuthCredentials{}, fmt.Errorf("parse credentials: %w", err)
	}

	return creds, nil
}

// WriteClientCredentials writes OAuth credentials for a client.
func WriteClientCredentials(client string, creds OAuthCredentials) error {
	_, err := EnsureClientsDir()
	if err != nil {
		return fmt.Errorf("ensure clients dir: %w", err)
	}

	path, err := ClientCredentialsPath(client)
	if err != nil {
		return err
	}

	// Set default redirect URI if not specified
	if creds.RedirectURI == "" {
		creds.RedirectURI = "https://localhost:8484/callback"
	}

	b, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("encode credentials: %w", err)
	}

	b = append(b, '\n')

	tmp := path + ".tmp"

	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return fmt.Errorf("write credentials: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("commit credentials: %w", err)
	}

	return nil
}

// DeleteClientCredentials removes OAuth credentials for a client.
func DeleteClientCredentials(client string) error {
	path, err := ClientCredentialsPath(client)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete credentials: %w", err)
	}

	return nil
}

// ClientInfo holds information about a configured OAuth client.
type ClientInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Default bool   `json:"default"`
}

// ListClients returns all configured OAuth clients.
func ListClients() ([]ClientInfo, error) {
	dir, err := ClientsDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("read clients dir: %w", err)
	}

	out := make([]ClientInfo, 0, len(entries))

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}

		clientName := strings.TrimSuffix(name, ".json")

		normalized, err := NormalizeClientName(clientName)
		if err != nil {
			continue
		}

		out = append(out, ClientInfo{
			Name:    normalized,
			Path:    filepath.Join(dir, name),
			Default: normalized == DefaultClientName,
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })

	return out, nil
}
