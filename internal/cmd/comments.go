package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/dedene/frontapp-cli/internal/api"
	"github.com/dedene/frontapp-cli/internal/errfmt"
	"github.com/dedene/frontapp-cli/internal/output"
)

type CommentCmd struct {
	List   CommentListCmd   `cmd:"" help:"List comments in a conversation"`
	Get    CommentGetCmd    `cmd:"" help:"Get a comment"`
	Create CommentCreateCmd `cmd:"" help:"Create a comment"`
}

type CommentListCmd struct {
	ConvID string `arg:"" help:"Conversation ID"`
}

func (c *CommentListCmd) Run(flags *RootFlags) error {
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
	if err := client.Get(ctx, fmt.Sprintf("/conversations/%s/comments", c.ConvID), &resp); err != nil {
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

type CommentCreateCmd struct {
	ConvID string `arg:"" help:"Conversation ID"`
	Body   string `required:"" help:"Comment body (@mentions supported)"`
}

func (c *CommentCreateCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	req := map[string]string{
		"body": c.Body,
	}

	var result api.Comment
	if err := client.Post(ctx, fmt.Sprintf("/conversations/%s/comments", c.ConvID), req, &result); err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, result)
	}

	fmt.Fprintf(os.Stdout, "Comment created: %s\n", result.ID)

	return nil
}

type CommentGetCmd struct {
	ID string `arg:"" help:"Comment ID"`
}

func (c *CommentGetCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	var comment api.Comment
	if err := client.Get(ctx, "/comments/"+c.ID, &comment); err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, comment)
	}

	author := "-"
	if comment.Author != nil {
		author = comment.Author.Email
		if author == "" {
			author = comment.Author.Username
		}
	}

	fmt.Fprintf(os.Stdout, "ID:     %s\n", comment.ID)
	fmt.Fprintf(os.Stdout, "Author: %s\n", author)
	fmt.Fprintf(os.Stdout, "Date:   %s\n", output.FormatTimestamp(comment.PostedAt))
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, comment.Body)

	return nil
}
