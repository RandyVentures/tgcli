package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newGroupsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groups",
		Short: "Group operations (list, info, members)",
	}

	cmd.AddCommand(newGroupsListCmd(flags))
	cmd.AddCommand(newGroupsInfoCmd(flags))
	cmd.AddCommand(newGroupsMembersCmd(flags))

	return cmd
}

func newGroupsListCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List groups",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🚧 Groups list - coming soon")
			return nil
		},
	}
}

func newGroupsInfoCmd(flags *rootFlags) *cobra.Command {
	var chatID int64

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show group details",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("🚧 Group info for %d - coming soon\n", chatID)
			return nil
		},
	}

	cmd.Flags().Int64Var(&chatID, "chat", 0, "group chat ID (required)")
	_ = cmd.MarkFlagRequired("chat")

	return cmd
}

func newGroupsMembersCmd(flags *rootFlags) *cobra.Command {
	var chatID int64

	cmd := &cobra.Command{
		Use:   "members",
		Short: "List group members",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("🚧 Group members for %d - coming soon\n", chatID)
			return nil
		},
	}

	cmd.Flags().Int64Var(&chatID, "chat", 0, "group chat ID (required)")
	_ = cmd.MarkFlagRequired("chat")

	return cmd
}
