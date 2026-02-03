package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/dedene/frontapp-cli/internal/errfmt"
	"github.com/dedene/frontapp-cli/internal/output"
)

type ChannelCmd struct {
	List ChannelListCmd `cmd:"" help:"List channels"`
	Get  ChannelGetCmd  `cmd:"" help:"Get a channel"`
}

type ChannelListCmd struct{}

func (c *ChannelListCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	resp, err := client.ListChannels(ctx)
	if err != nil {
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

type ChannelGetCmd struct {
	ID string `arg:"" help:"Channel ID"`
}

func (c *ChannelGetCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	ch, err := client.GetChannel(ctx, c.ID)
	if err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, ch)
	}

	fmt.Fprintf(os.Stdout, "ID:      %s\n", ch.ID)
	fmt.Fprintf(os.Stdout, "Type:    %s\n", ch.Type)
	fmt.Fprintf(os.Stdout, "Name:    %s\n", ch.Name)
	fmt.Fprintf(os.Stdout, "Address: %s\n", ch.Address)
	fmt.Fprintf(os.Stdout, "Private: %v\n", ch.IsPrivate)

	return nil
}
