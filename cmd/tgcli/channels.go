package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newChannelsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channels",
		Short: "Channel operations (list, info)",
	}

	cmd.AddCommand(newChannelsListCmd(flags))
	cmd.AddCommand(newChannelsInfoCmd(flags))

	return cmd
}

func newChannelsListCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List channels",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🚧 Channels list - coming soon")
			return nil
		},
	}
}

func newChannelsInfoCmd(flags *rootFlags) *cobra.Command {
	var chatID int64

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show channel details",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("🚧 Channel info for %d - coming soon\n", chatID)
			return nil
		},
	}

	cmd.Flags().Int64Var(&chatID, "chat", 0, "channel chat ID (required)")
	_ = cmd.MarkFlagRequired("chat")

	return cmd
}
