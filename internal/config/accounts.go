package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	errInvalidDomain = errors.New("invalid domain")
	errMissingEmail  = errors.New("missing email")
	errEmptyAlias    = errors.New("alias cannot be empty")
)

// NormalizeAccountAlias normalizes an account alias to lowercase.
func NormalizeAccountAlias(alias string) string {
	return strings.ToLower(strings.TrimSpace(alias))
}

// DomainFromEmail extracts the domain from an email address.
func DomainFromEmail(email string) string {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return ""
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

// NormalizeDomain validates and normalizes a domain name.
func NormalizeDomain(raw string) (string, error) {
	domain := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(raw)), "@")
	if domain == "" {
		return "", fmt.Errorf("%w: empty", errInvalidDomain)
	}

	if !strings.Contains(domain, ".") {
		return "", fmt.Errorf("%w: %q", errInvalidDomain, raw)
	}

	for _, r := range domain {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			continue
		}

		return "", fmt.Errorf("%w: %q", errInvalidDomain, raw)
	}

	return domain, nil
}

// ResolveAccount resolves the account email from flags, env, or config.
// Resolution order: flag → env (FRONT_ACCOUNT) → config default → error
func ResolveAccount(flagAccount string) (string, error) {
	// 1. Explicit flag
	if flagAccount != "" {
		return resolveAlias(flagAccount)
	}

	// 2. Environment variable
	if envAccount := os.Getenv("FRONT_ACCOUNT"); envAccount != "" {
		return resolveAlias(envAccount)
	}

	// 3. Config default
	cfg, err := ReadConfig()
	if err != nil {
		return "", err
	}

	if cfg.DefaultAccount != "" {
		return resolveAlias(cfg.DefaultAccount)
	}

	return "", nil
}

// resolveAlias expands an alias to its email if configured.
func resolveAlias(alias string) (string, error) {
	normalized := NormalizeAccountAlias(alias)
	if normalized == "" {
		return "", nil
	}

	// Check if it's an alias
	cfg, err := ReadConfig()
	if err != nil {
		return "", err
	}

	if cfg.AccountAliases != nil {
		if email, ok := cfg.AccountAliases[normalized]; ok && email != "" {
			return strings.ToLower(strings.TrimSpace(email)), nil
		}
	}

	// Not an alias, return as-is (might be an email or account ID)
	return normalized, nil
}

// ResolveClientForAccount resolves the OAuth client for an account.
// Resolution order: explicit override → domain mapping → default
func ResolveClientForAccount(email, override string) (string, error) {
	// 1. Explicit override
	if strings.TrimSpace(override) != "" {
		return NormalizeClientNameOrDefault(override)
	}

	cfg, err := ReadConfig()
	if err != nil {
		return "", err
	}

	// 2. Domain-based resolution
	domain := DomainFromEmail(email)
	if domain != "" && cfg.AccountDomains != nil {
		if client, ok := cfg.AccountDomains[domain]; ok && strings.TrimSpace(client) != "" {
			return NormalizeClientNameOrDefault(client)
		}
	}

	// 3. Default
	return DefaultClientName, nil
}

// SetAccountAlias sets an alias for an account email.
func SetAccountAlias(alias, email string) error {
	alias = NormalizeAccountAlias(alias)
	email = strings.ToLower(strings.TrimSpace(email))

	if alias == "" {
		return errEmptyAlias
	}

	if email == "" {
		return errMissingEmail
	}

	cfg, err := ReadConfig()
	if err != nil {
		return err
	}

	if cfg.AccountAliases == nil {
		cfg.AccountAliases = map[string]string{}
	}

	cfg.AccountAliases[alias] = email

	return WriteConfig(cfg)
}

// DeleteAccountAlias removes an alias.
func DeleteAccountAlias(alias string) (bool, error) {
	alias = NormalizeAccountAlias(alias)

	cfg, err := ReadConfig()
	if err != nil {
		return false, err
	}

	if cfg.AccountAliases == nil {
		return false, nil
	}

	if _, ok := cfg.AccountAliases[alias]; !ok {
		return false, nil
	}

	delete(cfg.AccountAliases, alias)

	return true, WriteConfig(cfg)
}

// ListAccountAliases returns all configured aliases.
func ListAccountAliases() (map[string]string, error) {
	cfg, err := ReadConfig()
	if err != nil {
		return nil, err
	}

	if cfg.AccountAliases == nil {
		return map[string]string{}, nil
	}

	out := make(map[string]string, len(cfg.AccountAliases))
	for k, v := range cfg.AccountAliases {
		out[k] = v
	}

	return out, nil
}

// SetDefaultAccount sets the default account.
func SetDefaultAccount(account string) error {
	cfg, err := ReadConfig()
	if err != nil {
		return err
	}

	cfg.DefaultAccount = strings.ToLower(strings.TrimSpace(account))

	return WriteConfig(cfg)
}

// SetAccountDomain maps a domain to an OAuth client.
func SetAccountDomain(domain, client string) error {
	normalizedDomain, err := NormalizeDomain(domain)
	if err != nil {
		return err
	}

	normalizedClient, err := NormalizeClientNameOrDefault(client)
	if err != nil {
		return err
	}

	cfg, err := ReadConfig()
	if err != nil {
		return err
	}

	if cfg.AccountDomains == nil {
		cfg.AccountDomains = map[string]string{}
	}

	cfg.AccountDomains[normalizedDomain] = normalizedClient

	return WriteConfig(cfg)
}
