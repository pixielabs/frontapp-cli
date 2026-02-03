package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/dedene/frontapp-cli/internal/api"
	"github.com/dedene/frontapp-cli/internal/errfmt"
	"github.com/dedene/frontapp-cli/internal/output"
)

type ContactNotesCmd struct {
	ID string `arg:"" help:"Contact ID"`
}

func (c *ContactNotesCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	var resp api.ListResponse[api.ContactNote]
	if err := client.Get(ctx, fmt.Sprintf("/contacts/%s/notes", c.ID), &resp); err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, resp)
	}

	if len(resp.Results) == 0 {
		fmt.Fprintln(os.Stdout, "No notes found.")

		return nil
	}

	tbl := output.NewTableWriter(os.Stdout, mode.Plain)
	tbl.AddRow("ID", "AUTHOR", "NOTE", "DATE")

	for _, note := range resp.Results {
		author := "-"
		if note.Author != nil {
			author = note.Author.Email
			if author == "" {
				author = note.Author.Username
			}
		}

		body := note.Body
		if len(body) > 50 {
			body = body[:47] + "..."
		}

		tbl.AddRow(note.ID, author, body, output.FormatTimestamp(note.CreatedAt))
	}

	return tbl.Flush()
}

type ContactNoteCmd struct {
	Add ContactNoteAddCmd `cmd:"" help:"Add a note to a contact"`
}

type ContactNoteAddCmd struct {
	ContactID string `arg:"" help:"Contact ID"`
	Body      string `required:"" help:"Note body"`
}

func (c *ContactNoteAddCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	req := map[string]string{"body": c.Body}

	var result api.ContactNote
	if err := client.Post(ctx, fmt.Sprintf("/contacts/%s/notes", c.ContactID), req, &result); err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, result)
	}

	fmt.Fprintf(os.Stdout, "Note added: %s\n", result.ID)

	return nil
}
