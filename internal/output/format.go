package output

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type Mode struct {
	JSON  bool
	Plain bool
}

type ctxKey struct{}

func WithMode(ctx context.Context, mode Mode) context.Context {
	return context.WithValue(ctx, ctxKey{}, mode)
}

func FromContext(ctx context.Context) Mode {
	if v := ctx.Value(ctxKey{}); v != nil {
		if m, ok := v.(Mode); ok {
			return m
		}
	}

	return Mode{}
}

func IsJSON(ctx context.Context) bool {
	return FromContext(ctx).JSON
}

func FromEnv() Mode {
	return Mode{
		JSON:  envBool("FRONT_JSON"),
		Plain: envBool("FRONT_PLAIN"),
	}
}

func WriteJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	return nil
}

func envBool(key string) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}
