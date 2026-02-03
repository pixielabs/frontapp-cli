package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dedene/frontapp-cli/internal/auth"
	"github.com/dedene/frontapp-cli/internal/config"
)

type AuthCmd struct {
	Setup  AuthSetupCmd  `cmd:"" help:"Configure OAuth credentials"`
	Login  AuthLoginCmd  `cmd:"" help:"Authenticate with Front"`
	Logout AuthLogoutCmd `cmd:"" help:"Remove stored tokens"`
	Status AuthStatusCmd `cmd:"" help:"Show authentication status"`
	List   AuthListCmd   `cmd:"" help:"List authenticated accounts"`
}

type AuthSetupCmd struct {
	ClientID     string `arg:"" help:"OAuth client ID"`
	ClientSecret string `arg:"" help:"OAuth client secret"`
	ClientName   string `help:"Client name (default: default)" default:"default" name:"client-name"`
	RedirectURI  string `help:"OAuth redirect URI" default:"https://localhost:8484/callback"`
}

func (c *AuthSetupCmd) Run() error {
	creds := config.OAuthCredentials{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURI:  c.RedirectURI,
	}

	if err := config.WriteClientCredentials(c.ClientName, creds); err != nil {
		return fmt.Errorf("save credentials: %w", err)
	}

	path, _ := config.ClientCredentialsPath(c.ClientName)
	fmt.Fprintf(os.Stdout, "Credentials saved to %s\n", path)
	fmt.Fprintln(os.Stdout, "Run 'frontcli auth login' to authenticate.")

	return nil
}

type AuthLoginCmd struct {
	Email        string `help:"Email/identifier to associate with this token" name:"email"`
	ClientName   string `help:"Client name" default:"default" name:"client-name"`
	ForceConsent bool   `help:"Force consent prompt even if already authorized"`
	Manual       bool   `help:"Manual authorization (paste URL instead of callback server)"`
}

func (c *AuthLoginCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	refreshToken, err := auth.Authorize(ctx, auth.AuthorizeOptions{
		Client:       c.ClientName,
		ForceConsent: c.ForceConsent,
		Manual:       c.Manual,
		Timeout:      3 * time.Minute,
	})
	if err != nil {
		return fmt.Errorf("authorization failed: %w", err)
	}

	store, err := auth.OpenDefault()
	if err != nil {
		return fmt.Errorf("open keyring: %w", err)
	}

	// Use email or a placeholder
	email := c.Email
	if email == "" && flags != nil && flags.Account != "" {
		email = flags.Account
	}
	if email == "" {
		email = "user@front.app" // Placeholder; will be updated when we have /me endpoint
	}

	tok := auth.Token{
		Email:        email,
		RefreshToken: refreshToken,
		CreatedAt:    time.Now().UTC(),
	}

	if err := store.SetToken(c.ClientName, email, tok); err != nil {
		return fmt.Errorf("store token: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Successfully authenticated as %s\n", email)

	return nil
}

type AuthLogoutCmd struct {
	Email      string `help:"Email/account to log out" name:"email"`
	ClientName string `help:"Client name" default:"default" name:"client-name"`
	All        bool   `help:"Log out all accounts for this client"`
}

func (c *AuthLogoutCmd) Run() error {
	store, err := auth.OpenDefault()
	if err != nil {
		return fmt.Errorf("open keyring: %w", err)
	}

	if c.All {
		tokens, err := store.ListTokens()
		if err != nil {
			return fmt.Errorf("list tokens: %w", err)
		}

		normalizedClient, err := config.NormalizeClientNameOrDefault(c.ClientName)
		if err != nil {
			return err
		}

		count := 0

		for _, tok := range tokens {
			if tok.Client == normalizedClient {
				if err := store.DeleteToken(tok.Client, tok.Email); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to remove token for %s: %v\n", tok.Email, err)
				} else {
					count++
				}
			}
		}

		fmt.Fprintf(os.Stdout, "Logged out %d account(s)\n", count)

		return nil
	}

	if c.Email == "" {
		// Try to find the only account
		tokens, err := store.ListTokens()
		if err != nil {
			return fmt.Errorf("list tokens: %w", err)
		}

		normalizedClient, err := config.NormalizeClientNameOrDefault(c.ClientName)
		if err != nil {
			return err
		}

		var match auth.Token
		count := 0

		for _, tok := range tokens {
			if tok.Client == normalizedClient {
				match = tok
				count++
			}
		}

		if count == 0 {
			return fmt.Errorf("no authenticated accounts found")
		}

		if count > 1 {
			return fmt.Errorf("multiple accounts found; specify --email or use --all")
		}

		c.Email = match.Email
	}

	if err := store.DeleteToken(c.ClientName, c.Email); err != nil {
		return fmt.Errorf("remove token: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Logged out %s\n", c.Email)

	return nil
}

type AuthStatusCmd struct {
	ClientName string `help:"Client name" default:"default" name:"client-name"`
}

func (c *AuthStatusCmd) Run() error {
	// Check if credentials exist
	exists, err := config.ClientCredentialsExists(c.ClientName)
	if err != nil {
		return err
	}

	if !exists {
		fmt.Fprintln(os.Stdout, "Not configured")
		fmt.Fprintln(os.Stdout, "Run 'frontcli auth setup <client_id> <client_secret>' to configure.")

		return nil
	}

	store, err := auth.OpenDefault()
	if err != nil {
		return fmt.Errorf("open keyring: %w", err)
	}

	tokens, err := store.ListTokens()
	if err != nil {
		return fmt.Errorf("list tokens: %w", err)
	}

	normalizedClient, err := config.NormalizeClientNameOrDefault(c.ClientName)
	if err != nil {
		return err
	}

	count := 0

	for _, tok := range tokens {
		if tok.Client == normalizedClient {
			count++
		}
	}

	if count == 0 {
		fmt.Fprintln(os.Stdout, "OAuth credentials configured but not authenticated.")
		fmt.Fprintln(os.Stdout, "Run 'frontcli auth login' to authenticate.")

		return nil
	}

	fmt.Fprintf(os.Stdout, "Authenticated: %d account(s)\n", count)

	for _, tok := range tokens {
		if tok.Client == normalizedClient {
			fmt.Fprintf(os.Stdout, "  - %s (since %s)\n", tok.Email, tok.CreatedAt.Format("2006-01-02"))
		}
	}

	return nil
}

type AuthListCmd struct{}

func (c *AuthListCmd) Run() error {
	store, err := auth.OpenDefault()
	if err != nil {
		return fmt.Errorf("open keyring: %w", err)
	}

	tokens, err := store.ListTokens()
	if err != nil {
		return fmt.Errorf("list tokens: %w", err)
	}

	if len(tokens) == 0 {
		fmt.Fprintln(os.Stdout, "No authenticated accounts.")

		return nil
	}

	fmt.Fprintln(os.Stdout, "Authenticated accounts:")

	for _, tok := range tokens {
		fmt.Fprintf(os.Stdout, "  %s (client: %s, since %s)\n",
			tok.Email, tok.Client, tok.CreatedAt.Format("2006-01-02"))
	}

	return nil
}
