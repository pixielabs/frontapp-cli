package cmd

type ConvCmd struct {
	List      ConvListCmd      `cmd:"" help:"List conversations"`
	Get       ConvGetCmd       `cmd:"" help:"Get a conversation"`
	Search    ConvSearchCmd    `cmd:"" help:"Search conversations"`
	Messages  ConvMessagesCmd  `cmd:"" help:"List messages in a conversation"`
	Comments  ConvCommentsCmd  `cmd:"" help:"List comments in a conversation"`
	Archive   ConvArchiveCmd   `cmd:"" help:"Archive conversations"`
	Open      ConvOpenCmd      `cmd:"" help:"Open (unarchive) conversations"`
	Trash     ConvTrashCmd     `cmd:"" help:"Move conversations to trash"`
	Assign    ConvAssignCmd    `cmd:"" help:"Assign a conversation"`
	Unassign  ConvUnassignCmd  `cmd:"" help:"Unassign a conversation"`
	Snooze    ConvSnoozeCmd    `cmd:"" help:"Snooze a conversation"`
	Unsnooze  ConvUnsnoozeCmd  `cmd:"" help:"Unsnooze a conversation"`
	Followers ConvFollowersCmd `cmd:"" help:"List followers of a conversation"`
	Follow    ConvFollowCmd    `cmd:"" help:"Follow a conversation"`
	Unfollow  ConvUnfollowCmd  `cmd:"" help:"Unfollow a conversation"`
	Tag       ConvTagCmd       `cmd:"" help:"Add tag to conversation"`
	Untag     ConvUntagCmd     `cmd:"" help:"Remove tag from conversation"`
	Update    ConvUpdateCmd    `cmd:"" help:"Update conversation custom fields"`
}
