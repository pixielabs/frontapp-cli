package markdown

import (
	"fmt"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

func ToMarkdown(input string) (string, error) {
	if strings.TrimSpace(input) == "" {
		return "", nil
	}

	md, err := htmltomarkdown.ConvertString(input)
	if err != nil {
		return "", fmt.Errorf("convert html: %w", err)
	}

	return md, nil
}
