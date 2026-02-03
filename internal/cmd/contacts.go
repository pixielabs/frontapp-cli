package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

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
	Query    string `arg:"" help:"Search query (matches name or handle)"`
	Limit    int    `help:"Maximum results" default:"25"`
	MaxPages int    `help:"Maximum pages to search (100 contacts/page)" default:"25"`
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

	// Front API doesn't have a contact search endpoint, so we paginate
	// through contacts and filter client-side until we have enough matches.
	query := strings.ToLower(c.Query)
	var matches []api.Contact

	// First page: fetch max 100 contacts
	resp, err := client.ListContacts(ctx, 100)
	if err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))
		return err
	}

	for page := 1; page <= c.MaxPages; page++ {
		for _, contact := range resp.Results {
			if contactMatches(contact, query) {
				matches = append(matches, contact)
				if len(matches) >= c.Limit {
					break
				}
			}
		}

		// Stop if we have enough matches or no more pages
		if len(matches) >= c.Limit || resp.Pagination.Next == "" {
			break
		}

		// Fetch next page
		resp, err = client.ListContactsPage(ctx, resp.Pagination.Next)
		if err != nil {
			fmt.Fprint(os.Stderr, errfmt.Format(err))
			return err
		}
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, matches)
	}

	if len(matches) == 0 {
		fmt.Fprintln(os.Stdout, "No contacts found.")
		return nil
	}

	tbl := output.NewTableWriter(os.Stdout, mode.Plain)
	tbl.AddRow("ID", "NAME", "HANDLE")

	for _, contact := range matches {
		tbl.AddRow(output.FormatContact(contact)...)
	}

	return tbl.Flush()
}

func contactMatches(contact api.Contact, query string) bool {
	if strings.Contains(strings.ToLower(contact.Name), query) {
		return true
	}

	for _, h := range contact.Handles {
		if strings.Contains(strings.ToLower(h.Handle), query) {
			return true
		}
	}

	return false
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
