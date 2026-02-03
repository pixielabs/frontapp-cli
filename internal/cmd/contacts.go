package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/dedene/frontapp-cli/internal/api"
	"github.com/dedene/frontapp-cli/internal/errfmt"
	"github.com/dedene/frontapp-cli/internal/output"
)

type ContactCmd struct {
	List    ContactListCmd    `cmd:"" help:"List contacts"`
	Search  ContactSearchCmd  `cmd:"" help:"Search contacts"`
	Get     ContactGetCmd     `cmd:"" help:"Get a contact"`
	Handles ContactHandlesCmd `cmd:"" help:"List contact handles"`
	Handle  ContactHandleCmd  `cmd:"" help:"Manage contact handles"`
	Notes   ContactNotesCmd   `cmd:"" help:"List contact notes"`
	Note    ContactNoteCmd    `cmd:"" help:"Manage contact notes"`
	Convos  ContactConvosCmd  `cmd:"" help:"List conversations for a contact"`
	Create  ContactCreateCmd  `cmd:"" help:"Create a contact"`
	Update  ContactUpdateCmd  `cmd:"" help:"Update a contact"`
	Delete  ContactDeleteCmd  `cmd:"" help:"Delete a contact"`
	Merge   ContactMergeCmd   `cmd:"" help:"Merge contacts"`
}

type ContactListCmd struct {
	Limit int `help:"Maximum results" default:"25"`
}

func (c *ContactListCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	resp, err := client.ListContacts(ctx, c.Limit)
	if err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, resp)
	}

	if len(resp.Results) == 0 {
		fmt.Fprintln(os.Stdout, "No contacts found.")

		return nil
	}

	tbl := output.NewTableWriter(os.Stdout, mode.Plain)
	tbl.AddRow("ID", "NAME", "HANDLE")

	for _, contact := range resp.Results {
		tbl.AddRow(output.FormatContact(contact)...)
	}

	return tbl.Flush()
}

type ContactSearchCmd struct {
	Query string `arg:"" help:"Search query"`
	Limit int    `help:"Maximum results" default:"25"`
}

func (c *ContactSearchCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Set("q", c.Query)
	if c.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", c.Limit))
	}

	var resp api.ListResponse[api.Contact]
	if err := client.Get(ctx, "/contacts/search?"+params.Encode(), &resp); err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, resp)
	}

	if len(resp.Results) == 0 {
		fmt.Fprintln(os.Stdout, "No contacts found.")

		return nil
	}

	tbl := output.NewTableWriter(os.Stdout, mode.Plain)
	tbl.AddRow("ID", "NAME", "HANDLE")

	for _, contact := range resp.Results {
		tbl.AddRow(output.FormatContact(contact)...)
	}

	return tbl.Flush()
}

type ContactGetCmd struct {
	ID string `arg:"" help:"Contact ID"`
}

func (c *ContactGetCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	contact, err := client.GetContact(ctx, c.ID)
	if err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, contact)
	}

	fmt.Fprintf(os.Stdout, "ID:   %s\n", contact.ID)
	fmt.Fprintf(os.Stdout, "Name: %s\n", contact.Name)

	if contact.Description != "" {
		fmt.Fprintf(os.Stdout, "Desc: %s\n", contact.Description)
	}

	if len(contact.Handles) > 0 {
		fmt.Fprintln(os.Stdout, "\nHandles:")

		for _, h := range contact.Handles {
			fmt.Fprintf(os.Stdout, "  %s: %s\n", h.Source, h.Handle)
		}
	}

	return nil
}
