package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/dedene/frontapp-cli/internal/errfmt"
	"github.com/dedene/frontapp-cli/internal/output"
)

type WhoamiCmd struct{}

func (c *WhoamiCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	me, err := client.Me(ctx)
	if err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, me)
	}

	fmt.Fprintf(os.Stdout, "ID:        %s\n", me.ID)
	fmt.Fprintf(os.Stdout, "Email:     %s\n", me.Email)
	fmt.Fprintf(os.Stdout, "Username:  %s\n", me.Username)
	fmt.Fprintf(os.Stdout, "Name:      %s %s\n", me.FirstName, me.LastName)
	fmt.Fprintf(os.Stdout, "Admin:     %v\n", me.IsAdmin)
	fmt.Fprintf(os.Stdout, "Available: %v\n", me.IsAvailable)

	return nil
}
