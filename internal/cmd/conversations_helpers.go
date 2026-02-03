package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func buildConvSearchQuery(c *ConvSearchCmd) (string, error) {
	if strings.TrimSpace(c.RawQuery) != "" {
		return strings.TrimSpace(c.RawQuery), nil
	}

	parts := make([]string, 0, 8)

	add := func(prefix, value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}

		parts = append(parts, fmt.Sprintf("%s:%s", prefix, value))
	}

	add("from", c.From)
	add("to", c.To)
	add("recipient", c.Recipient)
	add("inbox", c.Inbox)
	for _, tag := range c.Tag {
		add("tag", tag)
	}

	if c.Status != "" {
		status := strings.ToLower(strings.TrimSpace(c.Status))
		switch status {
		case "open", "archived", "snoozed", "trashed":
			parts = append(parts, "is:"+status)
		default:
			return "", fmt.Errorf("invalid status: %s", c.Status)
		}
	}

	if c.Assignee != "" {
		parts = append(parts, "assignee:"+strings.TrimSpace(c.Assignee))
	}

	if c.Unassigned {
		parts = append(parts, "is:unassigned")
	}

	add("before", c.Before)
	add("after", c.After)

	if strings.TrimSpace(c.Query) != "" {
		parts = append(parts, strings.TrimSpace(c.Query))
	}

	if len(parts) == 0 {
		return "", fmt.Errorf("no query provided")
	}

	return strings.Join(parts, " "), nil
}

func readIDsFromInput(source string) ([]string, error) {
	if strings.TrimSpace(source) == "" {
		return nil, nil
	}

	if source != "-" {
		return nil, fmt.Errorf("unsupported ids-from: %s", source)
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("read ids from stdin: %w", err)
	}

	fields := strings.Fields(string(data))
	return fields, nil
}
