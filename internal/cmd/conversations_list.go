package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/dedene/frontapp-cli/internal/api"
	"github.com/dedene/frontapp-cli/internal/errfmt"
	"github.com/dedene/frontapp-cli/internal/output"
)

type ConvListCmd struct {
	Inbox  string `help:"Filter by inbox ID"`
	Tag    string `help:"Filter by tag ID"`
	Status string `help:"Filter by status (open, archived, snoozed, trashed)"`
	Limit  int    `help:"Maximum number of results" default:"25"`
}

func (c *ConvListCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	resp, err := client.ListConversations(ctx, api.ListConversationsOptions{
		InboxID: c.Inbox,
		TagID:   c.Tag,
		Status:  c.Status,
		Limit:   c.Limit,
	})
	if err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, resp)
	}

	if len(resp.Results) == 0 {
		fmt.Fprintln(os.Stdout, "No conversations found.")

		return nil
	}

	tbl := output.NewTableWriter(os.Stdout, mode.Plain)
	tbl.AddRow("ID", "STATUS", "ASSIGNEE", "SUBJECT", "CREATED")

	for _, conv := range resp.Results {
		tbl.AddRow(output.FormatConversation(conv)...)
	}

	return tbl.Flush()
}

type ConvGetCmd struct {
	ID       string `arg:"" help:"Conversation ID"`
	Messages bool   `help:"Include messages" short:"m"`
}

func (c *ConvGetCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	conv, err := client.GetConversation(ctx, c.ID)
	if err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		result := map[string]any{"conversation": conv}

		if c.Messages {
			msgs, err := client.ListConversationMessages(ctx, c.ID, 50)
			if err != nil {
				return err
			}

			result["messages"] = msgs.Results
		}

		return output.WriteJSON(os.Stdout, result)
	}

	fmt.Fprintf(os.Stdout, "ID:       %s\n", conv.ID)
	fmt.Fprintf(os.Stdout, "Subject:  %s\n", conv.Subject)
	fmt.Fprintf(os.Stdout, "Status:   %s\n", conv.Status)

	if conv.Assignee != nil {
		fmt.Fprintf(os.Stdout, "Assignee: %s\n", conv.Assignee.Email)
	}

	if len(conv.Tags) > 0 {
		var tagNames []string
		for _, t := range conv.Tags {
			tagNames = append(tagNames, t.Name)
		}

		fmt.Fprintf(os.Stdout, "Tags:     %s\n", strings.Join(tagNames, ", "))
	}

	fmt.Fprintf(os.Stdout, "Created:  %s\n", output.FormatTimestamp(conv.CreatedAt))

	if c.Messages {
		msgs, err := client.ListConversationMessages(ctx, c.ID, 50)
		if err != nil {
			return err
		}

		fmt.Fprintln(os.Stdout, "\nMessages:")

		tbl := output.NewTableWriter(os.Stdout, mode.Plain)
		tbl.AddRow("ID", "DIR", "FROM", "PREVIEW", "DATE")

		for _, msg := range msgs.Results {
			tbl.AddRow(output.FormatMessage(msg)...)
		}

		return tbl.Flush()
	}

	return nil
}

type ConvSearchCmd struct {
	Query      string   `arg:"" optional:"" help:"Search query"`
	RawQuery   string   `help:"Raw query override" short:"q" name:"query"`
	From       string   `help:"Filter by sender (from:)"`
	To         string   `help:"Filter by recipient (to:)"`
	Recipient  string   `help:"Filter by recipient (recipient:)"`
	Inbox      string   `help:"Filter by inbox (inbox:)"`
	Tag        []string `help:"Filter by tag (tag:)"`
	Status     string   `help:"Filter by status (open, archived, snoozed, trashed)"`
	Assignee   string   `help:"Filter by assignee (assignee: or me)"`
	Unassigned bool     `help:"Filter unassigned conversations"`
	Before     string   `help:"Filter before date/time (before:)"`
	After      string   `help:"Filter after date/time (after:)"`
	Limit      int      `help:"Maximum results" default:"25"`
}

func (c *ConvSearchCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	query, err := buildConvSearchQuery(c)
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Set("q", query)
	if c.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", c.Limit))
	}

	var resp api.ListResponse[api.Conversation]
	if err := client.Get(ctx, "/conversations/search?"+params.Encode(), &resp); err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, resp)
	}

	if len(resp.Results) == 0 {
		fmt.Fprintln(os.Stdout, "No conversations found.")

		return nil
	}

	tbl := output.NewTableWriter(os.Stdout, mode.Plain)
	tbl.AddRow("ID", "STATUS", "ASSIGNEE", "SUBJECT", "CREATED")

	for _, conv := range resp.Results {
		tbl.AddRow(output.FormatConversation(conv)...)
	}

	return tbl.Flush()
}

type ConvMessagesCmd struct {
	ID    string `arg:"" help:"Conversation ID"`
	Limit int    `help:"Maximum number of messages" default:"25"`
}

func (c *ConvMessagesCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	resp, err := client.ListConversationMessages(ctx, c.ID, c.Limit)
	if err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, resp)
	}

	if len(resp.Results) == 0 {
		fmt.Fprintln(os.Stdout, "No messages found.")

		return nil
	}

	tbl := output.NewTableWriter(os.Stdout, mode.Plain)
	tbl.AddRow("ID", "DIR", "FROM", "PREVIEW", "DATE")

	for _, msg := range resp.Results {
		tbl.AddRow(output.FormatMessage(msg)...)
	}

	return tbl.Flush()
}

type ConvCommentsCmd struct {
	ID    string `arg:"" help:"Conversation ID"`
	Limit int    `help:"Maximum number of comments" default:"25"`
}

func (c *ConvCommentsCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	var resp api.ListResponse[api.Comment]
	if err := client.Get(ctx, fmt.Sprintf("/conversations/%s/comments?limit=%d", c.ID, c.Limit), &resp); err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, resp)
	}

	if len(resp.Results) == 0 {
		fmt.Fprintln(os.Stdout, "No comments found.")

		return nil
	}

	tbl := output.NewTableWriter(os.Stdout, mode.Plain)
	tbl.AddRow("ID", "AUTHOR", "BODY", "DATE")

	for _, comment := range resp.Results {
		author := "-"
		if comment.Author != nil {
			author = comment.Author.Email
			if author == "" {
				author = comment.Author.Username
			}
		}

		body := comment.Body
		if len(body) > 50 {
			body = body[:47] + "..."
		}

		tbl.AddRow(
			comment.ID,
			author,
			body,
			output.FormatTimestamp(comment.PostedAt),
		)
	}

	return tbl.Flush()
}
