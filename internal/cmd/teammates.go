package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/dedene/frontapp-cli/internal/api"
	"github.com/dedene/frontapp-cli/internal/errfmt"
	"github.com/dedene/frontapp-cli/internal/output"
)

type TeammateCmd struct {
	List   TeammateListCmd   `cmd:"" help:"List teammates"`
	Get    TeammateGetCmd    `cmd:"" help:"Get a teammate"`
	Convos TeammateConvosCmd `cmd:"" help:"List conversations assigned to a teammate"`
}

type TeammateListCmd struct{}

func (c *TeammateListCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	resp, err := client.ListTeammates(ctx)
	if err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, resp)
	}

	if len(resp.Results) == 0 {
		fmt.Fprintln(os.Stdout, "No teammates found.")

		return nil
	}

	tbl := output.NewTableWriter(os.Stdout, mode.Plain)
	tbl.AddRow("ID", "EMAIL", "NAME")

	for _, tm := range resp.Results {
		tbl.AddRow(output.FormatTeammate(tm)...)
	}

	return tbl.Flush()
}

type TeammateGetCmd struct {
	ID string `arg:"" help:"Teammate ID"`
}

func (c *TeammateGetCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	tm, err := client.GetTeammate(ctx, c.ID)
	if err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, tm)
	}

	fmt.Fprintf(os.Stdout, "ID:        %s\n", tm.ID)
	fmt.Fprintf(os.Stdout, "Email:     %s\n", tm.Email)
	fmt.Fprintf(os.Stdout, "Username:  %s\n", tm.Username)
	fmt.Fprintf(os.Stdout, "Name:      %s %s\n", tm.FirstName, tm.LastName)
	fmt.Fprintf(os.Stdout, "Admin:     %v\n", tm.IsAdmin)
	fmt.Fprintf(os.Stdout, "Available: %v\n", tm.IsAvailable)

	return nil
}

type TeammateConvosCmd struct {
	ID    string `arg:"" help:"Teammate ID"`
	Limit int    `help:"Maximum number of results" default:"25"`
}

func (c *TeammateConvosCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/teammates/%s/conversations?limit=%d", c.ID, c.Limit)
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
