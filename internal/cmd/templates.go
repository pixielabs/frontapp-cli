package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/dedene/frontapp-cli/internal/api"
	"github.com/dedene/frontapp-cli/internal/errfmt"
	"github.com/dedene/frontapp-cli/internal/output"
)

type TemplateCmd struct {
	List TemplateListCmd `cmd:"" help:"List templates"`
	Get  TemplateGetCmd  `cmd:"" help:"Get a template"`
	Use  TemplateUseCmd  `cmd:"" help:"Output a template body for piping"`
}

type TemplateListCmd struct{}

func (c *TemplateListCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	var resp api.ListResponse[api.Template]
	if err := client.Get(ctx, "/message_templates", &resp); err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, resp)
	}

	if len(resp.Results) == 0 {
		fmt.Fprintln(os.Stdout, "No templates found.")

		return nil
	}

	tbl := output.NewTableWriter(os.Stdout, mode.Plain)
	tbl.AddRow("ID", "NAME", "SUBJECT")

	for _, tmpl := range resp.Results {
		tbl.AddRow(
			tmpl.ID,
			tmpl.Name,
			tmpl.Subject,
		)
	}

	return tbl.Flush()
}

type TemplateGetCmd struct {
	ID string `arg:"" help:"Template ID"`
}

func (c *TemplateGetCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	var tmpl api.Template
	if err := client.Get(ctx, "/message_templates/"+c.ID, &tmpl); err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, tmpl)
	}

	fmt.Fprintf(os.Stdout, "ID:      %s\n", tmpl.ID)
	fmt.Fprintf(os.Stdout, "Name:    %s\n", tmpl.Name)

	if tmpl.Subject != "" {
		fmt.Fprintf(os.Stdout, "Subject: %s\n", tmpl.Subject)
	}

	fmt.Fprintln(os.Stdout, "\nBody:")
	fmt.Fprintln(os.Stdout, tmpl.Body)

	return nil
}

type TemplateUseCmd struct {
	ID string `arg:"" help:"Template ID"`
}

func (c *TemplateUseCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	var tmpl api.Template
	if err := client.Get(ctx, "/message_templates/"+c.ID, &tmpl); err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	fmt.Fprintln(os.Stdout, tmpl.Body)

	return nil
}
