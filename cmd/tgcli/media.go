package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newMediaCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "media",
		Short: "Media operations (download)",
	}

	cmd.AddCommand(newMediaDownloadCmd(flags))

	return cmd
}

func newMediaDownloadCmd(flags *rootFlags) *cobra.Command {
	var chatID int64
	var messageID int

	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download media from message",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("🚧 Download media from message %d in chat %d - coming soon\n", messageID, chatID)
			return nil
		},
	}

	cmd.Flags().Int64Var(&chatID, "chat", 0, "chat ID (required)")
	cmd.Flags().IntVar(&messageID, "message-id", 0, "message ID (required)")
	_ = cmd.MarkFlagRequired("chat")
	_ = cmd.MarkFlagRequired("message-id")

	return cmd
}
