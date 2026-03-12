package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newChatsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chats",
		Short: "Chat operations (list, info)",
	}

	cmd.AddCommand(newChatsListCmd(flags))
	cmd.AddCommand(newChatsInfoCmd(flags))

	return cmd
}

func newChatsListCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all chats (DMs, groups, channels)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🚧 Chats list - coming soon")
			return nil
		},
	}
}

func newChatsInfoCmd(flags *rootFlags) *cobra.Command {
	var chatID int64

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show chat details",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("🚧 Chat info for %d - coming soon\n", chatID)
			return nil
		},
	}

	cmd.Flags().Int64Var(&chatID, "chat", 0, "chat ID (required)")
	_ = cmd.MarkFlagRequired("chat")

	return cmd
}
