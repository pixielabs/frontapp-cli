package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/dedene/frontapp-cli/internal/api"
	"github.com/dedene/frontapp-cli/internal/errfmt"
	"github.com/dedene/frontapp-cli/internal/output"
)

type TagCmd struct {
	List     TagListCmd     `cmd:"" help:"List tags"`
	Get      TagGetCmd      `cmd:"" help:"Get a tag"`
	Create   TagCreateCmd   `cmd:"" help:"Create a tag"`
	Update   TagUpdateCmd   `cmd:"" help:"Update a tag"`
	Delete   TagDeleteCmd   `cmd:"" help:"Delete a tag"`
	Children TagChildrenCmd `cmd:"" help:"List child tags"`
	Convos   TagConvosCmd   `cmd:"" help:"List conversations with a tag"`
}

type TagListCmd struct {
	Tree bool `help:"Show hierarchical tree view"`
}

func (c *TagListCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	resp, err := client.ListTags(ctx)
	if err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, resp)
	}

	if len(resp.Results) == 0 {
		fmt.Fprintln(os.Stdout, "No tags found.")

		return nil
	}

	if c.Tree {
		return renderTagTree(resp.Results)
	}

	tbl := output.NewTableWriter(os.Stdout, mode.Plain)
	tbl.AddRow("ID", "NAME", "COLOR")

	for _, tag := range resp.Results {
		tbl.AddRow(output.FormatTag(tag)...)
	}

	return tbl.Flush()
}

type TagGetCmd struct {
	ID string `arg:"" help:"Tag ID"`
}

func (c *TagGetCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	tag, err := client.GetTag(ctx, c.ID)
	if err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, tag)
	}

	fmt.Fprintf(os.Stdout, "ID:          %s\n", tag.ID)
	fmt.Fprintf(os.Stdout, "Name:        %s\n", tag.Name)

	if tag.Description != "" {
		fmt.Fprintf(os.Stdout, "Description: %s\n", tag.Description)
	}

	if tag.Highlight != "" {
		fmt.Fprintf(os.Stdout, "Color:       %s\n", tag.Highlight)
	}

	fmt.Fprintf(os.Stdout, "Private:     %v\n", tag.IsPrivate)

	return nil
}

type TagCreateCmd struct {
	Name        string `required:"" help:"Tag name"`
	Description string `help:"Tag description"`
	Color       string `help:"Tag color (highlight)"`
	Parent      string `help:"Parent tag ID"`
}

func (c *TagCreateCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	req := map[string]any{
		"name": c.Name,
	}

	if c.Description != "" {
		req["description"] = c.Description
	}

	if c.Color != "" {
		req["highlight"] = c.Color
	}

	if c.Parent != "" {
		req["parent_tag_id"] = c.Parent
	}

	var result api.Tag
	if err := client.Post(ctx, "/tags", req, &result); err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, result)
	}

	fmt.Fprintf(os.Stdout, "Tag created: %s (%s)\n", result.Name, result.ID)

	return nil
}

type TagUpdateCmd struct {
	ID          string `arg:"" help:"Tag ID"`
	Name        string `help:"New name"`
	Description string `help:"New description"`
	Color       string `help:"New color"`
}

func (c *TagUpdateCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	req := map[string]any{}

	if c.Name != "" {
		req["name"] = c.Name
	}

	if c.Description != "" {
		req["description"] = c.Description
	}

	if c.Color != "" {
		req["highlight"] = c.Color
	}

	if len(req) == 0 {
		return fmt.Errorf("no updates specified")
	}

	var result api.Tag
	if err := client.Patch(ctx, "/tags/"+c.ID, req, &result); err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, result)
	}

	fmt.Fprintf(os.Stdout, "Tag updated: %s\n", result.Name)

	return nil
}

type TagDeleteCmd struct {
	ID string `arg:"" help:"Tag ID"`
}

func (c *TagDeleteCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	if err := client.Delete(ctx, "/tags/"+c.ID); err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	fmt.Fprintln(os.Stdout, "Tag deleted")

	return nil
}

type TagChildrenCmd struct {
	ID string `arg:"" help:"Parent tag ID"`
}

func (c *TagChildrenCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	var resp api.ListResponse[api.Tag]
	if err := client.Get(ctx, fmt.Sprintf("/tags/%s/children", c.ID), &resp); err != nil {
		fmt.Fprint(os.Stderr, errfmt.Format(err))

		return err
	}

	if mode.JSON {
		return output.WriteJSON(os.Stdout, resp)
	}

	if len(resp.Results) == 0 {
		fmt.Fprintln(os.Stdout, "No child tags found.")

		return nil
	}

	tbl := output.NewTableWriter(os.Stdout, mode.Plain)
	tbl.AddRow("ID", "NAME", "COLOR")

	for _, tag := range resp.Results {
		tbl.AddRow(output.FormatTag(tag)...)
	}

	return tbl.Flush()
}

type TagConvosCmd struct {
	ID    string `arg:"" help:"Tag ID"`
	Limit int    `help:"Maximum number of results" default:"25"`
}

func (c *TagConvosCmd) Run(flags *RootFlags) error {
	ctx := context.Background()

	client, err := getClient(flags)
	if err != nil {
		return err
	}

	mode, err := resolveOutputMode(flags)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/tags/%s/conversations?limit=%d", c.ID, c.Limit)
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

func renderTagTree(tags []api.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	byParent := make(map[string][]api.Tag)
	for _, tag := range tags {
		parent := strings.TrimSpace(tag.ParentTagID)
		byParent[parent] = append(byParent[parent], tag)
	}

	for parent := range byParent {
		sort.Slice(byParent[parent], func(i, j int) bool {
			return strings.ToLower(byParent[parent][i].Name) < strings.ToLower(byParent[parent][j].Name)
		})
	}

	var walk func(parent string, depth int)
	walk = func(parent string, depth int) {
		children := byParent[parent]
		for _, tag := range children {
			indent := strings.Repeat("  ", depth)
			fmt.Fprintf(os.Stdout, "%s- %s (%s)\n", indent, tag.Name, tag.ID)
			walk(tag.ID, depth+1)
		}
	}

	walk("", 0)

	return nil
}
