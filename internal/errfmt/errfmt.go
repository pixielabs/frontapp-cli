package errfmt

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dedene/frontapp-cli/internal/api"
	"github.com/dedene/frontapp-cli/internal/auth"
)

// Format formats an error into a user-friendly message with actionable suggestions.
func Format(err error) string {
	if err == nil {
		return ""
	}

	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		return formatAPIError(apiErr)
	}

	var authErr *api.AuthError
	if errors.As(err, &authErr) {
		return formatAuthError(authErr)
	}

	var rateLimitErr *api.RateLimitError
	if errors.As(err, &rateLimitErr) {
		return formatRateLimitError(rateLimitErr)
	}

	var circuitBreakerErr *api.CircuitBreakerError
	if errors.As(err, &circuitBreakerErr) {
		return formatCircuitBreakerError()
	}

	if errors.Is(err, auth.ErrNotAuthenticated) {
		return formatNotAuthenticatedError()
	}

	return fmt.Sprintf("Error: %v", err)
}

func formatAPIError(err *api.APIError) string {
	var sb strings.Builder

	switch err.StatusCode {
	case 401:
		sb.WriteString("Error: Not authenticated\n\n")
		sb.WriteString("  Run 'frontcli auth login' to authenticate with Front.\n")

	case 403:
		sb.WriteString("Error: Access denied (403)\n\n")
		sb.WriteString("  You don't have permission to perform this action.\n")
		sb.WriteString("  Check your account permissions in Front.\n")

	case 404:
		sb.WriteString("Error: Not found (404)\n\n")

		if err.Details != "" {
			sb.WriteString("  " + err.Details + "\n\n")
		}

		sb.WriteString("  The resource doesn't exist or you don't have access.\n")

	case 429:
		sb.WriteString("Error: Rate limit exceeded (429)\n\n")
		sb.WriteString("  You've hit Front's API rate limit.\n")
		sb.WriteString("  Tip: Use --limit flag to reduce result set size.\n")

	default:
		sb.WriteString(fmt.Sprintf("Error: %s (%d)\n", err.Message, err.StatusCode))

		if err.Details != "" {
			sb.WriteString("\n  " + err.Details + "\n")
		}
	}

	return sb.String()
}

func formatAuthError(err *api.AuthError) string {
	var sb strings.Builder

	sb.WriteString("Error: Authentication failed\n\n")
	sb.WriteString(fmt.Sprintf("  %v\n\n", err.Err))
	sb.WriteString("  Try running 'frontcli auth login' to re-authenticate.\n")

	return sb.String()
}

func formatRateLimitError(err *api.RateLimitError) string {
	var sb strings.Builder

	sb.WriteString("Error: Rate limit exceeded\n\n")

	if err.RetryAfter > 0 {
		sb.WriteString(fmt.Sprintf("  Retry after %d seconds.\n", err.RetryAfter))
	}

	sb.WriteString("  Tip: Use --limit flag to reduce result set size.\n")

	return sb.String()
}

func formatCircuitBreakerError() string {
	var sb strings.Builder

	sb.WriteString("Error: Service temporarily unavailable\n\n")
	sb.WriteString("  Too many consecutive failures. Please wait and try again.\n")

	return sb.String()
}

func formatNotAuthenticatedError() string {
	var sb strings.Builder

	sb.WriteString("Error: Not authenticated\n\n")
	sb.WriteString("  Run 'frontcli auth login' to authenticate with Front.\n\n")
	sb.WriteString("  If you need to set up OAuth credentials first:\n")
	sb.WriteString("    frontcli auth setup <client_id> <client_secret>\n")

	return sb.String()
}
