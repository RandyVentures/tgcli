package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newMessagesCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "messages",
		Short: "Message operations (list, search)",
	}

	cmd.AddCommand(newMessagesListCmd(flags))
	cmd.AddCommand(newMessagesSearchCmd(flags))

	return cmd
}

func newMessagesListCmd(flags *rootFlags) *cobra.Command {
	var chatID int64
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List messages in a chat",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("🚧 Messages list (chat: %d, limit: %d) - coming soon\n", chatID, limit)
			return nil
		},
	}

	cmd.Flags().Int64Var(&chatID, "chat", 0, "chat ID (required)")
	cmd.Flags().IntVar(&limit, "limit", 50, "max messages to show")
	_ = cmd.MarkFlagRequired("chat")

	return cmd
}

func newMessagesSearchCmd(flags *rootFlags) *cobra.Command {
	var chatID int64

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search messages (FTS)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			scope := "all chats"
			if chatID != 0 {
				scope = fmt.Sprintf("chat %d", chatID)
			}
			fmt.Printf("🚧 Messages search '%s' in %s - coming soon\n", query, scope)
			return nil
		},
	}

	cmd.Flags().Int64Var(&chatID, "chat", 0, "search within specific chat (optional)")

	return cmd
}
