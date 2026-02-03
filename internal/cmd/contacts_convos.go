package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/dedene/frontapp-cli/internal/api"
	"github.com/dedene/frontapp-cli/internal/errfmt"
	"github.com/dedene/frontapp-cli/internal/output"
)

type ContactConvosCmd struct {
	ID    string `arg:"" help:"Contact ID"`
	Limit int    `help:"Maximum results" default:"25"`
}

func (c *ContactConvosCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/contacts/%s/conversations?limit=%d", c.ID, c.Limit)
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
