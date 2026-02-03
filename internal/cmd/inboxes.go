package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/dedene/frontapp-cli/internal/api"
	"github.com/dedene/frontapp-cli/internal/errfmt"
	"github.com/dedene/frontapp-cli/internal/output"
)

type InboxCmd struct {
	List     InboxListCmd     `cmd:"" help:"List inboxes"`
	Get      InboxGetCmd      `cmd:"" help:"Get an inbox"`
	Convos   InboxConvosCmd   `cmd:"" help:"List conversations in an inbox"`
	Channels InboxChannelsCmd `cmd:"" help:"List channels in an inbox"`
}

type InboxListCmd struct{}

func (c *InboxListCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	resp, err := client.ListInboxes(ctx)
	if err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, resp)
	}

	if len(resp.Results) == 0 {
		fmt.Fprintln(os.Stdout, "No inboxes found.")

		return nil
	}

	tbl := output.NewTableWriter(os.Stdout, mode.Plain)
	tbl.AddRow("ID", "NAME")

	for _, inbox := range resp.Results {
		tbl.AddRow(output.FormatInbox(inbox)...)
	}

	return tbl.Flush()
}

type InboxGetCmd struct {
	ID string `arg:"" help:"Inbox ID"`
}

func (c *InboxGetCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	inbox, err := client.GetInbox(ctx, c.ID)
	if err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, inbox)
	}

	fmt.Fprintf(os.Stdout, "ID:      %s\n", inbox.ID)
	fmt.Fprintf(os.Stdout, "Name:    %s\n", inbox.Name)
	fmt.Fprintf(os.Stdout, "Private: %v\n", inbox.IsPrivate)

	return nil
}

type InboxConvosCmd struct {
	ID    string `arg:"" help:"Inbox ID"`
	Limit int    `help:"Maximum number of results" default:"25"`
}

func (c *InboxConvosCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/inboxes/%s/conversations?limit=%d", c.ID, c.Limit)
	var resp api.ListResponse[api.Conversation]
	if err := client.Get(ctx, path, &resp); err != nil {
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

type InboxChannelsCmd struct {
	ID string `arg:"" help:"Inbox ID"`
}

func (c *InboxChannelsCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	var resp api.ListResponse[api.Channel]
	if err := client.Get(ctx, fmt.Sprintf("/inboxes/%s/channels", c.ID), &resp); err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, resp)
	}

	if len(resp.Results) == 0 {
		fmt.Fprintln(os.Stdout, "No channels found.")

		return nil
	}

	tbl := output.NewTableWriter(os.Stdout, mode.Plain)
	tbl.AddRow("ID", "TYPE", "NAME", "ADDRESS")

	for _, ch := range resp.Results {
		tbl.AddRow(output.FormatChannel(ch)...)
	}

	return tbl.Flush()
}
