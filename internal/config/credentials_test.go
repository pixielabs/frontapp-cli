package config

import (
	"os"
	"testing"
)

func TestWriteClientCredentialsDefaultsRedirectURI(t *testing.T) {
	temp := t.TempDir()
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", temp)

	t.Cleanup(func() { _ = os.Setenv("HOME", oldHome) })

	creds := OAuthCredentials{
		ClientID:     "id",
		ClientSecret: "secret",
	}

	if err := WriteClientCredentials("default", creds); err != nil {
		t.Fatalf("write credentials: %v", err)
	}

	loaded, err := ReadClientCredentials("default")
	if err != nil {
		t.Fatalf("read credentials: %v", err)
	}

	if loaded.RedirectURI != "https://localhost:8484/callback" {
		t.Fatalf("unexpected redirect URI: %s", loaded.RedirectURI)
	}
}
