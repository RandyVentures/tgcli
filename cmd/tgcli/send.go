package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSendCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send messages (text, file, reaction)",
	}

	cmd.AddCommand(newSendTextCmd(flags))
	cmd.AddCommand(newSendFileCmd(flags))
	cmd.AddCommand(newSendReactionCmd(flags))

	return cmd
}

func newSendTextCmd(flags *rootFlags) *cobra.Command {
	var to int64
	var message string
	var replyTo int

	cmd := &cobra.Command{
		Use:   "text",
		Short: "Send text message",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("🚧 Send text to %d: '%s' - coming soon\n", to, message)
			if replyTo != 0 {
				fmt.Printf("   (reply to message %d)\n", replyTo)
			}
			return nil
		},
	}

	cmd.Flags().Int64Var(&to, "to", 0, "recipient chat ID (required)")
	cmd.Flags().StringVar(&message, "message", "", "message text (required)")
	cmd.Flags().IntVar(&replyTo, "reply-to", 0, "reply to message ID (optional)")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("message")

	return cmd
}

func newSendFileCmd(flags *rootFlags) *cobra.Command {
	var to int64
	var file string
	var caption string

	cmd := &cobra.Command{
		Use:   "file",
		Short: "Send file",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("🚧 Send file '%s' to %d - coming soon\n", file, to)
			if caption != "" {
				fmt.Printf("   Caption: %s\n", caption)
			}
			return nil
		},
	}

	cmd.Flags().Int64Var(&to, "to", 0, "recipient chat ID (required)")
	cmd.Flags().StringVar(&file, "file", "", "file path (required)")
	cmd.Flags().StringVar(&caption, "caption", "", "file caption (optional)")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func newSendReactionCmd(flags *rootFlags) *cobra.Command {
	var chatID int64
	var messageID int
	var emoji string

	cmd := &cobra.Command{
		Use:   "reaction",
		Short: "Send reaction to message",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("🚧 React with '%s' to message %d in chat %d - coming soon\n", emoji, messageID, chatID)
			return nil
		},
	}

	cmd.Flags().Int64Var(&chatID, "chat", 0, "chat ID (required)")
	cmd.Flags().IntVar(&messageID, "message-id", 0, "message ID (required)")
	cmd.Flags().StringVar(&emoji, "emoji", "", "emoji reaction (required)")
	_ = cmd.MarkFlagRequired("chat")
	_ = cmd.MarkFlagRequired("message-id")
	_ = cmd.MarkFlagRequired("emoji")

	return cmd
}
